package blackduck

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/package-url/packageurl-go"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/pkg/errors"
)

// CreateSarifResultFile creates a SARIF result from the Vulnerabilities that were brought up by the scan
func CreateSarifResultFile(vulns *Vulnerabilities) *format.SARIF {
	log.Entry().Debug("Creating SARIF file for data transfer")

	// Handle results/vulnerabilities
	rules := []format.SarifRule{}
	collectedRules := []string{}
	cweIdsForTaxonomies := []string{}
	results := []format.Results{}
	if vulns != nil && vulns.Items != nil {
		for _, v := range vulns.Items {
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
			}
			// append the result
			results = append(results, result)

			// append taxonomies
			if len(v.VulnerabilityWithRemediation.CweID) > 0 && !piperutils.ContainsString(cweIdsForTaxonomies, v.VulnerabilityWithRemediation.CweID) {
				cweIdsForTaxonomies = append(cweIdsForTaxonomies, v.VulnerabilityWithRemediation.CweID)
			}

			// only create rule on new CVE
			if !piperutils.ContainsString(collectedRules, result.RuleID) {
				collectedRules = append(collectedRules, result.RuleID)

				markdown, _ := v.ToMarkdown()
				tags := []string{
					"SECURITY_VULNERABILITY",
					v.Component.ToPackageUrl().ToString(),
					v.VulnerabilityWithRemediation.CweID,
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
			Name:           "Blackduck Hub Detect",
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
	switch strings.ToUpper(severity) {
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

func CreateCycloneSBOM(buildTool, groupID, artifactID, version, projectName, projectVersion string, libraries *HierarchicalComponents, alerts, assessedAlerts *Vulnerabilities) ([]byte, error) {
	componentLookup := map[string]HierarchicalComponent{}
	for _, comp := range libraries.Items {
		componentLookup[fmt.Sprintf("%v/%v", comp.Name, comp.Version)] = comp
	}

	ppurl := packageurl.NewPackageURL(format.TransformBuildToPurlType(buildTool), groupID, artifactID, version, nil, "")
	metadata := cdx.Metadata{
		// Define metadata about the main component
		// (the component which the BOM will describe)

		// TODO check whether we can identify library vs. application
		Component: &cdx.Component{
			BOMRef:     ppurl.ToString(),
			Type:       cdx.ComponentTypeLibrary,
			Name:       artifactID,
			Group:      groupID,
			Version:    version,
			PackageURL: ppurl.ToString(),
		},
		// Use properties to include an internal identifier for this BOM
		// https://cyclonedx.org/use-cases/#properties--name-value-store
		Properties: &[]cdx.Property{
			{
				Name:  "internal:bd-project-identifier",
				Value: projectName,
			},
			{
				Name:  "internal:bd-project-version-identifier",
				Value: projectVersion,
			},
		},
	}

	components := []cdx.Component{}
	uniqueComponents := []HierarchicalComponent{}
	transformToUniqueFlatList(&uniqueComponents, &componentLookup)
	log.Entry().Debugf("Got %v unique libraries in condensed flat list", len(uniqueComponents))
	sort.Slice(uniqueComponents, func(i, j int) bool {
		return uniqueComponents[i].ToPackageUrl().ToString() < uniqueComponents[j].ToPackageUrl().ToString()
	})
	for _, lib := range uniqueComponents {
		purl := lib.ToPackageUrl()
		// Define the components that the product ships with
		// https://cyclonedx.org/use-cases/#inventory
		component := cdx.Component{
			BOMRef:     purl.ToString(),
			Type:       cdx.ComponentTypeLibrary,
			Author:     transformComponentOriginToPurlParts(&lib)[1],
			Name:       lib.Name,
			Version:    lib.Version,
			PackageURL: purl.ToString(),
		}
		components = append(components, component)
	}

	dependencies := []cdx.Dependency{}
	declareDependency(ppurl, &libraries.Items, &dependencies)

	vulnerabilities := []cdx.Vulnerability{}
	transformAlertsToCdxVulnerabilities(alerts.Items)
	transformAlertsToCdxVulnerabilities(assessedAlerts.Items)

	// Assemble the BOM
	bom := cdx.NewBOM()
	bom.Vulnerabilities = &vulnerabilities
	bom.Metadata = &metadata
	bom.Components = &components
	bom.Dependencies = &dependencies

	// Encode the BOM
	var outputBytes []byte
	buffer := bytes.NewBuffer(outputBytes)
	encoder := cdx.NewBOMEncoder(buffer, cdx.BOMFileFormatXML)
	encoder.SetPretty(true)
	if err := encoder.Encode(bom); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func transformAlertsToCdxVulnerabilities(items []Vulnerability) []cdx.Vulnerability {
	vulnerabilities := []cdx.Vulnerability{}
	for _, alert := range items {
		// Define the vulnerabilities in VEX
		// https://cyclonedx.org/use-cases/#vulnerability-exploitability
		relatedComponent := alert.Component
		purl := relatedComponent.ToPackageUrl()
		cvss3Score := float64(alert.OverallScore)
		vuln := cdx.Vulnerability{
			BOMRef: purl.ToString(),
			ID:     alert.CweID,
			Source: &cdx.Source{URL: relatedComponent.Href},
			Tools: &[]cdx.Tool{
				{
					Name:    "Blackduck Hub Detect",
					Version: "Unknown",
					Vendor:  "Synopsis Inc.",
					ExternalReferences: &[]cdx.ExternalReference{
						{
							URL:  "https://www.blackducksoftware.com/",
							Type: cdx.ERTypeBuildMeta,
						},
					},
				},
			},
			Recommendation: alert.Description,
			Ratings: &[]cdx.VulnerabilityRating{
				{
					Score:    &cvss3Score,
					Severity: transformToCdxSeverity(alert.Severity),
					Method:   cdx.ScoringMethodCVSSv3,
				},
			},
			Description: alert.Description,
			Affects: &[]cdx.Affects{
				{
					Ref: purl.ToString(),
					Range: &[]cdx.AffectedVersions{
						{
							Version: relatedComponent.Version,
							Status:  cdx.VulnerabilityStatus(alert.RemediationStatus),
						},
					},
				},
			},
		}
		if alert.Assessment != nil {
			vuln.Analysis = &cdx.VulnerabilityAnalysis{
				State:         alert.Assessment.ToImpactAnalysisState(),
				Justification: alert.Assessment.ToImpactJustification(),
				Response:      alert.Assessment.ToImpactAnalysisResponse(),
			}
		}
		vulnerabilities = append(vulnerabilities, vuln)
	}
	return vulnerabilities
}

func transformToCdxSeverity(severity string) cdx.Severity {
	switch strings.ToLower(severity) {
	case "info":
		return cdx.SeverityInfo
	case "low":
		return cdx.SeverityLow
	case "medium":
		return cdx.SeverityMedium
	case "high":
		return cdx.SeverityHigh
	case "critical":
		return cdx.SeverityCritical
	case "":
		return cdx.SeverityNone
	}
	return cdx.SeverityUnknown
}

func WriteCycloneSBOM(sbom []byte, utils piperutils.FileUtils) ([]piperutils.Path, error) {
	paths := []piperutils.Path{}
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return paths, errors.Wrapf(err, "failed to create report directory")
	}

	sbomPath := filepath.Join(ReportsDirectory, "piper_hub_detect_sbom.xml")

	// Write file
	if err := utils.FileWrite(sbomPath, sbom, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return paths, errors.Wrapf(err, "failed to write BlackDuck SBOM file")
	}
	paths = append(paths, piperutils.Path{Name: "BlackDuck Hub Detect SBOM file", Target: sbomPath})

	return paths, nil
}

func transformToUniqueFlatList(libraries *[]HierarchicalComponent, flatMapRef *map[string]HierarchicalComponent) {
	log.Entry().Debugf("Got %v libraries reported", len(*libraries))
	for _, lib := range *libraries {
		key := lib.ToPackageUrl().ToString()
		flatMap := *flatMapRef
		lookup := flatMap[key]
		if lookup.ToPackageUrl() != lib.ToPackageUrl() {
			flatMap[key] = lib
		}
	}
}

func declareDependency(parentPurl *packageurl.PackageURL, dependents *[]HierarchicalComponent, collection *[]cdx.Dependency) {
	localDependencies := []cdx.Dependency{}
	for _, lib := range *dependents {
		purl := lib.ToPackageUrl()
		// Define the dependency graph
		// https://cyclonedx.org/use-cases/#dependency-graph
		localDependency := cdx.Dependency{Ref: purl.ToString()}
		localDependencies = append(localDependencies, localDependency)
	}
	dependency := cdx.Dependency{
		Ref:          parentPurl.ToString(),
		Dependencies: &localDependencies,
	}
	*collection = append(*collection, dependency)
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
