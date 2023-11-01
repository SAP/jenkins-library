package checkmarxOne

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"

	"github.com/pkg/errors"
)

type CheckmarxOneReportData struct {
	ToolName        string     `json:"toolName"`
	ToolVersion     string     `json:"toolVersion"`
	ProjectName     string     `json:"projectName"`
	ProjectID       string     `json:"projectID"`
	ScanID          string     `json:"scanID"`
	ApplicationName string     `json:"applicationName"`
	ApplicationID   string     `json:"applicationID"`
	GroupName       string     `json:"groupName"`
	GroupID         string     `json:"groupID"`
	DeepLink        string     `json:"deepLink"`
	Preset          string     `json:"preset"`
	ScanType        string     `json:"scanType"`
	Findings        *[]Finding `json:"findings"`
}

type Finding struct {
	ClassificationName string         `json:"classificationName"`
	Total              int            `json:"total,omitempty"`
	Audited            *int           `json:"audited,omitempty"`
	LowPerQuery        *[]LowPerQuery `json:"categories,omitempty"`
}

type LowPerQuery struct {
	QueryName string `json:"name"`
	Audited   int    `json:"audited"`
	Total     int    `json:"total"`
}

func CreateCustomReport(data *map[string]interface{}, insecure, neutral []string) reporting.ScanReport {
	deepLink := fmt.Sprintf(`<a href="%v" target="_blank">Link to scan in CX1 UI</a>`, (*data)["DeepLink"])

	scanReport := reporting.ScanReport{
		ReportTitle: "CheckmarxOne SAST Report",
		Subheaders: []reporting.Subheader{
			{Description: "Project name", Details: fmt.Sprint((*data)["ProjectName"])},
			{Description: "Project ID", Details: fmt.Sprint((*data)["ProjectId"])},
			{Description: "Owner", Details: fmt.Sprint((*data)["Owner"])},
			{Description: "Scan ID", Details: fmt.Sprint((*data)["ScanId"])},
			{Description: "Group", Details: fmt.Sprint((*data)["Group"])},
			{Description: "Group full path", Details: fmt.Sprint((*data)["GroupFullPathOnReportDate"])},
			{Description: "Scan start", Details: fmt.Sprint((*data)["ScanStart"])},
			{Description: "Scan duration", Details: fmt.Sprint((*data)["ScanTime"])},
			{Description: "Scan type", Details: fmt.Sprint((*data)["ScanType"])},
			{Description: "Preset", Details: fmt.Sprint((*data)["Preset"])},
			{Description: "Report creation time", Details: fmt.Sprint((*data)["ReportCreationTime"])},
			{Description: "Lines of code scanned", Details: fmt.Sprint((*data)["LinesOfCodeScanned)"])},
			{Description: "Files scanned", Details: fmt.Sprint((*data)["FilesScanned)"])},
			{Description: "Tool version", Details: fmt.Sprint((*data)["ToolVersion"])},
			{Description: "Deep link", Details: deepLink},
		},
		Overview:   []reporting.OverviewRow{},
		ReportTime: time.Now(),
	}

	for _, issue := range insecure {
		row := reporting.OverviewRow{}
		row.Description = fmt.Sprint(issue)
		row.Style = reporting.Red

		scanReport.Overview = append(scanReport.Overview, row)
	}
	for _, issue := range neutral {
		row := reporting.OverviewRow{}
		row.Description = fmt.Sprint(issue)

		scanReport.Overview = append(scanReport.Overview, row)
	}

	detailTable := reporting.ScanDetailTable{
		Headers: []string{
			"KPI",
			"Count",
		},
		WithCounter: false,
	}
	detailRows := []reporting.OverviewRow{
		{Description: "High issues", Details: fmt.Sprint((*data)["High"].(map[string]int)["Issues"])},
		{Description: "High not false positive issues", Details: fmt.Sprint((*data)["High"].(map[string]int)["NotFalsePositive"])},
		{Description: "High not exploitable issues", Details: fmt.Sprint((*data)["High"].(map[string]int)["NotExploitable"])},
		{Description: "High confirmed issues", Details: fmt.Sprint((*data)["High"].(map[string]int)["Confirmed"])},
		{Description: "High urgent issues", Details: fmt.Sprint((*data)["High"].(map[string]int)["Urgent"])},
		{Description: "High proposed not exploitable issues", Details: fmt.Sprint((*data)["High"].(map[string]int)["ProposedNotExploitable"])},
		{Description: "High to verify issues", Details: fmt.Sprint((*data)["High"].(map[string]int)["ToVerify"])},
		{Description: "Medium issues", Details: fmt.Sprint((*data)["Medium"].(map[string]int)["Issues"])},
		{Description: "Medium not false positive issues", Details: fmt.Sprint((*data)["Medium"].(map[string]int)["NotFalsePositive"])},
		{Description: "Medium not exploitable issues", Details: fmt.Sprint((*data)["Medium"].(map[string]int)["NotExploitable"])},
		{Description: "Medium confirmed issues", Details: fmt.Sprint((*data)["Medium"].(map[string]int)["Confirmed"])},
		{Description: "Medium urgent issues", Details: fmt.Sprint((*data)["Medium"].(map[string]int)["Urgent"])},
		{Description: "Medium proposed not exploitable issues", Details: fmt.Sprint((*data)["Medium"].(map[string]int)["ProposedNotExploitable"])},
		{Description: "Medium to verify issues", Details: fmt.Sprint((*data)["Medium"].(map[string]int)["ToVerify"])},
		{Description: "Low issues", Details: fmt.Sprint((*data)["Low"].(map[string]int)["Issues"])},
		{Description: "Low not false positive issues", Details: fmt.Sprint((*data)["Low"].(map[string]int)["NotFalsePositive"])},
		{Description: "Low not exploitable issues", Details: fmt.Sprint((*data)["Low"].(map[string]int)["NotExploitable"])},
		{Description: "Low confirmed issues", Details: fmt.Sprint((*data)["Low"].(map[string]int)["Confirmed"])},
		{Description: "Low urgent issues", Details: fmt.Sprint((*data)["Low"].(map[string]int)["Urgent"])},
		{Description: "Low proposed not exploitable issues", Details: fmt.Sprint((*data)["Low"].(map[string]int)["ProposedNotExploitable"])},
		{Description: "Low to verify issues", Details: fmt.Sprint((*data)["Low"].(map[string]int)["ToVerify"])},
		{Description: "Informational issues", Details: fmt.Sprint((*data)["Information"].(map[string]int)["Issues"])},
		{Description: "Informational not false positive issues", Details: fmt.Sprint((*data)["Information"].(map[string]int)["NotFalsePositive"])},
		{Description: "Informational not exploitable issues", Details: fmt.Sprint((*data)["Information"].(map[string]int)["NotExploitable"])},
		{Description: "Informational confirmed issues", Details: fmt.Sprint((*data)["Information"].(map[string]int)["Confirmed"])},
		{Description: "Informational urgent issues", Details: fmt.Sprint((*data)["Information"].(map[string]int)["Urgent"])},
		{Description: "Informational proposed not exploitable issues", Details: fmt.Sprint((*data)["Information"].(map[string]int)["ProposedNotExploitable"])},
		{Description: "Informational to verify issues", Details: fmt.Sprint((*data)["Information"].(map[string]int)["ToVerify"])},
	}
	for _, detailRow := range detailRows {
		row := reporting.ScanRow{}
		row.AddColumn(detailRow.Description, 0)
		row.AddColumn(detailRow.Details, 0)

		detailTable.Rows = append(detailTable.Rows, row)
	}
	scanReport.DetailTable = detailTable

	return scanReport
}

func CreateJSONHeaderReport(data *map[string]interface{}) CheckmarxOneReportData {
	checkmarxReportData := CheckmarxOneReportData{
		ToolName:        `CheckmarxOne`,
		ProjectName:     fmt.Sprint((*data)["ProjectName"]),
		GroupID:         fmt.Sprint((*data)["Group"]),
		GroupName:       fmt.Sprint((*data)["GroupFullPathOnReportDate"]),
		ApplicationID:   fmt.Sprint((*data)["Application"]),
		ApplicationName: fmt.Sprint((*data)["ApplicationFullPathOnReportDate"]),
		DeepLink:        fmt.Sprint((*data)["DeepLink"]),
		Preset:          fmt.Sprint((*data)["Preset"]),
		ToolVersion:     fmt.Sprint((*data)["ToolVersion"]),
		ScanType:        fmt.Sprint((*data)["ScanType"]),
		ProjectID:       fmt.Sprint((*data)["ProjectId"]),
		ScanID:          fmt.Sprint((*data)["ScanId"]),
	}

	findings := []Finding{}
	// High
	highFindings := Finding{}
	highFindings.ClassificationName = "High"
	highFindings.Total = (*data)["High"].(map[string]int)["Issues"]
	highAudited := (*data)["High"].(map[string]int)["Issues"] - (*data)["High"].(map[string]int)["NotFalsePositive"]
	highFindings.Audited = &highAudited
	findings = append(findings, highFindings)
	// Medium
	mediumFindings := Finding{}
	mediumFindings.ClassificationName = "Medium"
	mediumFindings.Total = (*data)["Medium"].(map[string]int)["Issues"]
	mediumAudited := (*data)["Medium"].(map[string]int)["Issues"] - (*data)["Medium"].(map[string]int)["NotFalsePositive"]
	mediumFindings.Audited = &mediumAudited
	findings = append(findings, mediumFindings)
	// Low
	lowFindings := Finding{}
	lowFindings.ClassificationName = "Low"
	if _, ok := (*data)["LowPerQuery"]; ok {
		lowPerQueryList := []LowPerQuery{}
		lowPerQueryMap := (*data)["LowPerQuery"].(map[string]map[string]int)
		for queryName, resultsLowQuery := range lowPerQueryMap {
			audited := resultsLowQuery["Confirmed"] + resultsLowQuery["NotExploitable"]
			total := resultsLowQuery["Issues"]
			lowPerQuery := LowPerQuery{}
			lowPerQuery.QueryName = queryName
			lowPerQuery.Audited = audited
			lowPerQuery.Total = total
			lowPerQueryList = append(lowPerQueryList, lowPerQuery)
		}
		lowFindings.LowPerQuery = &lowPerQueryList
		findings = append(findings, lowFindings)
	} else {
		lowFindings.Total = (*data)["Low"].(map[string]int)["Issues"]
		lowAudited := (*data)["Low"].(map[string]int)["Confirmed"] + (*data)["Low"].(map[string]int)["NotExploitable"]
		lowFindings.Audited = &lowAudited
		findings = append(findings, lowFindings)
	}

	checkmarxReportData.Findings = &findings

	return checkmarxReportData
}

func WriteJSONHeaderReport(jsonReport CheckmarxOneReportData) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	// Standard JSON Report
	jsonComplianceReportPath := filepath.Join(ReportsDirectory, "piper_checkmarxone_report.json")
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}

	file, _ := json.Marshal(jsonReport)
	if err := utils.FileWrite(jsonComplianceReportPath, file, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write CheckmarxOne JSON compliance report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "CheckmarxOne JSON Compliance Report", Target: jsonComplianceReportPath})

	return reportPaths, nil
}

// WriteSarif writes a json file to disk as a .sarif if it respects the specification declared in format.SARIF
func WriteSarif(sarif format.SARIF) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	sarifReportPath := filepath.Join(ReportsDirectory, "result.sarif")
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}

	// HTML characters will most likely be present: we need to use encode: create a buffer to hold JSON data
	buffer := new(bytes.Buffer)
	// create JSON encoder for buffer
	bufEncoder := json.NewEncoder(buffer)
	// set options
	bufEncoder.SetEscapeHTML(false)
	bufEncoder.SetIndent("", "  ")
	//encode to buffer
	bufEncoder.Encode(sarif)
	log.Entry().Info("Writing file to disk: ", sarifReportPath)
	if err := utils.FileWrite(sarifReportPath, buffer.Bytes(), 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write CheckmarxOne SARIF report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "CheckmarxOne SARIF Report", Target: sarifReportPath})

	return reportPaths, nil
}

func WriteCustomReports(scanReport reporting.ScanReport, projectName, projectID string) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := filepath.Join(ReportsDirectory, "piper_checkmarxone_report.html")
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write html report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "CheckmarxOne Report", Target: htmlReportPath})

	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("checkmarxOneExecuteScan_sast_%v.json", reportShaCheckmarxOne([]string{projectName, projectID}))), jsonReport, 0666); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to write json report")
	}
	// we do not add the json report to the overall list of reports for now,
	// since it is just an intermediary report used as input for later
	// and there does not seem to be real benefit in archiving it.

	return reportPaths, nil
}

func reportShaCheckmarxOne(parts []string) string {
	reportShaData := []byte(strings.Join(parts, ","))
	return fmt.Sprintf("%x", sha1.Sum(reportShaData))
}
