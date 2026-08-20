package checkmarxOne

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
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
	Confirmed          int            `json:"confirmed,omitempty"`
	LowPerQuery        *[]LowPerQuery `json:"categories,omitempty"`
}

type LowPerQuery struct {
	QueryName string `json:"name"`
	Audited   int    `json:"audited"`
	Total     int    `json:"total"`
	Confirmed int    `json:"confirmed"`
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
			{Description: "Preset", Details: fmt.Sprint((*data)["SastPreset"])},
			{Description: "IAC Preset", Details: fmt.Sprint((*data)["IacPreset"])},
			{Description: "Report creation time", Details: fmt.Sprint((*data)["ReportCreationTime"])},
			{Description: "Lines of code scanned", Details: fmt.Sprint((*data)["LinesOfCodeScanned"])},
			{Description: "Files scanned", Details: fmt.Sprint((*data)["FilesScanned"])},
			{Description: "IAC Lines of code scanned", Details: fmt.Sprint((*data)["IacLinesOfCodeScanned"])},
			{Description: "IAC Files scanned", Details: fmt.Sprint((*data)["IacFilesScanned"])},
			{Description: "Tool version", Details: fmt.Sprintf("%s, %s, %s", (*data)["ToolVersion"], (*data)["SASTVersion"], (*data)["IACVersion"])},
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

	getCount := func(severity, key string) string {
		count := 0

		if m, ok := (*data)[severity]; ok {
			if m, ok := m.(map[string]int); ok {
				count += m[key]
			}
		}
		if m, ok := (*data)["IAC"+severity]; ok {
			if m, ok := m.(map[string]int); ok {
				count += m[key]
			}
		}
		return fmt.Sprint(count)
	}

	detailRows := []reporting.OverviewRow{
		{Description: "Critical issues", Details: getCount("Critical", "Issues")},
		{Description: "Critical not false positive issues", Details: getCount("Critical", "NotFalsePositive")},
		{Description: "Critical not exploitable issues", Details: getCount("Critical", "NotExploitable")},
		{Description: "Critical confirmed issues", Details: getCount("Critical", "Confirmed")},
		{Description: "Critical urgent issues", Details: getCount("Critical", "Urgent")},
		{Description: "Critical proposed not exploitable issues", Details: getCount("Critical", "ProposedNotExploitable")},
		{Description: "Critical to verify issues", Details: getCount("Critical", "ToVerify")},
		{Description: "High issues", Details: getCount("High", "Issues")},
		{Description: "High not false positive issues", Details: getCount("High", "NotFalsePositive")},
		{Description: "High not exploitable issues", Details: getCount("High", "NotExploitable")},
		{Description: "High confirmed issues", Details: getCount("High", "Confirmed")},
		{Description: "High urgent issues", Details: getCount("High", "Urgent")},
		{Description: "High proposed not exploitable issues", Details: getCount("High", "ProposedNotExploitable")},
		{Description: "High to verify issues", Details: getCount("High", "ToVerify")},
		{Description: "Medium issues", Details: getCount("Medium", "Issues")},
		{Description: "Medium not false positive issues", Details: getCount("Medium", "NotFalsePositive")},
		{Description: "Medium not exploitable issues", Details: getCount("Medium", "NotExploitable")},
		{Description: "Medium confirmed issues", Details: getCount("Medium", "Confirmed")},
		{Description: "Medium urgent issues", Details: getCount("Medium", "Urgent")},
		{Description: "Medium proposed not exploitable issues", Details: getCount("Medium", "ProposedNotExploitable")},
		{Description: "Medium to verify issues", Details: getCount("Medium", "ToVerify")},
		{Description: "Low issues", Details: getCount("Low", "Issues")},
		{Description: "Low not false positive issues", Details: getCount("Low", "NotFalsePositive")},
		{Description: "Low not exploitable issues", Details: getCount("Low", "NotExploitable")},
		{Description: "Low confirmed issues", Details: getCount("Low", "Confirmed")},
		{Description: "Low urgent issues", Details: getCount("Low", "Urgent")},
		{Description: "Low proposed not exploitable issues", Details: getCount("Low", "ProposedNotExploitable")},
		{Description: "Low to verify issues", Details: getCount("Low", "ToVerify")},
		{Description: "Informational issues", Details: getCount("Information", "Issues")},
		{Description: "Informational not false positive issues", Details: getCount("Information", "NotFalsePositive")},
		{Description: "Informational not exploitable issues", Details: getCount("Information", "NotExploitable")},
		{Description: "Informational confirmed issues", Details: getCount("Information", "Confirmed")},
		{Description: "Informational urgent issues", Details: getCount("Information", "Urgent")},
		{Description: "Informational proposed not exploitable issues", Details: getCount("Information", "ProposedNotExploitable")},
		{Description: "Informational to verify issues", Details: getCount("Information", "ToVerify")},
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

func CreateJSONHeaderReport(data *map[string]interface{}, engine string) CheckmarxOneReportData {
	checkmarxReportData := CheckmarxOneReportData{
		ToolName:        `CheckmarxOne`,
		ProjectName:     fmt.Sprint((*data)["ProjectName"]),
		GroupID:         fmt.Sprint((*data)["Group"]),
		GroupName:       fmt.Sprint((*data)["GroupFullPathOnReportDate"]),
		ApplicationID:   fmt.Sprint((*data)["Application"]),
		ApplicationName: fmt.Sprint((*data)["ApplicationFullPathOnReportDate"]),
		DeepLink:        fmt.Sprint((*data)["DeepLink"]),
		Preset:          fmt.Sprint((*data)["SastPreset"]),
		ToolVersion:     fmt.Sprint((*data)["ToolVersion"]),
		ScanType:        fmt.Sprint((*data)["ScanType"]),
		ProjectID:       fmt.Sprint((*data)["ProjectId"]),
		ScanID:          fmt.Sprint((*data)["ScanId"]),
	}

	findings := []Finding{}
	pre := ""
	if strings.EqualFold(engine, "iac") {
		pre = "IAC"
		checkmarxReportData.Preset = fmt.Sprint((*data)["IacPreset"])
		checkmarxReportData.ScanType = "Full"
		checkmarxReportData.ToolVersion += ", " + fmt.Sprint((*data)["IACVersion"])
	} else {
		checkmarxReportData.ToolVersion += ", " + fmt.Sprint((*data)["SASTVersion"])
	}
	getCount := func(severity, key string) int {
		count := 0

		if m, ok := (*data)[pre+severity]; ok {
			if m, ok := m.(map[string]int); ok {
				count += m[key]
			}
		}
		return count
	}

	// Critical
	criticalFindings := Finding{}
	criticalFindings.ClassificationName = "Critical"
	criticalFindings.Total = getCount("Critical", "Issues")
	criticalAudited := getCount("Critical", "NotExploitable") + getCount("Critical", "Urgent") + getCount("Critical", "Confirmed")
	criticalFindings.Audited = &criticalAudited
	criticalFindings.Confirmed = getCount("Critical", "Confirmed") + getCount("Critical", "Urgent")
	findings = append(findings, criticalFindings)
	// High
	highFindings := Finding{}
	highFindings.ClassificationName = "High"
	highFindings.Total = getCount("High", "Issues")
	highAudited := getCount("High", "NotExploitable") + getCount("High", "Urgent") + getCount("High", "Confirmed")
	highFindings.Audited = &highAudited
	highFindings.Confirmed = getCount("High", "Confirmed") + getCount("High", "Urgent")
	findings = append(findings, highFindings)
	// Medium
	mediumFindings := Finding{}
	mediumFindings.ClassificationName = "Medium"
	mediumFindings.Total = getCount("Medium", "Issues")
	mediumAudited := getCount("Medium", "NotExploitable") + getCount("Medium", "Urgent") + getCount("Medium", "Confirmed")
	mediumFindings.Audited = &mediumAudited
	mediumFindings.Confirmed = getCount("Medium", "Confirmed") + getCount("Medium", "Urgent")
	findings = append(findings, mediumFindings)
	// Low
	lowFindings := Finding{}
	lowFindings.ClassificationName = "Low"

	if _, ok := (*data)[pre+"LowPerQuery"]; ok {
		lowPerQueryList := []LowPerQuery{}
		lowPerQueryMap := (*data)[pre+"LowPerQuery"].(map[string]map[string]int)
		for queryName, resultsLowQuery := range lowPerQueryMap {
			audited := resultsLowQuery["Confirmed"] + resultsLowQuery["NotExploitable"] + resultsLowQuery["Urgent"]
			total := resultsLowQuery["Issues"]
			lowPerQuery := LowPerQuery{}
			lowPerQuery.QueryName = queryName
			lowPerQuery.Audited = audited
			lowPerQuery.Confirmed = resultsLowQuery["Confirmed"] + resultsLowQuery["Urgent"]
			lowPerQuery.Total = total
			lowPerQueryList = append(lowPerQueryList, lowPerQuery)
		}
		lowFindings.LowPerQuery = &lowPerQueryList
		sort.Slice(lowPerQueryList, func(i, j int) bool {
			return lowPerQueryList[i].QueryName < lowPerQueryList[j].QueryName
		})
		findings = append(findings, lowFindings)
	} else {
		lowFindings.Total = getCount("Low", "Issues")
		lowAudited := getCount("Low", "Confirmed") + getCount("Low", "NotExploitable") + getCount("Low", "Urgent")
		lowFindings.Confirmed = getCount("Low", "Confirmed") + getCount("Low", "Urgent")
		lowFindings.Audited = &lowAudited
		findings = append(findings, lowFindings)
	}

	checkmarxReportData.Findings = &findings

	return checkmarxReportData
}

func WriteJSONHeaderReport(jsonReport CheckmarxOneReportData, engine string) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	filename := "piper_checkmarxone_report.json"
	reportName := "CheckmarxOne JSON compliance report"
	if strings.EqualFold(engine, "iac") {
		filename = "piper_checkmarxone_iac_report.json"
		reportName = "CheckmarxOne IAC JSON compliance report"
	}
	// Standard JSON Report
	jsonComplianceReportPath := filepath.Join(ReportsDirectory, filename)
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, fmt.Errorf("failed to create report directory: %w", err)
	}

	file, _ := json.Marshal(jsonReport)
	if err := utils.FileWrite(jsonComplianceReportPath, file, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, fmt.Errorf("failed to write %s: %w", reportName, err)
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: reportName, Target: jsonComplianceReportPath})

	return reportPaths, nil
}

// WriteSarif writes a json file to disk as a .sarif if it respects the specification declared in format.SARIF
func WriteSASTSarif(sarif format.SARIF) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	sarifReportPath := filepath.Join(ReportsDirectory, "result-sast.sarif")
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, fmt.Errorf("failed to create report directory: %w", err)
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
		return reportPaths, fmt.Errorf("failed to write CheckmarxOne SAST SARIF report: %w", err)
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "CheckmarxOne SAST SARIF Report", Target: sarifReportPath})

	return reportPaths, nil
}

// WriteSarif writes a json file to disk as a .sarif if it respects the specification declared in format.SARIF
func WriteIACSarif(sarif format.SARIF) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	sarifReportPath := filepath.Join(ReportsDirectory, "result-iac.sarif")
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, fmt.Errorf("failed to create report directory: %w", err)
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
		return reportPaths, fmt.Errorf("failed to write CheckmarxOne SAST SARIF report: %w", err)
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "CheckmarxOne SAST SARIF Report", Target: sarifReportPath})

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
		return reportPaths, fmt.Errorf("failed to create report directory: %w", err)
	}
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, fmt.Errorf("failed to write html report: %w", err)
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "CheckmarxOne Report", Target: htmlReportPath})

	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, fmt.Errorf("failed to create reporting directory: %w", err)
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("checkmarxOneExecuteScan_sast_%v.json", reportShaCheckmarxOne([]string{projectName, projectID}))), jsonReport, 0666); err != nil {
		return reportPaths, fmt.Errorf("failed to write json report: %w", err)
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
