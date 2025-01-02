//go:build unit
// +build unit

package blackduck

import (
	"fmt"
	"path/filepath"
	"slices"
	"testing"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/stretchr/testify/assert"
)

func TestCreateSarifResultFile(t *testing.T) {
	vulnerabilities := []string{"CVE-1", "CVE-2", "CVE-3", "CVE-4"}
	affectedComponent := Component{Name: "test1", Version: "1.2.3", ComponentOriginName: "Maven", PrimaryLanguage: "Java"}
	otherAffectedComponent := Component{Name: "test2", Version: "1.2.8", ComponentOriginName: "Maven", PrimaryLanguage: "Java"}
	alerts := []Vulnerability{
		{
			Name:      "test1",
			Version:   "1.2.3",
			Component: &affectedComponent,
			VulnerabilityWithRemediation: VulnerabilityWithRemediation{
				CweID:             "CWE-45456543",
				VulnerabilityName: "CVE-1",
				Severity:          "CRITICAL",
				Description:       "Some vulnerability that can be exploited by peeling the glue off.",
				BaseScore:         9.8, OverallScore: 10,
				RemediationStatus:  "IGNORED",
				RemediationComment: "CWE-45456543 Auto-remediated: CWE-45456543 is related to CVE-1, but the CWE team has determined that this component version is not affected.",
				RemidiatedBy:       "technical_user",
			},
		},
		{
			Name:      "test1",
			Version:   "1.2.3",
			Component: &affectedComponent,
			VulnerabilityWithRemediation: VulnerabilityWithRemediation{
				CweID:              "CWE-45456542",
				VulnerabilityName:  "CVE-2",
				Severity:           "CRITICAL",
				Description:        "Some other vulnerability that can be exploited by filling the glass.",
				BaseScore:          9,
				OverallScore:       9,
				RemediationStatus:  "NEW",
				RemediationComment: "",
			},
		},
		{
			Name:      "test1",
			Version:   "1.2.3",
			Component: &affectedComponent,
			VulnerabilityWithRemediation: VulnerabilityWithRemediation{
				CweID:             "CWE-45456541",
				VulnerabilityName: "CVE-3",
				Severity:          "HIGH",
				Description:       "Some vulnerability that can be exploited by turning it upside down.",
				BaseScore:         6.5,
				OverallScore:      7,
				RemediationStatus: "IGNORED",
			},
		},
		{
			Name:      "test2",
			Version:   "1.2.8",
			Component: &otherAffectedComponent,
			VulnerabilityWithRemediation: VulnerabilityWithRemediation{
				CweID:             "CWE-45789754",
				VulnerabilityName: "CVE-4",
				Severity:          "HIGH",
				Description:       "Some vulnerability that can be exploited by turning it upside down.",
				BaseScore:         6.5,
				OverallScore:      7,
			},
		},
		{
			Name:      "test2",
			Version:   "1.2.8",
			Component: &otherAffectedComponent,
			VulnerabilityWithRemediation: VulnerabilityWithRemediation{
				CweID:             "CWE-45456541",
				VulnerabilityName: "CVE-3",
				Severity:          "HIGH",
				Description:       "Some vulnerability that can be exploited by turning it upside down.",
				BaseScore:         6.5,
				OverallScore:      7,
			},
		},
	}
	vulns := Vulnerabilities{
		Items: alerts,
	}
	projectName := "theProjectName"
	projectVersion := "theProjectVersion"
	projectLink := "theProjectLink"

	sarif := CreateSarifResultFile(&vulns, projectName, projectVersion, projectLink)

	assert.Equal(t, "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos02/schemas/sarif-schema-2.1.0.json", sarif.Schema)
	assert.Equal(t, "2.1.0", sarif.Version)
	assert.Equal(t, 1, len(sarif.Runs))
	assert.Equal(t, "Black Duck", sarif.Runs[0].Tool.Driver.Name)
	assert.Equal(t, "unknown", sarif.Runs[0].Tool.Driver.Version)
	assert.Equal(t, 4, len(sarif.Runs[0].Tool.Driver.Rules))
	assert.Equal(t, 5, len(sarif.Runs[0].Results))
	assert.Equal(t, "CRITICAL", sarif.Runs[0].Results[0].Properties.ToolSeverity)

	// Test correctness of audit information
	assert.Equal(t, true, sarif.Runs[0].Results[0].Properties.Audited)
	assert.Equal(t, "IGNORED", sarif.Runs[0].Results[0].Properties.ToolState)
	assert.Equal(t, alerts[0].BaseScore, sarif.Runs[0].Results[0].Properties.UnifiedCriticality)
	assert.Equal(t, "critical", sarif.Runs[0].Results[0].Properties.UnifiedSeverity)
	assert.Equal(t, "new", sarif.Runs[0].Results[1].Properties.UnifiedAuditState)
	assert.Equal(t, "notRelevant", sarif.Runs[0].Results[0].Properties.UnifiedAuditState)
	assert.Equal(t, "technical_user", sarif.Runs[0].Results[0].Properties.UnifiedAuditUser)
	assert.Equal(t, format.AUDIT_REQUIREMENT_GROUP_1_DESC, sarif.Runs[0].Results[0].Properties.AuditRequirement)
	assert.Equal(t, format.AUDIT_REQUIREMENT_GROUP_1_INDEX, sarif.Runs[0].Results[0].Properties.AuditRequirementIndex)
	assert.Equal(t,
		"CWE-45456543 Auto-remediated: CWE-45456543 is related to CVE-1, but the CWE team has determined that this component version is not affected.",
		sarif.Runs[0].Results[0].Properties.ToolAuditMessage,
	)
	assert.Equal(t, 4, sarif.Runs[0].Results[0].Properties.ToolSeverityIndex)
	assert.Equal(t, 3, sarif.Runs[0].Results[2].Properties.ToolSeverityIndex)

	assert.Equal(t, false, sarif.Runs[0].Results[1].Properties.Audited)
	assert.Equal(t, "NEW", sarif.Runs[0].Results[1].Properties.ToolState)
	assert.Equal(t, "", sarif.Runs[0].Results[1].Properties.ToolAuditMessage)

	collectedRules := []string{}
	for _, rule := range sarif.Runs[0].Tool.Driver.Rules {
		slices.Contains(vulnerabilities, rule.ID)
		collectedRules = append(collectedRules, rule.ID)
	}

	collectedResults := []string{}
	for _, result := range sarif.Runs[0].Results {
		slices.Contains(vulnerabilities, result.RuleID)
		collectedResults = append(collectedResults, result.RuleID)
	}

	assert.Equal(t, 4, len(collectedRules))
	assert.Equal(t, 5, len(collectedResults))
	assert.Equal(t, vulnerabilities, collectedRules)
}

func TestWriteCustomVulnerabilityReports(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		scanReport := reporting.ScanReport{}
		utilsMock := &mock.FilesMock{}

		reportPaths, err := WriteVulnerabilityReports(scanReport, utilsMock)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(reportPaths))

		exists, err := utilsMock.FileExists(reportPaths[0].Target)
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = utilsMock.FileExists(filepath.Join(reporting.StepReportDirectory, "detectExecuteScan_oss_20220102-150405.json"))
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("failed to write json report", func(t *testing.T) {
		scanReport := reporting.ScanReport{}
		utilsMock := &mock.FilesMock{}
		utilsMock.FileWriteErrors = map[string]error{
			filepath.Join(reporting.StepReportDirectory, "detectExecuteScan_oss_20220102-150405.json"): fmt.Errorf("write error"),
		}

		_, err := WriteVulnerabilityReports(scanReport, utilsMock)
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

		exists, err = utilsMock.FileExists(filepath.Join(ReportsDirectory, "piper_detect_vulnerability.sarif"))
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("failed to write HTML report", func(t *testing.T) {
		sarif := format.SARIF{}
		utilsMock := &mock.FilesMock{}
		utilsMock.FileWriteErrors = map[string]error{
			filepath.Join(ReportsDirectory, "piper_detect_vulnerability.sarif"): fmt.Errorf("write error"),
		}

		_, err := WriteSarifFile(&sarif, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to write SARIF file")
	})
}
