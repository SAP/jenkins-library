package fortify

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pipergithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/piper-validation/fortify-client-go/models"

	"github.com/pkg/errors"
)

type FortifyReportData struct {
	ToolName                            string                  `json:"toolName"`
	ToolInstance                        string                  `json:"toolInstance"`
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
	URL                                 string                  `json:"url"`
	SpotChecksCategories                *[]SpotChecksAuditCount `json:"spotChecksCategories"`
}

type SpotChecksAuditCount struct {
	Audited int    `json:"spotChecksCategories"`
	Total   int    `json:"total"`
	Type    string `json:"type"`
}

func CreateCustomReport(data FortifyReportData, issueGroups []*models.ProjectVersionIssueGroup) reporting.ScanReport {

	scanReport := reporting.ScanReport{
		Title: "Fortify SAST Report",
		Subheaders: []reporting.Subheader{
			{Description: "Fortify project name", Details: data.ProjectName},
			{Description: "Fortify project version", Details: data.ProjectVersion},
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
	for _, spotChecksElement := range spotChecksCountByCategory {
		if spotChecksElement.Total > 0 && spotChecksElement.Audited == 0 {
			reportData.AtleastOneSpotChecksCategoryAudited = false
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

func UploadReportToGithub(scanReport reporting.ScanReport, token, APIURL, owner, repository string, assignees []string) error {
	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	markdownReport, _ := scanReport.ToMarkdown()
	err :=pipergithub.CreateIssue(token, APIURL, owner, repository, "Fortify SAST Results", markdownReport, assignees, true)
	if err != nil {
		return errors.Wrap(err, "failed to upload fortify results into GitHub issue")
	}
	return nil
}

func reportShaFortify(parts []string) string {
	reportShaData := []byte(strings.Join(parts, ","))
	return fmt.Sprintf("%x", sha1.Sum(reportShaData))
}
