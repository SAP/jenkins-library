//go:build unit
// +build unit

package whitesource

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/versioning"
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

		assert.Equal(t, "WhiteSource Security Vulnerability Report", scanReport.Title())
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

func TestCreateCycloneSBOM(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		config := &ScanOptions{}
		scan := &Scan{
			AgentName:            "Mend Unified Agent",
			AgentVersion:         "3.3.3",
			AggregateProjectName: config.ProjectName,
			BuildTool:            "maven",
			ProductVersion:       config.ProductVersion,
			Coordinates:          versioning.Coordinates{GroupID: "com.sap", ArtifactID: "myproduct", Version: "1.3.4"},
		}
		scan.AppendScannedProject("testProject")
		alerts := []Alert{
			{Library: Library{KeyID: 42, Name: "log4j", GroupID: "apache-logging", ArtifactID: "log4j", Filename: "vul1"}, Vulnerability: Vulnerability{CVSS3Score: 7.0, Score: 6}},
			{Library: Library{KeyID: 43, Name: "commons-lang", GroupID: "apache-commons", ArtifactID: "commons-lang", Filename: "vul2"}, Vulnerability: Vulnerability{CVSS3Score: 8.0, TopFix: Fix{Message: "this is the top fix"}}},
			{Library: Library{KeyID: 42, Name: "log4j", GroupID: "apache-logging", ArtifactID: "log4j", Filename: "vul3"}, Vulnerability: Vulnerability{Score: 6}},
		}

		assessedAlerts := []Alert{
			{Library: Library{KeyID: 42, Name: "log4j", GroupID: "apache-logging", ArtifactID: "log4j", Filename: "vul4"}, Vulnerability: Vulnerability{Name: "CVE-23456", CVSS3Score: 7.0, Score: 6}, Assessment: &format.Assessment{Vulnerability: "CVE-23456", Status: format.Relevant, Analysis: format.Mitigated}},
		}

		libraries := []Library{
			{KeyID: 42, Name: "log4j", GroupID: "apache-logging", ArtifactID: "log4j", Filename: "vul1", Dependencies: []Library{{KeyID: 43, Name: "commons-lang", GroupID: "apache-commons", ArtifactID: "commons-lang", Filename: "vul2"}}},
			{KeyID: 42, Name: "log4j", GroupID: "apache-logging", ArtifactID: "log4j", Filename: "vul3"},
		}

		contents, err := CreateCycloneSBOM(scan, &libraries, &alerts, &assessedAlerts)
		assert.NoError(t, err, "unexpected error")
		buffer := bytes.NewBuffer(contents)
		decoder := cdx.NewBOMDecoder(buffer, cdx.BOMFileFormatXML)
		bom := cdx.NewBOM()
		decoder.Decode(bom)

		assert.NotNil(t, bom, "BOM was nil")
		assert.NotEmpty(t, bom.SpecVersion)

		components := *bom.Components
		vulnerabilities := *bom.Vulnerabilities
		assert.Equal(t, 2, len(components))
		assert.Equal(t, true, components[0].Name == "log4j" || components[0].Name == "commons-lang")
		assert.Equal(t, true, components[1].Name == "log4j" || components[1].Name == "commons-lang")
		assert.Equal(t, true, components[0].Name != components[1].Name)
		assert.Equal(t, 4, len(vulnerabilities))
		assert.NotNil(t, vulnerabilities[3].Analysis)
		assert.Equal(t, cdx.IAJProtectedByMitigatingControl, vulnerabilities[3].Analysis.Justification)
	})

	t.Run("success - golden", func(t *testing.T) {
		config := &ScanOptions{ProjectName: "myproduct - 1.3.4", ProductVersion: "1"}
		scan := &Scan{
			AgentName:            "Mend Unified Agent",
			AgentVersion:         "3.3.3",
			scannedProjects:      map[string]Project{"testProject": {Name: "testProject", Token: "projectToken-567"}},
			AggregateProjectName: config.ProjectName,
			BuildTool:            "maven",
			ProductVersion:       config.ProductVersion,
			ProductToken:         "productToken-123",
			Coordinates:          versioning.Coordinates{GroupID: "com.sap", ArtifactID: "myproduct", Version: "1.3.4"},
		}
		scan.AppendScannedProject("testProject")

		lib3 := Library{KeyID: 43, Name: "commons-lang", GroupID: "apache-commons", ArtifactID: "commons-lang", Version: "2.4.30", LibType: "Java", Filename: "vul2"}
		lib4 := Library{KeyID: 45, Name: "commons-lang", GroupID: "apache-commons", ArtifactID: "commons-lang", Version: "3.15", LibType: "Java", Filename: "novul"}
		lib1 := Library{KeyID: 42, Name: "log4j", GroupID: "apache-logging", ArtifactID: "log4j", Version: "1.14", LibType: "Java", Filename: "vul1", Dependencies: []Library{lib3}}
		lib2 := Library{KeyID: 44, Name: "log4j", GroupID: "apache-logging", ArtifactID: "log4j", Version: "3.25", LibType: "Java", Filename: "vul3", Dependencies: []Library{lib4}}

		alerts := []Alert{
			{Library: lib1, Vulnerability: Vulnerability{Name: "CVE-2022-001", CVSS3Score: 7, Score: 6, CVSS3Severity: "high", Severity: "medium", PublishDate: "01.01.2022"}},
			{Library: lib3, Vulnerability: Vulnerability{Name: "CVE-2022-002", CVSS3Score: 8, CVSS3Severity: "high", PublishDate: "02.01.2022", TopFix: Fix{Message: "this is the top fix"}}},
			{Library: lib2, Vulnerability: Vulnerability{Name: "CVE-2022-003", Score: 6, Severity: "medium", PublishDate: "03.01.2022"}},
		}

		assessedAlerts := []Alert{}

		libraries := []Library{
			lib1,
			lib2,
		}

		contents, err := CreateCycloneSBOM(scan, &libraries, &alerts, &assessedAlerts)
		assert.NoError(t, err, "unexpected error")

		goldenFilePath := filepath.Join("testdata", "sbom.golden")
		expected, err := os.ReadFile(goldenFilePath)
		assert.NoError(t, err)

		assert.Equal(t, string(expected), string(contents))
	})
}

func TestWriteCycloneSBOM(t *testing.T) {
	t.Parallel()

	var utilsMock piperutils.FileUtils
	utilsMock = &mock.FilesMock{}

	t.Run("success case", func(t *testing.T) {
		paths, err := WriteCycloneSBOM([]byte{1, 2, 3, 4}, utilsMock)
		assert.NoError(t, err, "unexpexted error")
		assert.Equal(t, 1, len(paths))
		assert.Equal(t, "whitesource/piper_whitesource_sbom.xml", paths[0].Target)

		exists, err := utilsMock.FileExists(paths[0].Target)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestCreateSarifResultFile(t *testing.T) {
	scan := &Scan{ProductVersion: "1"}
	scan.AppendScannedProject("project1")
	scan.AgentName = "Some test agent"
	scan.AgentVersion = "1.2.6"
	alerts := []Alert{
		{Library: Library{Filename: "vul1", ArtifactID: "org.some.lib"}, Vulnerability: Vulnerability{Name: "CVE-2022-001", CVSS3Score: 7.0, Score: 6}},
		{Library: Library{Filename: "vul2", ArtifactID: "org.some.lib"}, Vulnerability: Vulnerability{Name: "CVE-2022-002", CVSS3Score: 8.0, TopFix: Fix{Message: "this is the top fix"}}},
		{Library: Library{Filename: "vul3", ArtifactID: "org.some.lib2"}, Vulnerability: Vulnerability{Name: "CVE-2022-003", Score: 6}},
	}

	sarif := CreateSarifResultFile(scan, &alerts)

	assert.Equal(t, "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos02/schemas/sarif-schema-2.1.0.json", sarif.Schema)
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

func TestGetAuditInformation(t *testing.T) {
	tt := []struct {
		name     string
		alert    Alert
		expected *format.SarifProperties
	}{
		{
			name: "New not audited alert",
			alert: Alert{
				Status: "OPEN",
			},
			expected: &format.SarifProperties{
				Audited:               false,
				ToolAuditMessage:      "",
				UnifiedAuditState:     "new",
				AuditRequirement:      format.AUDIT_REQUIREMENT_GROUP_1_DESC,
				AuditRequirementIndex: format.AUDIT_REQUIREMENT_GROUP_1_INDEX,
			},
		},
		{
			name: "Audited alert",
			alert: Alert{
				Status:   "IGNORE",
				Comments: "Not relevant alert",
				Vulnerability: Vulnerability{
					CVSS3Score:    9.3,
					CVSS3Severity: "critical",
				},
			},
			expected: &format.SarifProperties{
				Audited:               true,
				ToolAuditMessage:      "Not relevant alert",
				UnifiedAuditState:     "notRelevant",
				UnifiedSeverity:       "critical",
				UnifiedCriticality:    9.3,
				AuditRequirement:      format.AUDIT_REQUIREMENT_GROUP_1_DESC,
				AuditRequirementIndex: format.AUDIT_REQUIREMENT_GROUP_1_INDEX,
			},
		},
		{
			name: "Alert with incorrect status",
			alert: Alert{
				Status:   "Not correct",
				Comments: "Some comment",
			},
			expected: &format.SarifProperties{
				Audited:               false,
				ToolAuditMessage:      "",
				UnifiedAuditState:     "new",
				AuditRequirement:      format.AUDIT_REQUIREMENT_GROUP_1_DESC,
				AuditRequirementIndex: format.AUDIT_REQUIREMENT_GROUP_1_INDEX,
			},
		},
		{
			name: "Not audited alert",
			alert: Alert{
				Assessment: &format.Assessment{
					Status:   format.NotRelevant,
					Analysis: format.FixedByDevTeam,
				},
				Status:   "OPEN",
				Comments: "New alert",
			},
			expected: &format.SarifProperties{
				Audited:               true,
				ToolAuditMessage:      string(format.FixedByDevTeam),
				UnifiedAuditState:     "notRelevant",
				AuditRequirement:      format.AUDIT_REQUIREMENT_GROUP_1_DESC,
				AuditRequirementIndex: format.AUDIT_REQUIREMENT_GROUP_1_INDEX,
			},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, getAuditInformation(test.alert))
		})
	}
}
