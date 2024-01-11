package fortify

import (
	"bytes"
	"compress/gzip"
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
	"github.com/piper-validation/fortify-client-go/models"

	"github.com/pkg/errors"
)

type FortifyReportData struct {
	ToolName                            string                  `json:"toolName"`
	ToolInstance                        string                  `json:"toolInstance"`
	ProjectID                           int64                   `json:"projectID"`
	ProjectName                         string                  `json:"projectName"`
	ProjectVersion                      string                  `json:"projectVersion"`
	ProjectVersionID                    int64                   `json:"projectVersionID"`
	Violations                          int                     `json:"violations"`
	CorporateTotal                      int                     `json:"corporateTotal"`
	CorporateAudited                    int                     `json:"corporateAudited"`
	AuditAllTotal                       int                     `json:"auditAllTotal"`
	AuditAllAudited                     int                     `json:"auditAllAudited"`
	SpotChecksTotal                     int                     `json:"spotChecksTotal"`
	SpotChecksAudited                   int                     `json:"spotChecksAudited"`
	SpotChecksGap                       int                     `json:"spotChecksGap"`
	Suspicious                          int                     `json:"suspicious"`
	Exploitable                         int                     `json:"exploitable"`
	Suppressed                          int                     `json:"suppressed"`
	AtleastOneSpotChecksCategoryAudited bool                    `json:"atleastOneSpotChecksCategoryAudited"`
	IsSpotChecksPerCategoryAudited      bool                    `json:"isSpotChecksPerCategoryAudited"`
	URL                                 string                  `json:"url"`
	SpotChecksCategories                *[]SpotChecksAuditCount `json:"spotChecksCategories"`
}

type SpotChecksAuditCount struct {
	Audited int    `json:"audited"`
	Total   int    `json:"total"`
	Type    string `json:"type"`
}

func CreateCustomReport(data FortifyReportData, issueGroups []*models.ProjectVersionIssueGroup) reporting.ScanReport {

	scanReport := reporting.ScanReport{
		ReportTitle: "Fortify SAST Report",
		Subheaders: []reporting.Subheader{
			{Description: "Fortify project name", Details: data.ProjectName},
			{Description: "Fortify project version", Details: data.ProjectVersion},
			{Description: "Fortify URL", Details: data.URL},
		},
		Overview: []reporting.OverviewRow{
			{Description: "Number of compliance violations", Details: fmt.Sprint(data.Violations)},
			{Description: "Number of issues suppressed", Details: fmt.Sprint(data.Suppressed)},
			{Description: "Unaudited corporate issues", Details: fmt.Sprint(data.CorporateTotal - data.CorporateAudited)},
			{Description: "Unaudited audit all issues", Details: fmt.Sprint(data.AuditAllTotal - data.AuditAllAudited)},
			{Description: "Unaudited spot check issues", Details: fmt.Sprint(data.SpotChecksTotal - data.SpotChecksAudited)},
			{Description: "Number of suspicious issues", Details: fmt.Sprint(data.Suspicious)},
			{Description: "Number of exploitable issues", Details: fmt.Sprint(data.Exploitable)},
		},
		ReportTime: time.Now(),
	}

	detailTable := reporting.ScanDetailTable{
		NoRowsMessage: "No findings detected",
		Headers: []string{
			"Issue group",
			"Total count",
			"Audited count",
		},
		WithCounter:   true,
		CounterHeader: "Entry #",
	}

	for _, group := range issueGroups {
		row := reporting.ScanRow{}
		row.AddColumn(fmt.Sprint(*group.CleanName), 0)
		row.AddColumn(fmt.Sprint(*group.TotalCount), 0)
		row.AddColumn(fmt.Sprint(*group.AuditedCount), 0)

		detailTable.Rows = append(detailTable.Rows, row)
	}

	scanReport.DetailTable = detailTable
	scanReport.SuccessfulScan = data.Violations == 0

	return scanReport
}

func CreateJSONReport(reportData FortifyReportData, spotChecksCountByCategory []SpotChecksAuditCount, serverURL string) FortifyReportData {
	reportData.AtleastOneSpotChecksCategoryAudited = true
	reportData.IsSpotChecksPerCategoryAudited = true
	for _, spotChecksElement := range spotChecksCountByCategory {
		if spotChecksElement.Total > 0 && spotChecksElement.Audited == 0 {
			reportData.AtleastOneSpotChecksCategoryAudited = false
		}

		spotCheckMinimumPercentageValue := int(math.Ceil(float64(0.10 * float64(spotChecksElement.Total))))
		if spotChecksElement.Audited < spotCheckMinimumPercentageValue && spotChecksElement.Audited < 10 {
			reportData.IsSpotChecksPerCategoryAudited = false
		}

		if !reportData.IsSpotChecksPerCategoryAudited && !reportData.AtleastOneSpotChecksCategoryAudited {
			break
		}
	}

	reportData.SpotChecksCategories = &spotChecksCountByCategory
	reportData.URL = serverURL + "/html/ssc/version/" + strconv.FormatInt(reportData.ProjectVersionID, 10)
	reportData.ToolInstance = serverURL
	reportData.ToolName = "fortify"

	return reportData
}

func WriteJSONReport(jsonReport FortifyReportData) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	// Standard JSON Report
	jsonComplianceReportPath := filepath.Join(ReportsDirectory, "piper_fortify_report.json")
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}

	file, _ := json.Marshal(jsonReport)
	if err := utils.FileWrite(jsonComplianceReportPath, file, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write fortify json compliance report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Fortify JSON Compliance Report", Target: jsonComplianceReportPath})

	return reportPaths, nil
}

func WriteSarif(sarif format.SARIF, fileName string) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	sarifReportPath := filepath.Join(ReportsDirectory, fileName)
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}

	// This solution did not allow for special HTML characters. If this causes any issue, revert l148-l157 with these two
	/*file, _ := json.MarshalIndent(sarif, "", "  ")
	if err := utils.FileWrite(sarifReportPath, file, 0666); err != nil {*/

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
		return reportPaths, errors.Wrapf(err, "failed to write fortify SARIF report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Fortify SARIF Report", Target: sarifReportPath})

	return reportPaths, nil
}

func WriteGzipSarif(sarif format.SARIF, fileName string) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	sarifReportPath := filepath.Join(ReportsDirectory, fileName)
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}

	// HTML characters will most likely be present: we need to use encode: create a buffer to hold JSON data
	// https://stackoverflow.com/questions/28595664/how-to-stop-json-marshal-from-escaping-and
	buffer := new(bytes.Buffer)
	// create JSON encoder for buffer
	bufEncoder := json.NewEncoder(buffer)
	// set options
	bufEncoder.SetEscapeHTML(false)
	bufEncoder.SetIndent("", "  ")
	//encode to buffer
	bufEncoder.Encode(sarif)

	// Initialize gzip
	gzBuffer := &bytes.Buffer{}
	gzWriter := gzip.NewWriter(gzBuffer)
	gzWriter.Write([]byte(buffer.Bytes()))
	gzWriter.Close()

	log.Entry().Info("Writing file to disk: ", sarifReportPath)
	if err := utils.FileWrite(sarifReportPath, gzBuffer.Bytes(), 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write Fortify SARIF gzip report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Fortify SARIF gzip Report", Target: sarifReportPath})

	return reportPaths, nil
}

func WriteCustomReports(scanReport reporting.ScanReport) ([]piperutils.Path, error) {
	utils := piperutils.Files{}
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := filepath.Join(ReportsDirectory, "piper_fortify_report.html")
	// Ensure reporting directory exists
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write html report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Fortify Report", Target: htmlReportPath})

	// we do not add the json report to the overall list of reports for now,
	// since it is just an intermediary report used as input for later
	// and there does not seem to be real benefit in archiving it.
	return reportPaths, nil
}

func reportShaFortify(parts []string) string {
	reportShaData := []byte(strings.Join(parts, ","))
	return fmt.Sprintf("%x", sha1.Sum(reportShaData))
}
