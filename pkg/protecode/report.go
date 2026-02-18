package protecode

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
)

// ReportData is representing the data of the step report JSON
type ReportData struct {
	Target                      string `json:"target,omitempty"`
	Mandatory                   bool   `json:"mandatory,omitempty"`
	ProductID                   string `json:"productID,omitempty"`
	ServerURL                   string `json:"serverUrl,omitempty"`
	FailOnSevereVulnerabilities bool   `json:"failOnSevereVulnerabilities,omitempty"`
	ExcludeCVEs                 string `json:"excludeCVEs,omitempty"`
	Count                       string `json:"count,omitempty"`
	Cvss2GreaterOrEqualSeven    string `json:"cvss2GreaterOrEqualSeven,omitempty"`
	Cvss3GreaterOrEqualSeven    string `json:"cvss3GreaterOrEqualSeven,omitempty"`
	ExcludedVulnerabilities     string `json:"excludedVulnerabilities,omitempty"`
	TriagedVulnerabilities      string `json:"triagedVulnerabilities,omitempty"`
	HistoricalVulnerabilities   string `json:"historicalVulnerabilities,omitempty"`
	Vulnerabilities             []Vuln `json:"Vulnerabilities,omitempty"`
}

// WriteReport ...
func WriteReport(data ReportData, reportPath string, reportFileName string, result map[string]int, fileUtils piperutils.FileUtils) error {
	data.Mandatory = true
	data.Count = fmt.Sprintf("%v", result["count"])
	data.Cvss2GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss2GreaterOrEqualSeven"])
	data.Cvss3GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss3GreaterOrEqualSeven"])
	data.ExcludedVulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	data.TriagedVulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	data.HistoricalVulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])

	log.Entry().Infof("Protecode scan info, %v of which %v had a CVSS v2 score >= 7.0 and %v had a CVSS v3 score >= 7.0.\n %v vulnerabilities were excluded via configuration (%v) and %v vulnerabilities were triaged via the webUI.\nIn addition %v historical vulnerabilities were spotted. \n\n Vulnerabilities: %v",
		data.Count, data.Cvss2GreaterOrEqualSeven, data.Cvss3GreaterOrEqualSeven,
		data.ExcludedVulnerabilities, data.ExcludeCVEs, data.TriagedVulnerabilities,
		data.HistoricalVulnerabilities, data.Vulnerabilities)
	return writeJSON(reportPath, reportFileName, data, fileUtils)
}

func writeJSON(path, name string, data any, fileUtils piperutils.FileUtils) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return fileUtils.FileWrite(filepath.Join(path, name), jsonData, 0644)
}

func CreateCustomReport(productName string, productID int, data map[string]int, vulns []Vuln) reporting.ScanReport {
	scanReport := reporting.ScanReport{
		ReportTitle: "Protecode Vulnerability Report",
		Subheaders: []reporting.Subheader{
			{Description: "Product name", Details: productName},
			{Description: "Product ID", Details: fmt.Sprint(productID)},
		},
		Overview: []reporting.OverviewRow{
			{Description: "Vulnerabilities", Details: fmt.Sprint(data["vulnerabilities"])},
			{Description: "Major Vulnerabilities", Details: fmt.Sprint(data["major_vulnerabilities"])},
			{Description: "Minor Vulnerabilities", Details: fmt.Sprint(data["minor_vulnerabilities"])},
			{Description: "Triaged Vulnerabilities", Details: fmt.Sprint(data["triaged_vulnerabilities"])},
			{Description: "Excluded Vulnerabilities", Details: fmt.Sprint(data["excluded_vulnerabilities"])},
		},
		ReportTime: time.Now(),
	}

	detailTable := reporting.ScanDetailTable{
		NoRowsMessage: "No findings detected",
		Headers: []string{
			"Issue CVE",
			"CVSS Score",
			"CVSS v3 Score",
		},
		WithCounter:   true,
		CounterHeader: "Entry #",
	}

	for _, vuln := range vulns {
		row := reporting.ScanRow{}
		row.AddColumn(fmt.Sprint(*&vuln.Cve), 0)
		row.AddColumn(fmt.Sprint(*&vuln.Cvss), 0)
		row.AddColumn(fmt.Sprint(*&vuln.Cvss3Score), 0)

		detailTable.Rows = append(detailTable.Rows, row)
	}
	scanReport.DetailTable = detailTable

	return scanReport
}

func WriteCustomReports(scanReport reporting.ScanReport, projectName, projectID string, fileUtils piperutils.FileUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := filepath.Join(ReportsDirectory, "piper_protecode_report.html")
	// Ensure reporting directory exists
	if err := fileUtils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, fmt.Errorf("failed to create report directory: %w", err)
	}
	if err := fileUtils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, fmt.Errorf("failed to write html report: %w", err)
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Protecode Vulnerability Report", Target: htmlReportPath})

	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := fileUtils.DirExists(reporting.StepReportDirectory); !exists {
		err := fileUtils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, fmt.Errorf("failed to create reporting directory: %w", err)
		}
	}
	if err := fileUtils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("protecodeExecuteScan_osvm_%v.json", reportShaProtecode([]string{projectName, projectID}))), jsonReport, 0666); err != nil {
		return reportPaths, fmt.Errorf("failed to write json report: %w", err)
	}
	// we do not add the json report to the overall list of reports for now,
	// since it is just an intermediary report used as input for later
	// and there does not seem to be real benefit in archiving it.

	return reportPaths, nil
}

func reportShaProtecode(parts []string) string {
	reportShaData := []byte(strings.Join(parts, ","))
	return fmt.Sprintf("%x", sha1.Sum(reportShaData))
}
