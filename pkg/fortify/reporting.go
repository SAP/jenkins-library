package fortify

import (
	"crypto/sha1"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/piper-validation/fortify-client-go/models"

	"github.com/pkg/errors"
)

type FortifyReportData struct {
	ProjectName       string
	ProjectVersion    string
	Violations        int
	CorporateTotal    int
	CorporateAudited  int
	AuditAllTotal     int
	AuditAllAudited   int
	SpotChecksTotal   int
	SpotChecksAudited int
	SpotChecksGap     int
	Suspicious        int
	Exploitable       int
	Suppressed        int
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

	return scanReport
}

func WriteCustomReports(scanReport reporting.ScanReport, projectName, projectVersion string) ([]piperutils.Path, error) {
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

	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("fortifyExecuteScan_sast_%v.json", reportShaFortify([]string{projectName, projectVersion}))), jsonReport, 0666); err != nil {
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
