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
func CreateSarifResultFile(vulns *Vulnerabilities, components *Components) *format.SARIF {
	// create component lookup map
	componentLookup := map[string]Component{}
	for _, comp := range components.Items {
		componentLookup[fmt.Sprintf("%v/%v", comp.Name, comp.Version)] = comp
	}

	//Now, we handle the sarif
	log.Entry().Debug("Creating SARIF file for data transfer")
	var sarif format.SARIF
	sarif.Schema = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos02/schemas/sarif-schema-2.1.0.json"
	sarif.Version = "2.1.0"
	var wsRun format.Runs
	sarif.Runs = append(sarif.Runs, wsRun)

	//handle the tool object
	tool := *new(format.Tool)
	tool.Driver = *new(format.Driver)
	tool.Driver.Name = "Blackduck Hub Detect"
	tool.Driver.Version = "unknown"
	tool.Driver.InformationUri = "https://community.synopsys.com/s/document-item?bundleId=integrations-detect&topicId=introduction.html&_LANG=enus"

	// Handle results/vulnerabilities
	collectedRules := []string{}
	cweIdsForTaxonomies := []string{}
	if vulns != nil && vulns.Items != nil {
		for _, v := range vulns.Items {
			component := componentLookup[fmt.Sprintf("%v/%v", v.Name, v.Version)]
			result := *new(format.Results)
			ruleId := v.Title()
			log.Entry().Debugf("Transforming alert %v into SARIF format", ruleId)
			result.RuleID = ruleId
			result.Level = transformToLevel(v.VulnerabilityWithRemediation.Severity)
			result.Message = new(format.Message)
			result.Message.Text = v.VulnerabilityWithRemediation.Description
			result.AnalysisTarget = new(format.ArtifactLocation)
			result.AnalysisTarget.URI = v.Name
			result.AnalysisTarget.Index = 0
			location := format.Location{PhysicalLocation: format.PhysicalLocation{ArtifactLocation: format.ArtifactLocation{URI: v.Name}}}
			result.Locations = append(result.Locations, location)
			partialFingerprints := new(format.PartialFingerprints)
			partialFingerprints.PackageURLPlusCVEHash = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%v+%v", component.ToPackageUrl().ToString(), v.Title())))
			result.PartialFingerprints = *partialFingerprints
			cweIdsForTaxonomies = append(cweIdsForTaxonomies, v.VulnerabilityWithRemediation.CweID)

			// append the result
			sarif.Runs[0].Results = append(sarif.Runs[0].Results, result)

			// only create rule on new CVE
			if !piperutils.ContainsString(collectedRules, ruleId) {
				collectedRules = append(collectedRules, ruleId)

				sarifRule := *new(format.SarifRule)
				sarifRule.ID = ruleId
				sarifRule.ShortDescription = new(format.Message)
				sarifRule.ShortDescription.Text = fmt.Sprintf("%v Package %v", v.VulnerabilityName, component.Name)
				sarifRule.FullDescription = new(format.Message)
				sarifRule.FullDescription.Text = v.VulnerabilityWithRemediation.Description
				sarifRule.DefaultConfiguration = new(format.DefaultConfiguration)
				sarifRule.DefaultConfiguration.Level = transformToLevel(v.VulnerabilityWithRemediation.Severity)
				sarifRule.HelpURI = ""
				markdown, _ := v.ToMarkdown(&component)
				sarifRule.Help = new(format.Help)
				sarifRule.Help.Text = v.ToTxt(&component)
				sarifRule.Help.Markdown = string(markdown)

				ruleProp := *new(format.SarifRuleProperties)
				ruleProp.Tags = append(ruleProp.Tags, "SECURITY_VULNERABILITY")
				ruleProp.Tags = append(ruleProp.Tags, component.ToPackageUrl().ToString())
				ruleProp.Tags = append(ruleProp.Tags, v.VulnerabilityWithRemediation.CweID)
				ruleProp.Precision = "very-high"
				ruleProp.Impact = fmt.Sprint(v.VulnerabilityWithRemediation.ImpactSubscore)
				ruleProp.Probability = fmt.Sprint(v.VulnerabilityWithRemediation.ExploitabilitySubscore)
				ruleProp.SecuritySeverity = fmt.Sprint(v.OverallScore)
				sarifRule.Properties = &ruleProp

				// append the rule
				tool.Driver.Rules = append(tool.Driver.Rules, sarifRule)
			}
		}
	}
	//Finalize: tool
	sarif.Runs[0].Tool = tool

	// Threadflowlocations is no loger useful: voiding it will make for smaller reports
	sarif.Runs[0].ThreadFlowLocations = []format.Locations{}

	// Add a conversion object to highlight this isn't native SARIF
	conversion := new(format.Conversion)
	conversion.Tool.Driver.Name = "Piper FPR to SARIF converter"
	conversion.Tool.Driver.InformationUri = "https://github.com/SAP/jenkins-library"
	conversion.Invocation.ExecutionSuccessful = true
	convInvocProp := new(format.InvocationProperties)
	convInvocProp.Platform = runtime.GOOS
	conversion.Invocation.Properties = convInvocProp
	sarif.Runs[0].Conversion = conversion

	//handle taxonomies
	//Only one exists apparently: CWE. It is fixed
	taxonomy := *new(format.Taxonomies)
	taxonomy.GUID = "25F72D7E-8A92-459D-AD67-64853F788765"
	taxonomy.Name = "CWE"
	taxonomy.Organization = "MITRE"
	taxonomy.ShortDescription.Text = "The MITRE Common Weakness Enumeration"
	for key := range cweIdsForTaxonomies {
		taxa := *new(format.Taxa)
		taxa.Id = fmt.Sprint(key)
		taxonomy.Taxa = append(taxonomy.Taxa, taxa)
	}
	sarif.Runs[0].Taxonomies = append(sarif.Runs[0].Taxonomies, taxonomy)

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

func CreateCycloneSBOM(buildTool, groupID, artifactID, version, projectName, projectVersion string, libraries *Components, alerts *Vulnerabilities) ([]byte, error) {
	componentLookup := map[string]Component{}
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
	uniqueComponents := []Component{}
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
			Author:     lib.ComponentOriginName,
			Name:       lib.Name,
			Version:    lib.Version,
			PackageURL: purl.ToString(),
		}
		components = append(components, component)
	}

	dependencies := []cdx.Dependency{}
	declareDependency(ppurl, &libraries.Items, &dependencies)

	vulnerabilities := []cdx.Vulnerability{}
	for _, alert := range *&alerts.Items {
		// Define the vulnerabilities in VEX
		// https://cyclonedx.org/use-cases/#vulnerability-exploitability
		relatedComponent := componentLookup[fmt.Sprintf("%v/%v", alert.Name, alert.Version)]
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
					Vendor:  "Mend",
					ExternalReferences: &[]cdx.ExternalReference{
						{
							URL:  "https://www.mend.io/",
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
		vulnerabilities = append(vulnerabilities, vuln)
	}

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

func transformToUniqueFlatList(libraries *[]Component, flatMapRef *map[string]Component) {
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

func declareDependency(parentPurl *packageurl.PackageURL, dependents *[]Component, collection *[]cdx.Dependency) {
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
