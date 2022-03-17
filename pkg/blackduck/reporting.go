package blackduck

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/pkg/errors"
)

// CreateSarifResultFile creates a SARIF result from the Vulnerabilities that were brought up by the scan
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
	if vulns != nil && vulns.Items != nil {
		for i := 0; i < len(vulns.Items); i++ {
			v := vulns.Items[i]
			result := *new(format.Results)
			id := v.Title()
			log.Entry().Debugf("Transforming alert %v into SARIF format", id)
			result.RuleID = id
			result.Level = v.VulnerabilityWithRemediation.Severity
			result.RuleIndex = i //Seems very abstract
			result.Message = new(format.Message)
			result.Message.Text = v.VulnerabilityWithRemediation.Description
			result.AnalysisTarget = new(format.ArtifactLocation)
			result.AnalysisTarget.URI = v.Name
			result.AnalysisTarget.Index = 0
			location := format.Location{PhysicalLocation: format.PhysicalLocation{ArtifactLocation: format.ArtifactLocation{URI: v.Name}, Region: format.Region{}, LogicalLocations: []format.LogicalLocation{{FullyQualifiedName: ""}}}}
			result.Locations = append(result.Locations, location)

			sarifRule := *new(format.SarifRule)
			sarifRule.ID = id
			sarifRule.ShortDescription = new(format.Message)
			sarifRule.ShortDescription.Text = fmt.Sprintf("%v Package %v", v.VulnerabilityName, v.Name)
			sarifRule.FullDescription = new(format.Message)
			sarifRule.FullDescription.Text = v.VulnerabilityWithRemediation.Description
			sarifRule.DefaultConfiguration.Level = v.Severity
			sarifRule.HelpURI = ""
			markdown, _ := v.ToMarkdown()
			sarifRule.Help = new(format.Help)
			sarifRule.Help.Text = v.ToTxt()
			sarifRule.Help.Markdown = string(markdown)

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
	}
	//Finalize: tool
	sarif.Runs[0].Tool = tool

	return &sarif
}

// WriteVulnerabilityReports writes vulnerability information from ScanReport into dedicated outputs e.g. HTML
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
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("detectExecuteScan_oss_%v.json", fmt.Sprintf("%v", utils.CurrentTime("")))), jsonReport, 0666); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to write json report")
	}

	return reportPaths, nil
}

// WriteSarifFile write a JSON sarif format file for upload into e.g. GCP
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
