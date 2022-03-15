package blackduck

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/format"
	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/pkg/errors"
)

// Creates a SARIF result from the Vulnerabilities that were brought up by the scan
func CreateSarifResultFile(vulns *Vulnerabilities) *format.SARIF {
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
	tool.Driver.Name = "Blackduck Hub Detect"
	tool.Driver.Version = "unknown"
	tool.Driver.InformationUri = "https://community.synopsys.com/s/document-item?bundleId=integrations-detect&topicId=introduction.html&_LANG=enus"

	// Handle results/vulnerabilities
	for i := 0; i < len(vulns.Items); i++ {
		v := vulns.Items[i]
		result := *new(format.Results)
		id := fmt.Sprintf("%v/%v/%v%v", "SECURITY_VULNERABILITY", v.VulnerabilityName, v.Name, v.Version)
		log.Entry().Debugf("Transforming alert %v into SARIF format", id)
		result.RuleID = id
		result.Level = v.VulnerabilityWithRemediation.Severity
		result.RuleIndex = i //Seems very abstract
		result.Message = format.Message{Text: v.VulnerabilityWithRemediation.Description}
		result.AnalysisTarget = format.ArtifactLocation{URI: v.Name, Index: 0}
		location := format.Location{PhysicalLocation: format.PhysicalLocation{ArtifactLocation: format.ArtifactLocation{URI: v.Name}, Region: format.Region{}, LogicalLocations: []format.LogicalLocation{{FullyQualifiedName: ""}}}}
		result.Locations = append(result.Locations, location)

		sarifRule := *new(format.SarifRule)
		sarifRule.ID = id
		sarifRule.ShortDescription = format.Message{Text: fmt.Sprintf("%v Package %v", v.VulnerabilityName, v.Name)}
		sarifRule.FullDescription = format.Message{Text: v.VulnerabilityWithRemediation.Description}
		sarifRule.DefaultConfiguration.Level = v.Severity
		sarifRule.HelpURI = ""
		sarifRule.Help = format.Help{Text: fmt.Sprintf("Vulnerability %v\nSeverity: %v\nPackage: %v\nInstalled Version: %v\nFix Resolution: %v\nLink: [%v](%v)", v.VulnerabilityName, v.Severity, v.Name, v.Version, v.VulnerabilityWithRemediation.RemediationStatus, "", ""), Markdown: v.ToMarkdown()}

		// Avoid empty descriptions to respect standard
		if sarifRule.ShortDescription.Text == "" {
			sarifRule.ShortDescription.Text = "None."
		}
		if sarifRule.FullDescription.Text == "" { // OR USE OMITEMPTY
			sarifRule.FullDescription.Text = "None."
		}

		ruleProp := *new(format.SarifRuleProperties)
		ruleProp.Tags = append(ruleProp.Tags, "SECURITY_VULNERABILITY")
		ruleProp.Tags = append(ruleProp.Tags, v.VulnerabilityWithRemediation.Description)
		ruleProp.Tags = append(ruleProp.Tags, v.Name)
		ruleProp.Precision = "very-high"
		sarifRule.Properties = &ruleProp

		//Finalize: append the result and the rule
		sarif.Runs[0].Results = append(sarif.Runs[0].Results, result)
		tool.Driver.Rules = append(tool.Driver.Rules, sarifRule)
	}
	//Finalize: tool
	sarif.Runs[0].Tool = tool

	return &sarif
}

func WriteVulnerabilityReports(scanReport reporting.ScanReport, utils piperutils.FileUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := "piper_detect_vulnerability_report.html"
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write html report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "BlackDuck Vulnerability Report", Target: htmlReportPath})

	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("detectExecuteScan_oss_%v.json", fmt.Sprintf("%v", time.Now()))), jsonReport, 0666); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to write json report")
	}

	return reportPaths, nil
}

// WriteSarifFile write a JSON sarif format file for upload into Cumulus
func WriteSarifFile(sarif *format.SARIF, utils piperutils.FileUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	sarifReport, errorMarshall := json.Marshal(sarif)
	if errorMarshall != nil {
		return reportPaths, errors.Wrapf(errorMarshall, "failed to marshall SARIF json file")
	}
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}
	sarifReportPath := filepath.Join(ReportsDirectory, "piper_detect_vulnerability.sarif")
	if err := utils.FileWrite(sarifReportPath, sarifReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write SARIF file")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Blackduck Detect Vulnerability SARIF file", Target: sarifReportPath})

	return reportPaths, nil
}

// CreateGithubResultIssues creates a number of GitHub issues, one per Vulnerability to create transparency on the findings
func CreateGithubResultIssues(vulns *Vulnerabilities, token, APIURL, owner, repository string, assignees, trustedCerts []string) error {
	for i := 0; i < len(vulns.Items); i++ {
		vuln := vulns.Items[i]
		title := fmt.Sprintf("%v/%v/%v%v", "SECURITY_VULNERABILITY", vuln.VulnerabilityName, vuln.Name, vuln.Version)
		markdownReport := vuln.ToMarkdown()
		options := piperGithub.CreateIssueOptions{
			Token:          token,
			APIURL:         APIURL,
			Owner:          owner,
			Repository:     repository,
			Title:          title,
			Body:           []byte(markdownReport),
			Assignees:      assignees,
			UpdateExisting: true,
			TrustedCerts:   trustedCerts,
		}

		log.Entry().Debugf("Creating/updating GitHub issue(s) with title %v in org %v and repo %v", title, owner, repository)
		err := piperGithub.CreateIssue(&options)
		if err != nil {
			return errors.Wrapf(err, "Failed to upload WhiteSource result for %v into GitHub issue", vuln.VulnerabilityName)
		}
	}

	return nil
}
