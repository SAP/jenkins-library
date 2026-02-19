package whitesource

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/package-url/packageurl-go"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
)

// CreateCustomVulnerabilityReport creates a vulnerability ScanReport to be used for uploading into various sinks
func CreateCustomVulnerabilityReport(productName string, scan *Scan, alerts *[]Alert, cvssSeverityLimit float64) reporting.ScanReport {
	severe, _ := CountSecurityVulnerabilities(alerts, cvssSeverityLimit)

	// sort according to vulnerability severity
	sort.Slice(*alerts, func(i, j int) bool {
		return vulnerabilityScore((*alerts)[i]) > vulnerabilityScore((*alerts)[j])
	})

	projectNames := scan.ScannedProjectNames()

	scanReport := reporting.ScanReport{
		ReportTitle: "WhiteSource Security Vulnerability Report",
		Subheaders: []reporting.Subheader{
			{Description: "WhiteSource product name", Details: productName},
			{Description: "Filtered project names", Details: strings.Join(projectNames, ", ")},
		},
		Overview: []reporting.OverviewRow{
			{Description: "Total number of vulnerabilities", Details: fmt.Sprint(len(*alerts))},
			{Description: "Total number of high/critical vulnerabilities with CVSS score >= 7.0", Details: fmt.Sprint(severe)},
		},
		SuccessfulScan: severe == 0,
		ReportTime:     time.Now(),
	}

	detailTable := reporting.ScanDetailTable{
		NoRowsMessage: "No publicly known vulnerabilities detected",
		Headers: []string{
			"Date",
			"CVE",
			"CVSS Score",
			"CVSS Version",
			"Project",
			"Library file name",
			"Library group ID",
			"Library artifact ID",
			"Library version",
			"Description",
			"Top fix",
		},
		WithCounter:   true,
		CounterHeader: "Entry #",
	}

	for _, alert := range *alerts {
		var score float64
		var scoreStyle reporting.ColumnStyle = reporting.Yellow
		if isSevereVulnerability(alert, cvssSeverityLimit) {
			scoreStyle = reporting.Red
		}
		var cveVersion string
		if alert.Vulnerability.CVSS3Score > 0 {
			score = alert.Vulnerability.CVSS3Score
			cveVersion = "v3"
		} else {
			score = alert.Vulnerability.Score
			cveVersion = "v2"
		}

		var topFix string
		emptyFix := Fix{}
		if alert.Vulnerability.TopFix != emptyFix {
			topFix = fmt.Sprintf(`%v<br>%v<br><a href="%v">%v</a>}"`, alert.Vulnerability.TopFix.Message, alert.Vulnerability.TopFix.FixResolution, alert.Vulnerability.TopFix.URL, alert.Vulnerability.TopFix.URL)
		}

		row := reporting.ScanRow{}
		row.AddColumn(alert.Vulnerability.PublishDate, 0)
		row.AddColumn(fmt.Sprintf(`<a href="%v">%v</a>`, alert.Vulnerability.URL, alert.Vulnerability.Name), 0)
		row.AddColumn(score, scoreStyle)
		row.AddColumn(cveVersion, 0)
		row.AddColumn(alert.Project, 0)
		row.AddColumn(alert.Library.Filename, 0)
		row.AddColumn(alert.Library.GroupID, 0)
		row.AddColumn(alert.Library.ArtifactID, 0)
		row.AddColumn(alert.Library.Version, 0)
		row.AddColumn(alert.Vulnerability.Description, 0)
		row.AddColumn(topFix, 0)

		detailTable.Rows = append(detailTable.Rows, row)
	}
	scanReport.DetailTable = detailTable

	return scanReport
}

// CountSecurityVulnerabilities counts the security vulnerabilities above severityLimit
func CountSecurityVulnerabilities(alerts *[]Alert, cvssSeverityLimit float64) (int, int) {
	severeVulnerabilities := 0
	for _, alert := range *alerts {
		if isSevereVulnerability(alert, cvssSeverityLimit) {
			severeVulnerabilities++
		}
	}

	nonSevereVulnerabilities := len(*alerts) - severeVulnerabilities
	return severeVulnerabilities, nonSevereVulnerabilities
}

func isSevereVulnerability(alert Alert, cvssSeverityLimit float64) bool {

	if vulnerabilityScore(alert) >= cvssSeverityLimit && cvssSeverityLimit >= 0 {
		return true
	}
	return false
}

func vulnerabilityScore(alert Alert) float64 {
	if alert.Vulnerability.CVSS3Score > 0 {
		return alert.Vulnerability.CVSS3Score
	}
	return alert.Vulnerability.Score
}

// ReportSha creates a SHA unique to the WS product and scan to be used as part of the report filename
func ReportSha(productName string, scan *Scan) string {
	reportShaData := []byte(productName + "," + strings.Join(scan.ScannedProjectNames(), ","))
	return fmt.Sprintf("%x", sha1.Sum(reportShaData))
}

// WriteCustomVulnerabilityReports creates an HTML and a JSON format file based on the alerts brought up by the scan
func WriteCustomVulnerabilityReports(productName string, scan *Scan, scanReport reporting.ScanReport, utils piperutils.FileUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	htmlReport, _ := scanReport.ToHTML()
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, fmt.Errorf("failed to create report directory: %w", err)
	}
	htmlReportPath := filepath.Join(ReportsDirectory, "piper_whitesource_vulnerability_report.html")
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, fmt.Errorf("failed to write html report: %w", err)
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "WhiteSource Vulnerability Report", Target: htmlReportPath})

	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, fmt.Errorf("failed to create step reporting directory: %w", err)
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("whitesourceExecuteScan_oss_%v.json", ReportSha(productName, scan))), jsonReport, 0666); err != nil {
		return reportPaths, fmt.Errorf("failed to write json report: %w", err)
	}
	// we do not add the json report to the overall list of reports for now,
	// since it is just an intermediary report used as input for later
	// and there does not seem to be real benefit in archiving it.

	return reportPaths, nil
}

// Creates a SARIF result from the Alerts that were brought up by the scan
func CreateSarifResultFile(scan *Scan, alerts *[]Alert) *format.SARIF {
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
	tool.Driver.Name = scan.AgentName
	tool.Driver.Version = scan.AgentVersion
	tool.Driver.InformationUri = "https://mend.io"

	// Handle results/vulnerabilities
	collectedRules := []string{}
	for _, alert := range *alerts {
		result := *new(format.Results)
		ruleId := alert.Vulnerability.Name
		log.Entry().Debugf("Transforming alert %v into SARIF format", ruleId)
		result.RuleID = ruleId
		result.Message = new(format.Message)
		result.Message.Text = alert.Vulnerability.Description
		artLoc := new(format.ArtifactLocation)
		artLoc.Index = 0
		artLoc.URI = alert.Library.Filename
		result.AnalysisTarget = artLoc
		location := format.Location{PhysicalLocation: format.PhysicalLocation{ArtifactLocation: format.ArtifactLocation{URI: alert.Library.Filename}}}
		result.Locations = append(result.Locations, location)
		partialFingerprints := new(format.PartialFingerprints)
		partialFingerprints.PackageURLPlusCVEHash = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%v+%v", alert.Library.ToPackageUrl().ToString(), alert.Vulnerability.Name)))
		result.PartialFingerprints = *partialFingerprints
		result.Properties = getAuditInformation(alert)

		//append the result
		sarif.Runs[0].Results = append(sarif.Runs[0].Results, result)

		// only create rule on new CVE
		if !slices.Contains(collectedRules, ruleId) {
			collectedRules = append(collectedRules, ruleId)

			sarifRule := *new(format.SarifRule)
			sarifRule.ID = ruleId
			sarifRule.Name = alert.Vulnerability.Name
			sd := new(format.Message)
			sd.Text = fmt.Sprintf("%v Package %v", alert.Vulnerability.Name, alert.Library.ArtifactID)
			sarifRule.ShortDescription = sd
			fd := new(format.Message)
			fd.Text = alert.Vulnerability.Description
			sarifRule.FullDescription = fd
			defaultConfig := new(format.DefaultConfiguration)
			defaultConfig.Level = transformToLevel(alert.Vulnerability.Severity, alert.Vulnerability.CVSS3Severity)
			sarifRule.DefaultConfiguration = defaultConfig
			sarifRule.HelpURI = alert.Vulnerability.URL
			markdown, _ := alert.ToMarkdown()
			sarifRule.Help = new(format.Help)
			sarifRule.Help.Text = alert.ToTxt()
			sarifRule.Help.Markdown = string(markdown)

			ruleProp := *new(format.SarifRuleProperties)
			ruleProp.Tags = append(ruleProp.Tags, alert.Type)
			ruleProp.Tags = append(ruleProp.Tags, alert.Library.ToPackageUrl().ToString())
			ruleProp.Tags = append(ruleProp.Tags, alert.Vulnerability.URL)
			ruleProp.SecuritySeverity = fmt.Sprint(consolidateScores(alert.Vulnerability.Score, alert.Vulnerability.CVSS3Score))
			ruleProp.Precision = "very-high"

			sarifRule.Properties = &ruleProp

			// append the rule
			tool.Driver.Rules = append(tool.Driver.Rules, sarifRule)
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

	return &sarif
}

func getAuditInformation(alert Alert) *format.SarifProperties {
	unifiedAuditState := "new"
	auditMessage := ""
	isAudited := false

	// unified audit state
	switch alert.Status {
	case "OPEN":
		unifiedAuditState = "new"
	case "IGNORE":
		unifiedAuditState = "notRelevant"
		auditMessage = alert.Comments
	}

	if alert.Assessment != nil {
		unifiedAuditState = string(alert.Assessment.Status)
		auditMessage = string(alert.Assessment.Analysis)
	}

	if unifiedAuditState == string(format.Relevant) ||
		unifiedAuditState == string(format.NotRelevant) {
		isAudited = true
	}

	return &format.SarifProperties{
		Audited:               isAudited,
		ToolAuditMessage:      auditMessage,
		UnifiedAuditState:     unifiedAuditState,
		AuditRequirement:      format.AUDIT_REQUIREMENT_GROUP_1_DESC,
		AuditRequirementIndex: format.AUDIT_REQUIREMENT_GROUP_1_INDEX,
		UnifiedSeverity:       alert.Vulnerability.CVSS3Severity,
		UnifiedCriticality:    float32(alert.Vulnerability.CVSS3Score),
	}
}

func transformToLevel(cvss2severity, cvss3severity string) string {
	cvssseverity := consolidateSeverities(cvss2severity, cvss3severity)
	switch cvssseverity {
	case "low":
		return "warning"
	case "medium":
		return "warning"
	case "high":
		return "error"
	case "critical":
		return "error"
	}
	return "none"
}

func consolidateSeverities(cvss2severity, cvss3severity string) string {
	if len(cvss3severity) > 0 {
		return cvss3severity
	}
	return cvss2severity
}

// WriteSarifFile write a JSON sarif format file for upload into e.g. GCP
func WriteSarifFile(sarif *format.SARIF, utils piperutils.FileUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	sarifReport, errorMarshall := json.Marshal(sarif)
	if errorMarshall != nil {
		return reportPaths, fmt.Errorf("failed to marshall SARIF json file: %w", errorMarshall)
	}
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return reportPaths, fmt.Errorf("failed to create report directory: %w", err)
	}
	sarifReportPath := filepath.Join(ReportsDirectory, "piper_whitesource_vulnerability.sarif")
	if err := utils.FileWrite(sarifReportPath, sarifReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, fmt.Errorf("failed to write SARIF file: %w", err)
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "WhiteSource Vulnerability SARIF file", Target: sarifReportPath})

	return reportPaths, nil
}

func transformToCdxSeverity(severity string) cdx.Severity {
	switch severity {
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

func transformBuildToPurlType(buildType string) string {
	switch buildType {
	case "maven":
		return packageurl.TypeMaven
	case "npm":
		return packageurl.TypeNPM
	case "docker":
		return packageurl.TypeDocker
	case "kaniko":
		return packageurl.TypeDocker
	case "golang":
		return packageurl.TypeGolang
	case "mta":
		return packageurl.TypeComposer
	}
	return packageurl.TypeGeneric
}

func CreateCycloneSBOM(scan *Scan, libraries *[]Library, alerts, assessedAlerts *[]Alert) ([]byte, error) {
	ppurl := packageurl.NewPackageURL(transformBuildToPurlType(scan.BuildTool), scan.Coordinates.GroupID, scan.Coordinates.ArtifactID, scan.Coordinates.Version, nil, "")
	metadata := cdx.Metadata{
		// Define metadata about the main component
		// (the component which the BOM will describe)

		// TODO check whether we can identify library vs. application
		Component: &cdx.Component{
			BOMRef:     ppurl.ToString(),
			Type:       cdx.ComponentTypeLibrary,
			Name:       scan.Coordinates.ArtifactID,
			Group:      scan.Coordinates.GroupID,
			Version:    scan.Coordinates.Version,
			PackageURL: ppurl.ToString(),
		},
		// Use properties to include an internal identifier for this BOM
		// https://cyclonedx.org/use-cases/#properties--name-value-store
		Properties: &[]cdx.Property{
			{
				Name:  "internal:ws-product-identifier",
				Value: scan.ProductToken,
			},
			{
				Name:  "internal:ws-project-identifier",
				Value: strings.Join(scan.ScannedProjectTokens(), ", "),
			},
		},
	}

	components := []cdx.Component{}
	flatUniqueLibrariesMap := map[string]Library{}
	transformToUniqueFlatList(libraries, &flatUniqueLibrariesMap, 1)
	flatUniqueLibraries := piperutils.Values(flatUniqueLibrariesMap)
	log.Entry().Debugf("Got %v unique libraries in condensed flat list", len(flatUniqueLibraries))
	sort.Slice(flatUniqueLibraries, func(i, j int) bool {
		return flatUniqueLibraries[i].ToPackageUrl().ToString() < flatUniqueLibraries[j].ToPackageUrl().ToString()
	})
	for _, lib := range flatUniqueLibraries {
		purl := lib.ToPackageUrl()
		// Define the components that the product ships with
		// https://cyclonedx.org/use-cases/#inventory
		component := cdx.Component{
			BOMRef:     purl.ToString(),
			Type:       cdx.ComponentTypeLibrary,
			Author:     lib.GroupID,
			Name:       lib.ArtifactID,
			Version:    lib.Version,
			PackageURL: purl.ToString(),
			Hashes:     &[]cdx.Hash{{Algorithm: cdx.HashAlgoSHA1, Value: lib.Sha1}},
		}
		components = append(components, component)
	}

	dependencies := []cdx.Dependency{}
	declareDependency(ppurl, libraries, &dependencies)

	// Encode vulnerabilities
	vulnerabilities := []cdx.Vulnerability{}
	vulnerabilities = append(vulnerabilities, transformAlertsToVulnerabilities(scan, alerts)...)
	vulnerabilities = append(vulnerabilities, transformAlertsToVulnerabilities(scan, assessedAlerts)...)

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

func transformAlertsToVulnerabilities(scan *Scan, alerts *[]Alert) []cdx.Vulnerability {
	vulnerabilities := []cdx.Vulnerability{}
	for _, alert := range *alerts {
		// Define the vulnerabilities in VEX
		// https://cyclonedx.org/use-cases/#vulnerability-exploitability
		purl := alert.Library.ToPackageUrl()
		advisories := []cdx.Advisory{}
		for _, fix := range alert.Vulnerability.AllFixes {
			advisory := cdx.Advisory{
				Title: fix.Message,
				URL:   alert.Vulnerability.TopFix.URL,
			}
			advisories = append(advisories, advisory)
		}
		cvss3Score := alert.Vulnerability.CVSS3Score
		cvssScore := alert.Vulnerability.Score
		vuln := cdx.Vulnerability{
			BOMRef: purl.ToString(),
			ID:     alert.Vulnerability.Name,
			Source: &cdx.Source{URL: alert.Vulnerability.URL},
			Tools: &[]cdx.Tool{
				{
					Name:    scan.AgentName,
					Version: scan.AgentVersion,
					Vendor:  "Mend",
					ExternalReferences: &[]cdx.ExternalReference{
						{
							URL:  "https://www.mend.io/",
							Type: cdx.ERTypeBuildMeta,
						},
					},
				},
			},
			Recommendation: alert.Vulnerability.FixResolutionText,
			Detail:         alert.Vulnerability.URL,
			Ratings: &[]cdx.VulnerabilityRating{
				{
					Score:    &cvss3Score,
					Severity: transformToCdxSeverity(alert.Vulnerability.CVSS3Severity),
					Method:   cdx.ScoringMethodCVSSv3,
				},
				{
					Score:    &cvssScore,
					Severity: transformToCdxSeverity(alert.Vulnerability.Severity),
					Method:   cdx.ScoringMethodCVSSv2,
				},
			},
			Advisories:  &advisories,
			Description: alert.Vulnerability.Description,
			Created:     alert.CreationDate,
			Published:   alert.Vulnerability.PublishDate,
			Updated:     alert.ModifiedDate,
			Affects: &[]cdx.Affects{
				{
					Ref: purl.ToString(),
					Range: &[]cdx.AffectedVersions{
						{
							Version: alert.Library.Version,
							Status:  cdx.VulnerabilityStatus(alert.Status),
						},
					},
				},
			},
		}
		references := []cdx.VulnerabilityReference{}
		for _, ref := range alert.Vulnerability.References {
			reference := cdx.VulnerabilityReference{
				Source: &cdx.Source{Name: ref.Homepage, URL: ref.URL},
				ID:     ref.GenericPackageIndex,
			}
			references = append(references, reference)
		}
		vuln.References = &references
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

func WriteCycloneSBOM(sbom []byte, utils piperutils.FileUtils) ([]piperutils.Path, error) {
	paths := []piperutils.Path{}
	if err := utils.MkdirAll(ReportsDirectory, 0777); err != nil {
		return paths, fmt.Errorf("failed to create report directory: %w", err)
	}

	sbomPath := filepath.Join(ReportsDirectory, "piper_whitesource_sbom.xml")

	// Write file
	if err := utils.FileWrite(sbomPath, sbom, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return paths, fmt.Errorf("failed to write SARIF file: %w", err)
	}
	paths = append(paths, piperutils.Path{Name: "WhiteSource SBOM file", Target: sbomPath})

	return paths, nil
}

func transformToUniqueFlatList(libraries *[]Library, flatMapRef *map[string]Library, level int) {
	log.Entry().Debugf("Got %v libraries reported on level %v", len(*libraries), level)
	for _, lib := range *libraries {
		key := lib.ToPackageUrl().ToString()
		flatMap := *flatMapRef
		lookup := flatMap[key]
		if lookup.KeyID != lib.KeyID {
			flatMap[key] = lib
			if len(lib.Dependencies) > 0 {
				transformToUniqueFlatList(&lib.Dependencies, flatMapRef, level+1)
			}

		}
	}
}

func declareDependency(parentPurl *packageurl.PackageURL, dependents *[]Library, collection *[]cdx.Dependency) {
	localDependencies := []cdx.Dependency{}
	for _, lib := range *dependents {
		purl := lib.ToPackageUrl()
		// Define the dependency graph
		// https://cyclonedx.org/use-cases/#dependency-graph
		localDependency := cdx.Dependency{Ref: purl.ToString()}
		localDependencies = append(localDependencies, localDependency)

		if len(lib.Dependencies) > 0 {
			declareDependency(purl, &lib.Dependencies, collection)
		}
	}
	dependency := cdx.Dependency{
		Ref:          parentPurl.ToString(),
		Dependencies: &localDependencies,
	}
	*collection = append(*collection, dependency)
}
