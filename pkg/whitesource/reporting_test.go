package whitesource

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/stretchr/testify/assert"
)

func TestCreateCustomVulnerabilityReport(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		config := &ScanOptions{}
		scan := &Scan{
			AggregateProjectName: config.ProjectName,
			ProductVersion:       config.ProductVersion,
		}
		scan.AppendScannedProject("testProject")
		alerts := []Alert{
			{Library: Library{Filename: "vul1"}, Vulnerability: Vulnerability{CVSS3Score: 7.0, Score: 6}},
			{Library: Library{Filename: "vul2"}, Vulnerability: Vulnerability{CVSS3Score: 8.0, TopFix: Fix{Message: "this is the top fix"}}},
			{Library: Library{Filename: "vul3"}, Vulnerability: Vulnerability{Score: 6}},
		}

		scanReport := CreateCustomVulnerabilityReport(config.ProductName, scan, &alerts, 7.0)

		assert.Equal(t, "WhiteSource Security Vulnerability Report", scanReport.Title)
		assert.Equal(t, 3, len(scanReport.DetailTable.Rows))

		// assert that library info is filled and sorting has been executed
		assert.Equal(t, "vul2", scanReport.DetailTable.Rows[0].Columns[5].Content)
		assert.Equal(t, "vul1", scanReport.DetailTable.Rows[1].Columns[5].Content)
		assert.Equal(t, "vul3", scanReport.DetailTable.Rows[2].Columns[5].Content)

		// assert that CVSS version identification has been done
		assert.Equal(t, "v3", scanReport.DetailTable.Rows[0].Columns[3].Content)
		assert.Equal(t, "v3", scanReport.DetailTable.Rows[1].Columns[3].Content)
		assert.Equal(t, "v2", scanReport.DetailTable.Rows[2].Columns[3].Content)

		// assert proper rating and styling of high prio issues
		assert.Equal(t, "8", scanReport.DetailTable.Rows[0].Columns[2].Content)
		assert.Equal(t, "7", scanReport.DetailTable.Rows[1].Columns[2].Content)
		assert.Equal(t, "6", scanReport.DetailTable.Rows[2].Columns[2].Content)
		assert.Equal(t, "red-cell", scanReport.DetailTable.Rows[0].Columns[2].Style.String())
		assert.Equal(t, "red-cell", scanReport.DetailTable.Rows[1].Columns[2].Style.String())
		assert.Equal(t, "yellow-cell", scanReport.DetailTable.Rows[2].Columns[2].Style.String())

		assert.Contains(t, scanReport.DetailTable.Rows[0].Columns[10].Content, "this is the top fix")

	})
}

func TestCreateSarifResultFile(t *testing.T) {
	scan := &Scan{ProductVersion: "1"}
	scan.AppendScannedProject("project1")
	scan.AgentName = "Some test agent"
	scan.AgentVersion = "1.2.6"
	alerts := []Alert{
		{Library: Library{Filename: "vul1", ArtifactID: "org.some.lib"}, Vulnerability: Vulnerability{CVSS3Score: 7.0, Score: 6}},
		{Library: Library{Filename: "vul2", ArtifactID: "org.some.lib"}, Vulnerability: Vulnerability{CVSS3Score: 8.0, TopFix: Fix{Message: "this is the top fix"}}},
		{Library: Library{Filename: "vul3", ArtifactID: "org.some.lib2"}, Vulnerability: Vulnerability{Score: 6}},
	
	}

	sarif := CreateSarifResultFile(scan, &alerts)

	assert.Equal(t, "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos01/schemas/sarif-schema-2.1.0.json", sarif.Schema)
	assert.Equal(t, "2.1.0", sarif.Version)
	assert.Equal(t, 1, len(sarif.Runs))
	assert.Equal(t, "Some test agent", sarif.Runs[0].Tool.Driver.Name)
	assert.Equal(t, "1.2.6", sarif.Runs[0].Tool.Driver.Version)
	assert.Equal(t, 3, len(sarif.Runs[0].Tool.Driver.Rules))
	assert.Equal(t, 3, len(sarif.Runs[0].Results))
	// TODO add more extensive verification once we agree on the format details
}

func TestWriteCustomVulnerabilityReports(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		productName := "mock-product"
		scan := &Scan{ProductVersion: "1"}
		scan.AppendScannedProject("project1")
		scan.AppendScannedProject("project2")

		scanReport := reporting.ScanReport{}
		var utilsMock piperutils.FileUtils
		utilsMock = &mock.FilesMock{}

		reportPaths, err := WriteCustomVulnerabilityReports(productName, scan, scanReport, utilsMock)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(reportPaths))

		exists, err := utilsMock.FileExists(reportPaths[0].Target)
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = utilsMock.FileExists(filepath.Join(reporting.StepReportDirectory, "whitesourceExecuteScan_oss_27322f16a39c10c852ba6639538140a03e08e93f.json"))
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("failed to write HTML report", func(t *testing.T) {
		productName := "mock-product"
		scan := &Scan{ProductVersion: "1"}
		scanReport := reporting.ScanReport{}
		utilsMock := &mock.FilesMock{}
		utilsMock.FileWriteErrors = map[string]error{
			filepath.Join(ReportsDirectory, "piper_whitesource_vulnerability_report.html"): fmt.Errorf("write error"),
		}

		_, err := WriteCustomVulnerabilityReports(productName, scan, scanReport, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to write html report")
	})

	t.Run("failed to write json report", func(t *testing.T) {
		productName := "mock-product"
		scan := &Scan{ProductVersion: "1"}
		scan.AppendScannedProject("project1")
		scanReport := reporting.ScanReport{}
		utilsMock := &mock.FilesMock{}
		utilsMock.FileWriteErrors = map[string]error{
			filepath.Join(reporting.StepReportDirectory, "whitesourceExecuteScan_oss_e860d3a7cc8ca3261f065773404ba43e9a0b9d5b.json"): fmt.Errorf("write error"),
		}

		_, err := WriteCustomVulnerabilityReports(productName, scan, scanReport, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to write json report")
	})
}

func TestWriteSarifFile(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		sarif := format.SARIF{}
		var utilsMock piperutils.FileUtils
		utilsMock = &mock.FilesMock{}

		reportPaths, err := WriteSarifFile(&sarif, utilsMock)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(reportPaths))

		exists, err := utilsMock.FileExists(reportPaths[0].Target)
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = utilsMock.FileExists(filepath.Join(ReportsDirectory, "piper_whitesource_vulnerability.sarif"))
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("failed to write HTML report", func(t *testing.T) {
		sarif := format.SARIF{}
		utilsMock := &mock.FilesMock{}
		utilsMock.FileWriteErrors = map[string]error{
			filepath.Join(ReportsDirectory, "piper_whitesource_vulnerability.sarif"): fmt.Errorf("write error"),
		}

		_, err := WriteSarifFile(&sarif, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to write SARIF file")
	})
}

func TestCountSecurityVulnerabilities(t *testing.T) {
	t.Parallel()

	alerts := []Alert{
		{Vulnerability: Vulnerability{CVSS3Score: 7.1}},
		{Vulnerability: Vulnerability{CVSS3Score: 7}},
		{Vulnerability: Vulnerability{CVSS3Score: 6}},
	}

	severe, nonSevere := CountSecurityVulnerabilities(&alerts, 7.0)
	assert.Equal(t, 2, severe)
	assert.Equal(t, 1, nonSevere)
}

func TestIsSevereVulnerability(t *testing.T) {
	tt := []struct {
		alert    Alert
		limit    float64
		expected bool
	}{
		{alert: Alert{Vulnerability: Vulnerability{CVSS3Score: 0}}, limit: 0, expected: true},
		{alert: Alert{Vulnerability: Vulnerability{CVSS3Score: 6.9, Score: 6}}, limit: 7.0, expected: false},
		{alert: Alert{Vulnerability: Vulnerability{CVSS3Score: 7.0, Score: 6}}, limit: 7.0, expected: true},
		{alert: Alert{Vulnerability: Vulnerability{CVSS3Score: 7.1, Score: 6}}, limit: 7.0, expected: true},
		{alert: Alert{Vulnerability: Vulnerability{CVSS3Score: 6, Score: 6.9}}, limit: 7.0, expected: false},
		{alert: Alert{Vulnerability: Vulnerability{CVSS3Score: 6, Score: 7.0}}, limit: 7.0, expected: false},
		{alert: Alert{Vulnerability: Vulnerability{CVSS3Score: 6, Score: 7.1}}, limit: 7.0, expected: false},
		{alert: Alert{Vulnerability: Vulnerability{Score: 6.9}}, limit: 7.0, expected: false},
		{alert: Alert{Vulnerability: Vulnerability{Score: 7.0}}, limit: 7.0, expected: true},
		{alert: Alert{Vulnerability: Vulnerability{Score: 7.1}}, limit: 7.0, expected: true},
	}

	for i, test := range tt {
		assert.Equalf(t, test.expected, isSevereVulnerability(test.alert, test.limit), "run %v failed", i)
	}
}

func TestVulnerabilityScore(t *testing.T) {
	t.Parallel()

	tt := []struct {
		alert    Alert
		expected float64
	}{
		{alert: Alert{Vulnerability: Vulnerability{CVSS3Score: 7.0, Score: 6}}, expected: 7.0},
		{alert: Alert{Vulnerability: Vulnerability{CVSS3Score: 7.0}}, expected: 7.0},
		{alert: Alert{Vulnerability: Vulnerability{Score: 6}}, expected: 6},
	}
	for i, test := range tt {
		assert.Equalf(t, test.expected, vulnerabilityScore(test.alert), "run %v failed", i)
	}
}
