package whitesource

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/format"
	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/pkg/errors"
)

func CreateCustomVulnerabilityReport(productName string, scan *Scan, alerts *[]Alert, cvssSeverityLimit float64) reporting.ScanReport {
	severe, _ := CountSecurityVulnerabilities(alerts, cvssSeverityLimit)

	// sort according to vulnerability severity
	sort.Slice(*alerts, func(i, j int) bool {
		return vulnerabilityScore((*alerts)[i]) > vulnerabilityScore((*alerts)[j])
	})

	projectNames := scan.ScannedProjectNames()

	scanReport := reporting.ScanReport{
		Title: "WhiteSource Security Vulnerability Report",
		Subheaders: []reporting.Subheader{
			{Description: "WhiteSource product name", Details: productName},
			{Description: "Filtered project names", Details: strings.Join(projectNames, ", ")},
		},
		Overview: []reporting.OverviewRow{
			{Description: "Total number of vulnerabilities", Details: fmt.Sprint(len((*alerts)))},
			{Description: "Total number of high/critical vulnerabilities with CVSS score >= 7.0", Details: fmt.Sprint(severe)},
		},
		SuccessfulScan: severe == 0,
		ReportTime:     time.Now(),
	}

	detailTable := reporting.ScanDetailTable{
		NoRowsMessage: "No publicly known vulnerabilities detected",
		Headers: []string{
			"Date",
			"CVE",
			"CVSS Score",
			"CVSS Version",
			"Project",
			"Library file name",
			"Library group ID",
			"Library artifact ID",
			"Library version",
			"Description",
			"Top fix",
		},
		WithCounter:   true,
		CounterHeader: "Entry #",
	}

	for _, alert := range (*alerts) {
		var score float64
		var scoreStyle reporting.ColumnStyle = reporting.Yellow
		if isSevereVulnerability(alert, cvssSeverityLimit) {
			scoreStyle = reporting.Red
		}
		var cveVersion string
		if alert.Vulnerability.CVSS3Score > 0 {
			score = alert.Vulnerability.CVSS3Score
			cveVersion = "v3"
		} else {
			score = alert.Vulnerability.Score
			cveVersion = "v2"
		}

		var topFix string
		emptyFix := Fix{}
		if alert.Vulnerability.TopFix != emptyFix {
			topFix = fmt.Sprintf(`%v<br>%v<br><a href="%v">%v</a>}"`, alert.Vulnerability.TopFix.Message, alert.Vulnerability.TopFix.FixResolution, alert.Vulnerability.TopFix.URL, alert.Vulnerability.TopFix.URL)
		}

		row := reporting.ScanRow{}
		row.AddColumn(alert.Vulnerability.PublishDate, 0)
		row.AddColumn(fmt.Sprintf(`<a href="%v">%v</a>`, alert.Vulnerability.URL, alert.Vulnerability.Name), 0)
		row.AddColumn(score, scoreStyle)
		row.AddColumn(cveVersion, 0)
		row.AddColumn(alert.Project, 0)
		row.AddColumn(alert.Library.Filename, 0)
		row.AddColumn(alert.Library.GroupID, 0)
		row.AddColumn(alert.Library.ArtifactID, 0)
		row.AddColumn(alert.Library.Version, 0)
		row.AddColumn(alert.Vulnerability.Description, 0)
		row.AddColumn(topFix, 0)

		detailTable.Rows = append(detailTable.Rows, row)
	}
	scanReport.DetailTable = detailTable

	return scanReport
}

func CountSecurityVulnerabilities(alerts *[]Alert, cvssSeverityLimit float64) (int, int) {
	severeVulnerabilities := 0
	for _, alert := range *alerts {
		if isSevereVulnerability(alert, cvssSeverityLimit) {
			severeVulnerabilities++
		}
	}

	nonSevereVulnerabilities := len(*alerts) - severeVulnerabilities
	return severeVulnerabilities, nonSevereVulnerabilities
}

func isSevereVulnerability(alert Alert, cvssSeverityLimit float64) bool {

	if vulnerabilityScore(alert) >= cvssSeverityLimit && cvssSeverityLimit >= 0 {
		return true
	}
	return false
}

func vulnerabilityScore(alert Alert) float64 {
	if alert.Vulnerability.CVSS3Score > 0 {
		return alert.Vulnerability.CVSS3Score
	}
	return alert.Vulnerability.Score
}

func ReportSha(productName string, scan *Scan) string {
	reportShaData := []byte(productName + "," + strings.Join(scan.ScannedProjectNames(), ","))
	return fmt.Sprintf("%x", sha1.Sum(reportShaData))
}

func WriteCustomVulnerabilityReports(productName string, scan *Scan, scanReport reporting.ScanReport, utils piperutils.FileUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := filepath.Join(ReportsDirectory, "piper_whitesource_vulnerability_report.html")
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write html report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "WhiteSource Vulnerability Report", Target: htmlReportPath})

	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("whitesourceExecuteScan_oss_%v.json", ReportSha(productName, scan))), jsonReport, 0666); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to write json report")
	}
	// we do not add the json report to the overall list of reports for now,
	// since it is just an intermediary report used as input for later
	// and there does not seem to be real benefit in archiving it.

	return reportPaths, nil
}

func CreateSarifResultFile(scan *Scan, alerts *[]Alert) *format.SARIF {
	//Now, we handle the sarif
	log.Entry().Debug("Creating SARIF file for data transfer")
	var sarif format.SARIF
	sarif.Schema = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos01/schemas/sarif-schema-2.1.0.json"
	sarif.Version = "2.1.0"
	var wsRun format.Runs
	sarif.Runs = append(sarif.Runs, wsRun)

	//handle the tool object
	tool := *new(format.Tool)
	tool.Driver = *new(format.Driver)
	tool.Driver.Name = scan.AgentName
	tool.Driver.Version = scan.AgentVersion
	tool.Driver.InformationUri = "https://whitesource.atlassian.net/wiki/spaces/WD/pages/804814917/Unified+Agent+Overview"

	// Handle results/vulnerabilities
	for i := 0; i < len(*alerts); i++ {
		alert := (*alerts)[i]
		result := *new(format.Results)
		id := fmt.Sprintf("%v/%v/%v", alert.Type, alert.Vulnerability.Name, alert.Library.ArtifactID)
		log.Entry().Debugf("Transforming alert %v into SARIF format", id)
		result.RuleID = id
		result.Level = alert.Level
		result.RuleIndex = i //Seems very abstract
		result.Message = format.Message{Text: alert.Vulnerability.Description}
		result.Level = alert.Level
		result.AnalysisTarget = format.ArtifactLocation{URI: alert.Library.Filename, Index: 0}
		location := format.Location{PhysicalLocation: format.ArtifactLocation{URI: alert.Library.Filename}, Region: format.Region{}, LogicalLocations: []format.LogicalLocation{{FullyQualifiedName: ""}}}
		result.Locations = append(result.Locations, location)

		sarifRule := *new(format.SarifRule)
		sarifRule.Id = id
		sarifRule.ShortDescription = format.Message{Text: fmt.Sprintf("%v Package %v", alert.Vulnerability.Name, alert.Library.ArtifactID)}
		sarifRule.FullDescription = format.Message{Text: alert.Vulnerability.Description}
		sarifRule.DefaultConfiguration.Level = alert.Level
		sarifRule.HelpURI = alert.Vulnerability.URL
		sarifRule.Help = format.Help{Text: fmt.Sprintf("Vulnerability %v\nSeverity: %v\nPackage: %v\nInstalled Version: %v\nFix Resolution: %v\nLink: [%v](%v)", alert.Vulnerability.Name, alert.Vulnerability.Severity, alert.Library.ArtifactID, alert.Library.Version, alert.Vulnerability.TopFix.FixResolution, alert.Vulnerability.Name, alert.Vulnerability.URL), Markdown: alert.ToMarkdown()}

		// Avoid empty descriptions to respect standard
		if sarifRule.ShortDescription.Text == "" {
			sarifRule.ShortDescription.Text = "None."
		}
		if sarifRule.FullDescription.Text == "" { // OR USE OMITEMPTY
			sarifRule.FullDescription.Text = "None."
		}

		var ruleProp *format.SarifRuleProperties
		ruleProp.Tags = append(ruleProp.Tags, alert.Type)
		ruleProp.Tags = append(ruleProp.Tags, alert.Description)
		ruleProp.Tags = append(ruleProp.Tags, alert.Library.ArtifactID)
		ruleProp.Precision = "very-high"
		sarifRule.Properties = ruleProp

		//Finalize: append the result and the rule
		sarif.Runs[0].Results = append(sarif.Runs[0].Results, result)
		tool.Driver.Rules = append(tool.Driver.Rules, sarifRule)
	}
	//Finalize: tool
	sarif.Runs[0].Tool = tool

	return &sarif
}

func WriteSarifFile(sarif *format.SARIF, utils piperutils.FileUtils) ([]piperutils.Path, error) {

	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	sarifReport, errorMarshall := json.Marshal(sarif)
	if errorMarshall != nil {
		return reportPaths, errors.Wrapf(errorMarshall, "failed to marshall SARIF json file")
	}
	sarifReportPath := filepath.Join(ReportsDirectory, "piper_whitesource_vulnerability.sarif")
	if err := utils.FileWrite(sarifReportPath, sarifReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write SARIF file")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "WhiteSource Vulnerability SARIF file", Target: sarifReportPath})

	return reportPaths, nil
}

func CreateGithubResultIssues(scan *Scan, alerts *[]Alert, token, APIURL, owner, repository string, assignees []string) error {
	log.Entry().Debug("Creating/updating GitHub issues with scan results")
	alertArray := *alerts
	for i := 0; i < len(alertArray); i++ {
		alert := alertArray[i]
		markdownReport := alert.ToMarkdown()
		options := piperGithub.CreateIssueOptions{
			Token:          token,
			APIURL:         APIURL,
			Owner:          owner,
			Repository:     repository,
			Title:          fmt.Sprintf("%v/%v/%v", alert.Type, alert.Vulnerability.Name, alert.Library.ArtifactID),
			Body:           []byte(markdownReport),
			Assignees:      assignees,
			UpdateExisting: true,
		}
		err := piperGithub.CreateIssue(&options)
		if err != nil {
			return errors.Wrapf(err, "failed to upload WhiteSource result for %v into GitHub issue", alert.Vulnerability.Name)
		}
	}

	return nil
}
