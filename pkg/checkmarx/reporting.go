package checkmarx

import (
	"crypto/sha1"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"

	"github.com/pkg/errors"
)

func CreateCustomReport(data map[string]interface{}) reporting.ScanReport {

	scanReport := reporting.ScanReport{
		Title: "Checkmarx SAST Report",
		Subheaders: []reporting.Subheader{
			{Description: "Project name", Details: fmt.Sprint(data["ProjectName"])},
			{Description: "Project ID", Details: fmt.Sprint(data["ProjectID"])},
			{Description: "Owner", Details: fmt.Sprint(data["Owner"])},
			{Description: "Scan ID", Details: fmt.Sprint(data["ScanID"])},
			{Description: "Team", Details: fmt.Sprint(data["Team"])},
			{Description: "Team full path", Details: fmt.Sprint(data["TeamFullPathOnReportDate"])},
			{Description: "Scan start", Details: fmt.Sprint(data["ScanStart"])},
			{Description: "Scan duration", Details: fmt.Sprint(data["ScanTime"])},
			{Description: "Scan type", Details: fmt.Sprint(data["ScanType"])},
			{Description: "Preset", Details: fmt.Sprint(data["Preset"])},
			{Description: "Report creation time", Details: fmt.Sprint(data["ReportCreationTime"])},
			{Description: "Lines of code scanned", Details: fmt.Sprint(data["LinesOfCodeScanned)"])},
			{Description: "Files scanned", Details: fmt.Sprint(data["FilesScanned)"])},
			{Description: "Checkmarx version", Details: fmt.Sprint(data["CheckmarxVersion"])},
			{Description: "Deep link", Details: fmt.Sprint(data["DeepLink"])},
		},
		Overview: []reporting.OverviewRow{
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
			{Description: "Information issues", Details: fmt.Sprint(data["Information"].(map[string]int)["Issues"])},
			{Description: "Information not false positive issues", Details: fmt.Sprint(data["Information"].(map[string]int)["NotFalsePositive"])},
			{Description: "Information not exploitable issues", Details: fmt.Sprint(data["Information"].(map[string]int)["NotExploitable"])},
			{Description: "Information confirmed issues", Details: fmt.Sprint(data["Information"].(map[string]int)["Confirmed"])},
			{Description: "Information urgent issues", Details: fmt.Sprint(data["Information"].(map[string]int)["Urgent"])},
			{Description: "Information proposed not exploitable issues", Details: fmt.Sprint(data["Information"].(map[string]int)["ProposedNotExploitable"])},
			{Description: "Information to verify issues", Details: fmt.Sprint(data["Information"].(map[string]int)["ToVerify"])},
		},
		ReportTime: time.Now(),
	}

	return scanReport
}

func WriteCustomReports(scanReport reporting.ScanReport, projectName, projectID string) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := filepath.Join(ReportsDirectory, "piper_checkmarx_report.html")
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write html report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Checkmarx Report", Target: htmlReportPath})

	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("checkmarxExecuteScan_sast_%v.json", reportShaFortify([]string{projectName, projectID}))), jsonReport, 0666); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to write json report")
	}
	// we do not add the json report to the overall list of reports for now,
	// since it is just an intermediary report used as input for later
	// and there does not seem to be real benefit in archiving it.

	return reportPaths, nil
}

func reportShaFortify(parts []string) string {
	reportShaData := []byte(strings.Join(parts, ","))
	return fmt.Sprintf("%x", sha1.Sum(reportShaData))
}
