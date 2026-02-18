package checkmarx

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
)

type CheckmarxReportData struct {
	ToolName             string         `json:"toolName"`
	ProjectName          string         `json:"projectName"`
	ProjectID            int64          `json:"projectID"`
	ScanID               int64          `json:"scanID"`
	TeamName             string         `json:"teamName"`
	TeamPath             string         `json:"teamPath"`
	DeepLink             string         `json:"deepLink"`
	Preset               string         `json:"preset"`
	CheckmarxVersion     string         `json:"checkmarxVersion"`
	ScanType             string         `json:"scanType"`
	HighTotal            int            `json:"highTotal"`
	HighAudited          int            `json:"highAudited"`
	MediumTotal          int            `json:"mediumTotal"`
	MediumAudited        int            `json:"mediumAudited"`
	LowTotal             int            `json:"lowTotal"`
	LowAudited           int            `json:"lowAudited"`
	InformationTotal     int            `json:"informationTotal"`
	InformationAudited   int            `json:"informationAudited"`
	IsLowPerQueryAudited bool           `json:"isLowPerQueryAudited"`
	LowPerQuery          *[]LowPerQuery `json:"lowPerQuery"`
}

type LowPerQuery struct {
	QueryName string `json:"query"`
	Audited   int    `json:"audited"`
	Total     int    `json:"total"`
}

func CreateCustomReport(data map[string]any, insecure, neutral []string) reporting.ScanReport {
	deepLink := fmt.Sprintf(`<a href="%v" target="_blank">Link to scan in CX UI</a>`, data["DeepLink"])

	scanReport := reporting.ScanReport{
		ReportTitle: "Checkmarx SAST Report",
		Subheaders: []reporting.Subheader{
			{Description: "Project name", Details: fmt.Sprint(data["ProjectName"])},
			{Description: "Project ID", Details: fmt.Sprint(data["ProjectId"])},
			{Description: "Owner", Details: fmt.Sprint(data["Owner"])},
			{Description: "Scan ID", Details: fmt.Sprint(data["ScanId"])},
			{Description: "Team", Details: fmt.Sprint(data["Team"])},
			{Description: "Team full path", Details: fmt.Sprint(data["TeamFullPathOnReportDate"])},
			{Description: "Scan start", Details: fmt.Sprint(data["ScanStart"])},
			{Description: "Scan duration", Details: fmt.Sprint(data["ScanTime"])},
			{Description: "Scan type", Details: fmt.Sprint(data["ScanType"])},
			{Description: "Preset", Details: fmt.Sprint(data["Preset"])},
			{Description: "Report creation time", Details: fmt.Sprint(data["ReportCreationTime"])},
			{Description: "Lines of code scanned", Details: fmt.Sprint(data["LinesOfCodeScanned"])},
			{Description: "Files scanned", Details: fmt.Sprint(data["FilesScanned"])},
			{Description: "Checkmarx version", Details: fmt.Sprint(data["CheckmarxVersion"])},
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
		{Description: "High issues", Details: fmt.Sprint(data["High"].(map[string]int)["Issues"])},
		{Description: "High not false positive issues", Details: fmt.Sprint(data["High"].(map[string]int)["NotFalsePositive"])},
		{Description: "High not exploitable issues", Details: fmt.Sprint(data["High"].(map[string]int)["NotExploitable"])},
		{Description: "High confirmed issues", Details: fmt.Sprint(data["High"].(map[string]int)["Confirmed"])},
		{Description: "High urgent issues", Details: fmt.Sprint(data["High"].(map[string]int)["Urgent"])},
		{Description: "High proposed not exploitable issues", Details: fmt.Sprint(data["High"].(map[string]int)["ProposedNotExploitable"])},
		{Description: "High to verify issues", Details: fmt.Sprint(data["High"].(map[string]int)["ToVerify"])},
		{Description: "Medium issues", Details: fmt.Sprint(data["Medium"].(map[string]int)["Issues"])},
		{Description: "Medium not false positive issues", Details: fmt.Sprint(data["Medium"].(map[string]int)["NotFalsePositive"])},
		{Description: "Medium not exploitable issues", Details: fmt.Sprint(data["Medium"].(map[string]int)["NotExploitable"])},
		{Description: "Medium confirmed issues", Details: fmt.Sprint(data["Medium"].(map[string]int)["Confirmed"])},
		{Description: "Medium urgent issues", Details: fmt.Sprint(data["Medium"].(map[string]int)["Urgent"])},
		{Description: "Medium proposed not exploitable issues", Details: fmt.Sprint(data["Medium"].(map[string]int)["ProposedNotExploitable"])},
		{Description: "Medium to verify issues", Details: fmt.Sprint(data["Medium"].(map[string]int)["ToVerify"])},
		{Description: "Low issues", Details: fmt.Sprint(data["Low"].(map[string]int)["Issues"])},
		{Description: "Low not false positive issues", Details: fmt.Sprint(data["Low"].(map[string]int)["NotFalsePositive"])},
		{Description: "Low not exploitable issues", Details: fmt.Sprint(data["Low"].(map[string]int)["NotExploitable"])},
		{Description: "Low confirmed issues", Details: fmt.Sprint(data["Low"].(map[string]int)["Confirmed"])},
		{Description: "Low urgent issues", Details: fmt.Sprint(data["Low"].(map[string]int)["Urgent"])},
		{Description: "Low proposed not exploitable issues", Details: fmt.Sprint(data["Low"].(map[string]int)["ProposedNotExploitable"])},
		{Description: "Low to verify issues", Details: fmt.Sprint(data["Low"].(map[string]int)["ToVerify"])},
		{Description: "Informational issues", Details: fmt.Sprint(data["Information"].(map[string]int)["Issues"])},
		{Description: "Informational not false positive issues", Details: fmt.Sprint(data["Information"].(map[string]int)["NotFalsePositive"])},
		{Description: "Informational not exploitable issues", Details: fmt.Sprint(data["Information"].(map[string]int)["NotExploitable"])},
		{Description: "Informational confirmed issues", Details: fmt.Sprint(data["Information"].(map[string]int)["Confirmed"])},
		{Description: "Informational urgent issues", Details: fmt.Sprint(data["Information"].(map[string]int)["Urgent"])},
		{Description: "Informational proposed not exploitable issues", Details: fmt.Sprint(data["Information"].(map[string]int)["ProposedNotExploitable"])},
		{Description: "Informational to verify issues", Details: fmt.Sprint(data["Information"].(map[string]int)["ToVerify"])},
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

func CreateJSONReport(data map[string]any) CheckmarxReportData {
	checkmarxReportData := CheckmarxReportData{
		ToolName:         `checkmarx`,
		ProjectName:      fmt.Sprint(data["ProjectName"]),
		TeamName:         fmt.Sprint(data["Team"]),
		TeamPath:         fmt.Sprint(data["TeamFullPathOnReportDate"]),
		DeepLink:         fmt.Sprint(data["DeepLink"]),
		Preset:           fmt.Sprint(data["Preset"]),
		CheckmarxVersion: fmt.Sprint(data["CheckmarxVersion"]),
		ScanType:         fmt.Sprint(data["ScanType"]),
	}

	if s, err := strconv.ParseInt(fmt.Sprint(data["ProjectId"]), 10, 64); err == nil {
		checkmarxReportData.ProjectID = s
	}

	if s, err := strconv.ParseInt(fmt.Sprint(data["ScanId"]), 10, 64); err == nil {
		checkmarxReportData.ScanID = s
	}

	checkmarxReportData.HighAudited = data["High"].(map[string]int)["Issues"] - data["High"].(map[string]int)["NotFalsePositive"]
	checkmarxReportData.HighTotal = data["High"].(map[string]int)["Issues"]

	checkmarxReportData.MediumAudited = data["Medium"].(map[string]int)["Issues"] - data["Medium"].(map[string]int)["NotFalsePositive"]
	checkmarxReportData.MediumTotal = data["Medium"].(map[string]int)["Issues"]

	checkmarxReportData.LowAudited = data["Low"].(map[string]int)["Confirmed"] + data["Low"].(map[string]int)["NotExploitable"]
	checkmarxReportData.LowTotal = data["Low"].(map[string]int)["Issues"]

	checkmarxReportData.InformationAudited = data["Information"].(map[string]int)["Confirmed"] + data["Information"].(map[string]int)["NotExploitable"]
	checkmarxReportData.InformationTotal = data["Information"].(map[string]int)["Issues"]

	lowPerQueryList := []LowPerQuery{}
	checkmarxReportData.IsLowPerQueryAudited = true
	if _, ok := data["LowPerQuery"]; ok {
		lowPerQueryMap := data["LowPerQuery"].(map[string]map[string]int)
		for queryName, resultsLowQuery := range lowPerQueryMap {
			audited := resultsLowQuery["Confirmed"] + resultsLowQuery["NotExploitable"]
			total := resultsLowQuery["Issues"]
			lowPerQuery := LowPerQuery{}
			lowPerQuery.QueryName = queryName
			lowPerQuery.Audited = audited
			lowPerQuery.Total = total
			lowAuditedRequiredPerQuery := int(math.Ceil(0.10 * float64(total)))
			if audited < lowAuditedRequiredPerQuery && audited < 10 {
				checkmarxReportData.IsLowPerQueryAudited = false
			}
			lowPerQueryList = append(lowPerQueryList, lowPerQuery)
		}
	}
	checkmarxReportData.LowPerQuery = &lowPerQueryList

	return checkmarxReportData
}

func WriteJSONReport(jsonReport CheckmarxReportData) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	// Standard JSON Report
	jsonComplianceReportPath := filepath.Join(ReportsDirectory, "piper_checkmarx_report.json")
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, fmt.Errorf("failed to create report directory: %w", err)
	}

	file, _ := json.Marshal(jsonReport)
	if err := utils.FileWrite(jsonComplianceReportPath, file, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, fmt.Errorf("failed to write Checkmarx JSON compliance report: %w", err)
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Checkmarx JSON Compliance Report", Target: jsonComplianceReportPath})

	return reportPaths, nil
}

// WriteSarif writes a json file to disk as a .sarif if it respects the specification declared in format.SARIF
func WriteSarif(sarif format.SARIF) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	sarifReportPath := filepath.Join(ReportsDirectory, "result.sarif")
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
		return reportPaths, fmt.Errorf("failed to write Checkmarx SARIF report: %w", err)
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Checkmarx SARIF Report", Target: sarifReportPath})

	return reportPaths, nil
}

func WriteCustomReports(scanReport reporting.ScanReport, projectName, projectID string) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := filepath.Join(ReportsDirectory, "piper_checkmarx_report.html")
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, fmt.Errorf("failed to create report directory: %w", err)
	}
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, fmt.Errorf("failed to write html report: %w", err)
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Checkmarx Report", Target: htmlReportPath})

	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, fmt.Errorf("failed to create reporting directory: %w", err)
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("checkmarxExecuteScan_sast_%v.json", reportShaCheckmarx([]string{projectName, projectID}))), jsonReport, 0666); err != nil {
		return reportPaths, fmt.Errorf("failed to write json report: %w", err)
	}
	// we do not add the json report to the overall list of reports for now,
	// since it is just an intermediary report used as input for later
	// and there does not seem to be real benefit in archiving it.

	return reportPaths, nil
}

func reportShaCheckmarx(parts []string) string {
	reportShaData := []byte(strings.Join(parts, ","))
	return fmt.Sprintf("%x", sha1.Sum(reportShaData))
}
