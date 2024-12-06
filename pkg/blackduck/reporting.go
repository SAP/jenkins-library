package blackduck

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/pkg/errors"
)

var severityIndex = map[string]int{"LOW": 1, "MEDIUM": 2, "HIGH": 3, "CRITICAL": 4}

// CreateSarifResultFile creates a SARIF result from the Vulnerabilities that were brought up by the scan
func CreateSarifResultFile(vulns *Vulnerabilities, projectName, projectVersion, projectLink string) *format.SARIF {
	log.Entry().Debug("Creating SARIF file for data transfer")

	// Handle results/vulnerabilities
	rules := []format.SarifRule{}
	collectedRules := []string{}
	cweIdsForTaxonomies := []string{}
	results := []format.Results{}

	if vulns != nil && vulns.Items != nil {
		for _, v := range vulns.Items {

			isAudited := true
			if v.RemediationStatus == "NEW" || v.RemediationStatus == "REMEDIATION_REQUIRED" ||
				v.RemediationStatus == "NEEDS_REVIEW" {
				isAudited = false
			}

			unifiedStatusValue := "new"

			switch v.RemediationStatus {
			case "NEW":
				unifiedStatusValue = "new"
			case "NEEDS_REVIEW":
				unifiedStatusValue = "inProcess"
			case "REMEDIATION_COMPLETE":
				unifiedStatusValue = "notRelevant"
			case "PATCHED":
				unifiedStatusValue = "notRelevant"
			case "MITIGATED":
				unifiedStatusValue = "notRelevant"
			case "DUPLICATE":
				unifiedStatusValue = "notRelevant"
			case "IGNORED":
				unifiedStatusValue = "notRelevant"
			case "REMEDIATION_REQUIRED":
				unifiedStatusValue = "relevant"
			}

			log.Entry().Debugf("Transforming alert %v on Package %v Version %v into SARIF format", v.VulnerabilityWithRemediation.VulnerabilityName, v.Component.Name, v.Component.Version)
			result := format.Results{
				RuleID:  v.VulnerabilityWithRemediation.VulnerabilityName,
				Level:   transformToLevel(v.VulnerabilityWithRemediation.Severity),
				Message: &format.Message{Text: v.VulnerabilityWithRemediation.Description},
				AnalysisTarget: &format.ArtifactLocation{
					URI:   v.Component.ToPackageUrl().ToString(),
					Index: 0,
				},
				Locations: []format.Location{{PhysicalLocation: format.PhysicalLocation{ArtifactLocation: format.ArtifactLocation{URI: v.Name}}}},
				PartialFingerprints: format.PartialFingerprints{
					PackageURLPlusCVEHash: base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%v+%v", v.Component.ToPackageUrl().ToString(), v.CweID))),
				},
				Properties: &format.SarifProperties{
					Audited:               isAudited,
					ToolSeverity:          v.Severity,
					ToolSeverityIndex:     severityIndex[v.Severity],
					ToolState:             v.RemediationStatus,
					ToolAuditMessage:      v.VulnerabilityWithRemediation.RemediationComment,
					UnifiedAuditState:     unifiedStatusValue,
					UnifiedSeverity:       strings.ToLower(v.Severity),
					UnifiedCriticality:    v.BaseScore,
					UnifiedAuditUser:      v.VulnerabilityWithRemediation.RemidiatedBy,
					AuditRequirement:      format.AUDIT_REQUIREMENT_GROUP_1_DESC,
					AuditRequirementIndex: format.AUDIT_REQUIREMENT_GROUP_1_INDEX,
				},
			}

			// append the result
			results = append(results, result)

			// append taxonomies
			if len(v.VulnerabilityWithRemediation.CweID) > 0 && !slices.Contains(cweIdsForTaxonomies, v.VulnerabilityWithRemediation.CweID) {
				cweIdsForTaxonomies = append(cweIdsForTaxonomies, v.VulnerabilityWithRemediation.CweID)
			}

			// only create rule on new CVE
			if !slices.Contains(collectedRules, result.RuleID) {
				collectedRules = append(collectedRules, result.RuleID)

				// set information about BlackDuck project
				v.projectVersionLink = projectLink
				v.projectName = projectName
				v.projectVersion = projectVersion

				markdown, _ := v.ToMarkdown()

				tags := []string{
					"SECURITY_VULNERABILITY",
					v.Component.ToPackageUrl().ToString(),
				}

				if CweID := v.VulnerabilityWithRemediation.CweID; CweID != "" {
					tags = append(tags, CweID)
				}

				if matchedType := v.Component.MatchedType(); matchedType != "" {
					tags = append(tags, matchedType)
				}

				ruleProp := format.SarifRuleProperties{
					Tags:             tags,
					Precision:        "very-high",
					Impact:           fmt.Sprint(v.VulnerabilityWithRemediation.ImpactSubscore),
					Probability:      fmt.Sprint(v.VulnerabilityWithRemediation.ExploitabilitySubscore),
					SecuritySeverity: fmt.Sprint(v.OverallScore),
				}
				sarifRule := format.SarifRule{
					ID:                   result.RuleID,
					ShortDescription:     &format.Message{Text: fmt.Sprintf("%v in Package %v", v.VulnerabilityName, v.Component.Name)},
					FullDescription:      &format.Message{Text: v.VulnerabilityWithRemediation.Description},
					DefaultConfiguration: &format.DefaultConfiguration{Level: transformToLevel(v.VulnerabilityWithRemediation.Severity)},
					HelpURI:              "",
					Help:                 &format.Help{Text: v.ToTxt(), Markdown: string(markdown)},
					Properties:           &ruleProp,
				}
				// append the rule
				rules = append(rules, sarifRule)
			}

		}
	}

	//handle taxonomies
	//Only one exists apparently: CWE. It is fixed
	taxas := []format.Taxa{}
	for _, value := range cweIdsForTaxonomies {
		taxa := format.Taxa{Id: value}
		taxas = append(taxas, taxa)
	}
	taxonomy := format.Taxonomies{
		GUID:             "25F72D7E-8A92-459D-AD67-64853F788765",
		Name:             "CWE",
		Organization:     "MITRE",
		ShortDescription: format.Message{Text: "The MITRE Common Weakness Enumeration"},
		Taxa:             taxas,
	}
	//handle the tool object
	tool := format.Tool{
		Driver: format.Driver{
			Name:           "Black Duck",
			Version:        "unknown",
			InformationUri: "https://community.synopsys.com/s/document-item?bundleId=integrations-detect&topicId=introduction.html&_LANG=enus",
			Rules:          rules,
		},
	}
	sarif := format.SARIF{
		Schema:  "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos02/schemas/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []format.Runs{
			{
				Results:             results,
				Tool:                tool,
				ThreadFlowLocations: []format.Locations{},
				Conversion: &format.Conversion{
					Tool: format.Tool{
						Driver: format.Driver{
							Name:           "Piper FPR to SARIF converter",
							InformationUri: "https://github.com/SAP/jenkins-library",
						},
					},
					Invocation: format.Invocation{
						ExecutionSuccessful: true,
						Properties:          &format.InvocationProperties{Platform: runtime.GOOS},
					},
				},
				Taxonomies: []format.Taxonomies{taxonomy},
			},
		},
	}

	return &sarif
}

func transformToLevel(severity string) string {
	switch severity {
	case "LOW":
		return "warning"
	case "MEDIUM":
		return "warning"
	case "HIGH":
		return "error"
	case "CRITICAL":
		return "error"
	}
	return "none"
}

// WriteVulnerabilityReports writes vulnerability information from ScanReport into dedicated outputs e.g. HTML
func WriteVulnerabilityReports(scanReport reporting.ScanReport, utils piperutils.FileUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := "piper_detect_vulnerability_report.html"
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write html report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "BlackDuck Vulnerability Report", Target: htmlReportPath})

	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("detectExecuteScan_oss_%v.json", fmt.Sprintf("%v", utils.CurrentTime("")))), jsonReport, 0666); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to write json report")
	}

	return reportPaths, nil
}

// WriteSarifFile write a JSON sarif format file for upload into e.g. GCP
func WriteSarifFile(sarif *format.SARIF, utils piperutils.FileUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	sarifReport, errorMarshall := json.Marshal(sarif)
	if errorMarshall != nil {
		return reportPaths, errors.Wrapf(errorMarshall, "failed to marshall SARIF json file")
	}
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to create report directory")
	}

	sarifReportPath := filepath.Join(ReportsDirectory, "piper_detect_vulnerability.sarif")
	if err := utils.FileWrite(sarifReportPath, sarifReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write SARIF file")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "Blackduck Detect Vulnerability SARIF file", Target: sarifReportPath})

	return reportPaths, nil
}
