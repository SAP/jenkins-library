package cmd

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	piperDocker "github.com/SAP/jenkins-library/pkg/docker"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	ws "github.com/SAP/jenkins-library/pkg/whitesource"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

// ScanOptions is just used to make the lines less long
type ScanOptions = whitesourceExecuteScanOptions

// WhiteSource defines the functions that are expected by the step implementation to
// be available from the WhiteSource system.
type whitesource interface {
	GetProductByName(productName string) (ws.Product, error)
	CreateProduct(productName string) (string, error)
	SetProductAssignments(productToken string, membership, admins, alertReceivers *ws.Assignment) error
	GetProjectsMetaInfo(productToken string) ([]ws.Project, error)
	GetProjectToken(productToken, projectName string) (string, error)
	GetProjectByToken(projectToken string) (ws.Project, error)
	GetProjectRiskReport(projectToken string) ([]byte, error)
	GetProjectVulnerabilityReport(projectToken string, format string) ([]byte, error)
	GetProjectAlerts(projectToken string) ([]ws.Alert, error)
	GetProjectAlertsByType(projectToken, alertType string) ([]ws.Alert, error)
	GetProjectLibraryLocations(projectToken string) ([]ws.Library, error)
}

type whitesourceUtils interface {
	ws.Utils
	DirExists(path string) (bool, error)
	GetArtifactCoordinates(buildTool, buildDescriptorFile string,
		options *versioning.Options) (versioning.Coordinates, error)

	Now() time.Time
}

type whitesourceUtilsBundle struct {
	*piperhttp.Client
	*command.Command
	*piperutils.Files
	npmExecutor npm.Executor
}

func (w *whitesourceUtilsBundle) FileOpen(name string, flag int, perm os.FileMode) (ws.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (w *whitesourceUtilsBundle) GetArtifactCoordinates(buildTool, buildDescriptorFile string, options *versioning.Options) (versioning.Coordinates, error) {
	artifact, err := versioning.GetArtifact(buildTool, buildDescriptorFile, options, w)
	if err != nil {
		return versioning.Coordinates{}, err
	}
	return artifact.GetCoordinates()
}

func (w *whitesourceUtilsBundle) getNpmExecutor(config *ws.ScanOptions) npm.Executor {
	if w.npmExecutor == nil {
		w.npmExecutor = npm.NewExecutor(npm.ExecutorOptions{DefaultNpmRegistry: config.DefaultNpmRegistry})
	}
	return w.npmExecutor
}

func (w *whitesourceUtilsBundle) FindPackageJSONFiles(config *ws.ScanOptions) ([]string, error) {
	return w.getNpmExecutor(config).FindPackageJSONFilesWithExcludes(config.BuildDescriptorExcludeList)
}

func (w *whitesourceUtilsBundle) InstallAllNPMDependencies(config *ws.ScanOptions, packageJSONFiles []string) error {
	return w.getNpmExecutor(config).InstallAllDependencies(packageJSONFiles)
}

func (w *whitesourceUtilsBundle) SetOptions(o piperhttp.ClientOptions) {
	w.Client.SetOptions(o)
}

func (w *whitesourceUtilsBundle) Now() time.Time {
	return time.Now()
}

func newWhitesourceUtils(config *ScanOptions) *whitesourceUtilsBundle {
	utils := whitesourceUtilsBundle{
		Client:  &piperhttp.Client{},
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute cmd output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	// Configure HTTP Client
	utils.SetOptions(piperhttp.ClientOptions{TransportTimeout: time.Duration(config.Timeout) * time.Second})
	return &utils
}

func newWhitesourceScan(config *ScanOptions) *ws.Scan {
	return &ws.Scan{
		AggregateProjectName: config.ProjectName,
		ProductVersion:       config.Version,
	}
}

func whitesourceExecuteScan(config ScanOptions, _ *telemetry.CustomData, commonPipelineEnvironment *whitesourceExecuteScanCommonPipelineEnvironment, influx *whitesourceExecuteScanInflux) {
	utils := newWhitesourceUtils(&config)
	scan := newWhitesourceScan(&config)
	sys := ws.NewSystem(config.ServiceURL, config.OrgToken, config.UserToken, time.Duration(config.Timeout)*time.Second)
	influx.step_data.fields.whitesource = false
	err := runWhitesourceExecuteScan(&config, scan, utils, sys, commonPipelineEnvironment, influx)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	influx.step_data.fields.whitesource = true
}

func runWhitesourceExecuteScan(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils, sys whitesource, commonPipelineEnvironment *whitesourceExecuteScanCommonPipelineEnvironment, influx *whitesourceExecuteScanInflux) error {
	if err := resolveAggregateProjectName(config, scan, sys); err != nil {
		return errors.Wrapf(err, "failed to resolve and aggregate project name")
	}

	if err := resolveProjectIdentifiers(config, scan, utils, sys); err != nil {
		if strings.Contains(fmt.Sprint(err), "User is not allowed to perform this action") {
			log.SetErrorCategory(log.ErrorConfiguration)
		}
		return errors.Wrapf(err, "failed to resolve project identifiers")
	}

	if config.AggregateVersionWideReport {
		// Generate a vulnerability report for all projects with version = config.ProjectVersion
		// Note that this is not guaranteed that all projects are from the same scan.
		// For example, if a module was removed from the source code, the project may still
		// exist in the WhiteSource system.
		if err := aggregateVersionWideLibraries(config, utils, sys); err != nil {
			return errors.Wrapf(err, "failed to aggregate version wide libraries")
		}
		if err := aggregateVersionWideVulnerabilities(config, utils, sys); err != nil {
			return errors.Wrapf(err, "failed to aggregate version wide vulnerabilities")
		}
	} else {
		if err := runWhitesourceScan(config, scan, utils, sys, commonPipelineEnvironment, influx); err != nil {
			return errors.Wrapf(err, "failed to execute WhiteSource scan")
		}
	}
	return nil
}

func runWhitesourceScan(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils, sys whitesource, commonPipelineEnvironment *whitesourceExecuteScanCommonPipelineEnvironment, influx *whitesourceExecuteScanInflux) error {
	// Download Docker image for container scan
	// ToDo: move it to improve testability
	if config.BuildTool == "docker" {
		saveImageOptions := containerSaveImageOptions{
			ContainerImage:       config.ScanImage,
			ContainerRegistryURL: config.ScanImageRegistryURL,
			IncludeLayers:        config.ScanImageIncludeLayers,
		}
		dClientOptions := piperDocker.ClientOptions{ImageName: saveImageOptions.ContainerImage, RegistryURL: saveImageOptions.ContainerRegistryURL, LocalPath: "", IncludeLayers: saveImageOptions.IncludeLayers}
		dClient := &piperDocker.Client{}
		dClient.SetOptions(dClientOptions)
		if err := runContainerSaveImage(&saveImageOptions, &telemetry.CustomData{}, "./cache", "", dClient); err != nil {
			if strings.Contains(fmt.Sprint(err), "no image found") {
				log.SetErrorCategory(log.ErrorConfiguration)
			}
			return errors.Wrapf(err, "failed to dowload Docker image %v", config.ScanImage)
		}

	}

	// Start the scan
	if err := executeScan(config, scan, utils); err != nil {
		return errors.Wrapf(err, "failed to execute Scan")
	}

	// ToDo: Check this:
	// Why is this required at all, resolveProjectIdentifiers() is already called before the scan in runWhitesourceExecuteScan()
	// Could perhaps use scan.updateProjects(sys) directly... have not investigated what could break
	if err := resolveProjectIdentifiers(config, scan, utils, sys); err != nil {
		return errors.Wrapf(err, "failed to resolve project identifiers")
	}

	log.Entry().Info("-----------------------------------------------------")
	log.Entry().Infof("Product Version: '%s'", config.Version)
	log.Entry().Info("Scanned projects:")
	for _, project := range scan.ScannedProjects() {
		log.Entry().Infof("  Name: '%s', token: %s", project.Name, project.Token)
	}
	log.Entry().Info("-----------------------------------------------------")

	paths, err := checkAndReportScanResults(config, scan, utils, sys, influx)
	piperutils.PersistReportsAndLinks("whitesourceExecuteScan", "", paths, nil)
	persistScannedProjects(config, scan, commonPipelineEnvironment)
	if err != nil {
		return errors.Wrapf(err, "failed to check and report scan results")
	}
	return nil
}

func checkAndReportScanResults(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils, sys whitesource, influx *whitesourceExecuteScanInflux) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}
	if !config.Reporting && !config.SecurityVulnerabilities {
		return reportPaths, nil
	}
	// Wait for WhiteSource backend to propagate the changes before downloading any reports.
	if err := scan.BlockUntilReportsAreReady(sys); err != nil {
		return reportPaths, err
	}

	if config.Reporting {
		var err error
		reportPaths, err = scan.DownloadReports(ws.ReportOptions{
			ReportDirectory:           ws.ReportsDirectory,
			VulnerabilityReportFormat: config.VulnerabilityReportFormat,
		}, utils, sys)
		if err != nil {
			return reportPaths, err
		}
	}

	checkErrors := []string{}

	rPath, err := checkPolicyViolations(config, scan, sys, utils, reportPaths, influx)
	if err != nil {
		checkErrors = append(checkErrors, fmt.Sprint(err))
	}
	reportPaths = append(reportPaths, rPath)

	if config.SecurityVulnerabilities {
		rPaths, err := checkSecurityViolations(config, scan, sys, utils, influx)
		reportPaths = append(reportPaths, rPaths...)
		if err != nil {
			checkErrors = append(checkErrors, fmt.Sprint(err))
		}
	}

	// create toolrecord file
	// tbd - how to handle verifyOnly
	toolRecordFileName, err := createToolRecordWhitesource("./", config, scan)
	if err != nil {
		// do not fail until the framework is well established
		log.Entry().Warning("TR_WHITESOURCE: Failed to create toolrecord file ...", err)
	} else {
		reportPaths = append(reportPaths, piperutils.Path{Target: toolRecordFileName})
	}

	if len(checkErrors) > 0 {
		return reportPaths, fmt.Errorf(strings.Join(checkErrors, ": "))
	}
	return reportPaths, nil
}

func createWhiteSourceProduct(config *ScanOptions, sys whitesource) (string, error) {
	log.Entry().Infof("Attempting to create new WhiteSource product for '%s'..", config.ProductName)
	productToken, err := sys.CreateProduct(config.ProductName)
	if err != nil {
		return "", fmt.Errorf("failed to create WhiteSource product: %w", err)
	}

	var admins ws.Assignment
	for _, address := range config.EmailAddressesOfInitialProductAdmins {
		admins.UserAssignments = append(admins.UserAssignments, ws.UserAssignment{Email: address})
	}

	err = sys.SetProductAssignments(productToken, nil, &admins, nil)
	if err != nil {
		return "", fmt.Errorf("failed to set admins on new WhiteSource product: %w", err)
	}

	return productToken, nil
}

func resolveProjectIdentifiers(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils, sys whitesource) error {
	if len(scan.AggregateProjectName) > 0 && (len(config.Version)+len(config.CustomScanVersion) > 0) {
		if config.Version == "" {
			config.Version = config.CustomScanVersion
		}
	} else {
		options := &versioning.Options{
			DockerImage:         config.ScanImage,
			ProjectSettingsFile: config.ProjectSettingsFile,
			GlobalSettingsFile:  config.GlobalSettingsFile,
			M2Path:              config.M2Path,
		}
		coordinates, err := utils.GetArtifactCoordinates(config.BuildTool, config.BuildDescriptorFile, options)
		if err != nil {
			return errors.Wrap(err, "failed to get build artifact description")
		}

		if len(config.Version) > 0 {
			log.Entry().Infof("Resolving product version from default provided '%s' with versioning '%s'", config.Version, config.VersioningModel)
			coordinates.Version = config.Version
		}

		nameTmpl := `{{list .GroupID .ArtifactID | join "-" | trimAll "-"}}`
		name, version := versioning.DetermineProjectCoordinatesWithCustomVersion(nameTmpl, config.VersioningModel, config.CustomScanVersion, coordinates)
		if scan.AggregateProjectName == "" {
			log.Entry().Infof("Resolved project name '%s' from descriptor file", name)
			scan.AggregateProjectName = name
		}

		config.Version = version
		log.Entry().Infof("Resolved product version '%s'", version)
	}

	scan.ProductVersion = validateProductVersion(config.Version)

	if err := resolveProductToken(config, sys); err != nil {
		return errors.Wrap(err, "error resolving product token")
	}
	if err := resolveAggregateProjectToken(config, sys); err != nil {
		return errors.Wrap(err, "error resolving aggregate project token")
	}

	return scan.UpdateProjects(config.ProductToken, sys)
}

// resolveProductToken resolves the token of the WhiteSource Product specified by config.ProductName,
// unless the user provided a token in config.ProductToken already, or it was previously resolved.
// If no Product can be found for the given config.ProductName, and the parameter
// config.CreatePipelineFromProduct is set, an attempt will be made to create the product and
// configure the initial product admins.
func resolveProductToken(config *ScanOptions, sys whitesource) error {
	if config.ProductToken != "" {
		return nil
	}
	log.Entry().Infof("Attempting to resolve product token for product '%s'..", config.ProductName)
	product, err := sys.GetProductByName(config.ProductName)
	if err != nil && config.CreateProductFromPipeline {
		product = ws.Product{}
		product.Token, err = createWhiteSourceProduct(config, sys)
		if err != nil {
			return errors.Wrapf(err, "failed to create whitesource product")
		}
	}
	if err != nil {
		return errors.Wrapf(err, "failed to get product by name")
	}
	log.Entry().Infof("Resolved product token: '%s'..", product.Token)
	config.ProductToken = product.Token
	return nil
}

// resolveAggregateProjectName checks if config.ProjectToken is configured, and if so, expects a WhiteSource
// project with that token to exist. The AggregateProjectName in the ws.Scan is then configured with that
// project's name.
func resolveAggregateProjectName(config *ScanOptions, scan *ws.Scan, sys whitesource) error {
	if config.ProjectToken == "" {
		return nil
	}
	log.Entry().Infof("Attempting to resolve aggregate project name for token '%s'..", config.ProjectToken)
	// If the user configured the "projectToken" parameter, we expect this project to exist in the backend.
	project, err := sys.GetProjectByToken(config.ProjectToken)
	if err != nil {
		return errors.Wrapf(err, "failed to get project by token")
	}
	nameVersion := strings.Split(project.Name, " - ")
	scan.AggregateProjectName = nameVersion[0]
	log.Entry().Infof("Resolve aggregate project name '%s'..", scan.AggregateProjectName)
	return nil
}

// resolveAggregateProjectToken fetches the token of the WhiteSource Project specified by config.ProjectName
// and stores it in config.ProjectToken.
// The user can configure a projectName or projectToken of the project to be used as for aggregation of scan results.
func resolveAggregateProjectToken(config *ScanOptions, sys whitesource) error {
	if config.ProjectToken != "" || config.ProjectName == "" {
		return nil
	}
	log.Entry().Infof("Attempting to resolve project token for project '%s'..", config.ProjectName)
	fullProjName := fmt.Sprintf("%s - %s", config.ProjectName, config.Version)
	projectToken, err := sys.GetProjectToken(config.ProductToken, fullProjName)
	if err != nil {
		return errors.Wrapf(err, "failed to get project token")
	}
	// A project may not yet exist for this project name-version combo.
	// It will be created by the scan, we retrieve the token again after scanning.
	if projectToken != "" {
		log.Entry().Infof("Resolved project token: '%s'..", projectToken)
		config.ProjectToken = projectToken
	} else {
		log.Entry().Infof("Project '%s' not yet present in WhiteSource", fullProjName)
	}
	return nil
}

// validateProductVersion makes sure that the version does not contain a dash "-".
func validateProductVersion(version string) string {
	// TrimLeft() removes all "-" from the beginning, unlike TrimPrefix()!
	version = strings.TrimLeft(version, "-")
	if strings.Contains(version, "-") {
		version = strings.SplitN(version, "-", 1)[0]
	}
	return version
}

func wsScanOptions(config *ScanOptions) *ws.ScanOptions {
	return &ws.ScanOptions{
		BuildTool:                  config.BuildTool,
		ScanType:                   "", // no longer provided via config
		OrgToken:                   config.OrgToken,
		UserToken:                  config.UserToken,
		ProductName:                config.ProductName,
		ProductToken:               config.ProductToken,
		ProductVersion:             config.Version,
		ProjectName:                config.ProjectName,
		BuildDescriptorFile:        config.BuildDescriptorFile,
		BuildDescriptorExcludeList: config.BuildDescriptorExcludeList,
		PomPath:                    config.BuildDescriptorFile,
		M2Path:                     config.M2Path,
		GlobalSettingsFile:         config.GlobalSettingsFile,
		ProjectSettingsFile:        config.ProjectSettingsFile,
		InstallArtifacts:           config.InstallArtifacts,
		DefaultNpmRegistry:         config.DefaultNpmRegistry,
		AgentDownloadURL:           config.AgentDownloadURL,
		AgentFileName:              config.AgentFileName,
		ConfigFilePath:             config.ConfigFilePath,
		Includes:                   config.Includes,
		Excludes:                   config.Excludes,
		JreDownloadURL:             config.JreDownloadURL,
		AgentURL:                   config.AgentURL,
		ServiceURL:                 config.ServiceURL,
		ScanPath:                   config.ScanPath,
		Verbose:                    GeneralConfig.Verbose,
	}
}

// Unified Agent is the only supported option by WhiteSource going forward:
// The Unified Agent will be used to perform the scan.
func executeScan(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils) error {

	options := wsScanOptions(config)

	// Execute scan with Unified Agent jar file
	if err := scan.ExecuteUAScan(options, utils); err != nil {
		return errors.Wrapf(err, "failed to execute Unified Agent scan")
	}
	return nil
}

func checkPolicyViolations(config *ScanOptions, scan *ws.Scan, sys whitesource, utils whitesourceUtils, reportPaths []piperutils.Path, influx *whitesourceExecuteScanInflux) (piperutils.Path, error) {

	policyViolationCount := 0
	for _, project := range scan.ScannedProjects() {
		alerts, err := sys.GetProjectAlertsByType(project.Token, "REJECTED_BY_POLICY_RESOURCE")
		if err != nil {
			return piperutils.Path{}, fmt.Errorf("failed to retrieve project policy alerts from WhiteSource: %w", err)
		}
		policyViolationCount += len(alerts)
	}

	violations := struct {
		PolicyViolations int      `json:"policyViolations"`
		Reports          []string `json:"reports"`
	}{
		PolicyViolations: policyViolationCount,
		Reports:          []string{},
	}
	for _, report := range reportPaths {
		_, reportFile := filepath.Split(report.Target)
		violations.Reports = append(violations.Reports, reportFile)
	}

	violationContent, err := json.Marshal(violations)
	if err != nil {
		return piperutils.Path{}, fmt.Errorf("failed to marshal policy violation data: %w", err)
	}

	jsonViolationReportPath := filepath.Join(ws.ReportsDirectory, "whitesource-ip.json")
	err = utils.FileWrite(jsonViolationReportPath, violationContent, 0666)
	if err != nil {
		return piperutils.Path{}, fmt.Errorf("failed to write policy violation report: %w", err)
	}

	policyReport := piperutils.Path{Name: "WhiteSource Policy Violation Report", Target: jsonViolationReportPath}

	// create a json report to be used later, e.g. issue creation in GitHub
	ipReport := reporting.ScanReport{
		Title: "WhiteSource IP Report",
		Subheaders: []reporting.Subheader{
			{Description: "WhiteSource product name", Details: config.ProductName},
			{Description: "Filtered project names", Details: strings.Join(scan.ScannedProjectNames(), ", ")},
		},
		Overview: []reporting.OverviewRow{
			{Description: "Total number of licensing vulnerabilities", Details: fmt.Sprint(policyViolationCount)},
		},
		SuccessfulScan: policyViolationCount == 0,
		ReportTime:     utils.Now(),
	}

	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	jsonReport, _ := ipReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return policyReport, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("whitesourceExecuteScan_ip_%v.json", reportSha(config, scan))), jsonReport, 0666); err != nil {
		return policyReport, errors.Wrapf(err, "failed to write json report")
	}
	// we do not add the json report to the overall list of reports for now,
	// since it is just an intermediary report used as input for later
	// and there does not seem to be real benefit in archiving it.

	if policyViolationCount > 0 {
		log.SetErrorCategory(log.ErrorCompliance)
		influx.whitesource_data.fields.policy_violations = policyViolationCount
		return policyReport, fmt.Errorf("%v policy violation(s) found", policyViolationCount)
	}

	return policyReport, nil
}

func checkSecurityViolations(config *ScanOptions, scan *ws.Scan, sys whitesource, utils whitesourceUtils, influx *whitesourceExecuteScanInflux) ([]piperutils.Path, error) {
	var reportPaths []piperutils.Path
	// Check for security vulnerabilities and fail the build if cvssSeverityLimit threshold is crossed
	// convert config.CvssSeverityLimit to float64
	cvssSeverityLimit, err := strconv.ParseFloat(config.CvssSeverityLimit, 64)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, fmt.Errorf("failed to parse parameter cvssSeverityLimit (%s) "+
			"as floating point number: %w", config.CvssSeverityLimit, err)
	}

	if config.ProjectToken != "" {
		project := ws.Project{Name: config.ProjectName, Token: config.ProjectToken}
		// ToDo: see if HTML report generation is really required here
		// we anyway need to do some refactoring here since config.ProjectToken != "" essentially indicates an aggregated project
		if _, _, err := checkProjectSecurityViolations(cvssSeverityLimit, project, sys, influx); err != nil {
			return reportPaths, err
		}
	} else {
		vulnerabilitiesCount := 0
		var errorsOccured []string
		allAlerts := []ws.Alert{}
		for _, project := range scan.ScannedProjects() {
			// collect errors and aggregate vulnerabilities from all projects
			if vulCount, alerts, err := checkProjectSecurityViolations(cvssSeverityLimit, project, sys, influx); err != nil {
				allAlerts = append(allAlerts, alerts...)
				vulnerabilitiesCount += vulCount
				errorsOccured = append(errorsOccured, fmt.Sprint(err))
			}
		}

		scanReport := createCustomVulnerabilityReport(config, scan, allAlerts, cvssSeverityLimit, utils)
		reportPaths, err = writeCustomVulnerabilityReports(config, scan, scanReport, utils)
		if err != nil {
			errorsOccured = append(errorsOccured, fmt.Sprint(err))
		}

		if len(errorsOccured) > 0 {
			if vulnerabilitiesCount > 0 {
				log.SetErrorCategory(log.ErrorCompliance)
			}
			return reportPaths, fmt.Errorf(strings.Join(errorsOccured, ": "))
		}
	}
	return reportPaths, nil
}

// checkSecurityViolations checks security violations and returns an error if the configured severity limit is crossed.
func checkProjectSecurityViolations(cvssSeverityLimit float64, project ws.Project, sys whitesource, influx *whitesourceExecuteScanInflux) (int, []ws.Alert, error) {
	// get project alerts (vulnerabilities)
	alerts, err := sys.GetProjectAlertsByType(project.Token, "SECURITY_VULNERABILITY")
	if err != nil {
		return 0, alerts, fmt.Errorf("failed to retrieve project alerts from WhiteSource: %w", err)
	}

	severeVulnerabilities, nonSevereVulnerabilities := countSecurityVulnerabilities(&alerts, cvssSeverityLimit)
	influx.whitesource_data.fields.minor_vulnerabilities = nonSevereVulnerabilities
	influx.whitesource_data.fields.major_vulnerabilities = severeVulnerabilities
	influx.whitesource_data.fields.vulnerabilities = nonSevereVulnerabilities + severeVulnerabilities
	if nonSevereVulnerabilities > 0 {
		log.Entry().Warnf("WARNING: %v Open Source Software Security vulnerabilities with "+
			"CVSS score below threshold %.1f detected in project %s.", nonSevereVulnerabilities,
			cvssSeverityLimit, project.Name)
	} else if len(alerts) == 0 {
		log.Entry().Infof("No Open Source Software Security vulnerabilities detected in project %s",
			project.Name)
	}
	// https://github.com/SAP/jenkins-library/blob/master/vars/whitesourceExecuteScan.groovy#L558
	if severeVulnerabilities > 0 {
		log.SetErrorCategory(log.ErrorCompliance)
		return severeVulnerabilities, alerts, fmt.Errorf("%v Open Source Software Security vulnerabilities with CVSS score greater "+
			"or equal to %.1f detected in project %s",
			severeVulnerabilities, cvssSeverityLimit, project.Name)
	}
	return 0, alerts, nil
}

func countSecurityVulnerabilities(alerts *[]ws.Alert, cvssSeverityLimit float64) (int, int) {
	severeVulnerabilities := 0
	for _, alert := range *alerts {
		if isSevereVulnerability(alert, cvssSeverityLimit) {
			severeVulnerabilities++
		}
	}

	nonSevereVulnerabilities := len(*alerts) - severeVulnerabilities
	return severeVulnerabilities, nonSevereVulnerabilities
}

func isSevereVulnerability(alert ws.Alert, cvssSeverityLimit float64) bool {

	if vulnerabilityScore(alert) >= cvssSeverityLimit && cvssSeverityLimit >= 0 {
		return true
	}
	return false
}

func createCustomVulnerabilityReport(config *ScanOptions, scan *ws.Scan, alerts []ws.Alert, cvssSeverityLimit float64, utils whitesourceUtils) reporting.ScanReport {

	severe, _ := countSecurityVulnerabilities(&alerts, cvssSeverityLimit)

	// sort according to vulnerability severity
	sort.Slice(alerts, func(i, j int) bool {
		return vulnerabilityScore(alerts[i]) > vulnerabilityScore(alerts[j])
	})

	projectNames := scan.ScannedProjectNames()

	scanReport := reporting.ScanReport{
		Title: "WhiteSource Security Vulnerability Report",
		Subheaders: []reporting.Subheader{
			{Description: "WhiteSource product name", Details: config.ProductName},
			{Description: "Filtered project names", Details: strings.Join(projectNames, ", ")},
		},
		Overview: []reporting.OverviewRow{
			{Description: "Total number of vulnerabilities", Details: fmt.Sprint(len(alerts))},
			{Description: "Total number of high/critical vulnerabilities with CVSS score >= 7.0", Details: fmt.Sprint(severe)},
		},
		SuccessfulScan: severe == 0,
		ReportTime:     utils.Now(),
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

	for _, alert := range alerts {
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
		emptyFix := ws.Fix{}
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

func writeCustomVulnerabilityReports(config *ScanOptions, scan *ws.Scan, scanReport reporting.ScanReport, utils whitesourceUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := filepath.Join(ws.ReportsDirectory, "piper_whitesource_vulnerability_report.html")
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write html report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "WhiteSource Vulnerability Report", Target: htmlReportPath})

	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("whitesourceExecuteScan_oss_%v.json", reportSha(config, scan))), jsonReport, 0666); err != nil {
		return reportPaths, errors.Wrapf(err, "failed to write json report")
	}
	// we do not add the json report to the overall list of reports for now,
	// since it is just an intermediary report used as input for later
	// and there does not seem to be real benefit in archiving it.

	return reportPaths, nil
}

func vulnerabilityScore(alert ws.Alert) float64 {
	if alert.Vulnerability.CVSS3Score > 0 {
		return alert.Vulnerability.CVSS3Score
	}
	return alert.Vulnerability.Score
}

func reportSha(config *ScanOptions, scan *ws.Scan) string {
	reportShaData := []byte(config.ProductName + "," + strings.Join(scan.ScannedProjectNames(), ","))
	return fmt.Sprintf("%x", sha1.Sum(reportShaData))
}

func aggregateVersionWideLibraries(config *ScanOptions, utils whitesourceUtils, sys whitesource) error {
	log.Entry().Infof("Aggregating list of libraries used for all projects with version: %s", config.Version)

	projects, err := sys.GetProjectsMetaInfo(config.ProductToken)
	if err != nil {
		return errors.Wrapf(err, "failed to get projects meta info")
	}

	versionWideLibraries := map[string][]ws.Library{} // maps project name to slice of libraries
	for _, project := range projects {
		projectVersion := strings.Split(project.Name, " - ")[1]
		projectName := strings.Split(project.Name, " - ")[0]
		if projectVersion == config.Version {
			libs, err := sys.GetProjectLibraryLocations(project.Token)
			if err != nil {
				return errors.Wrapf(err, "failed to get project library locations")
			}
			log.Entry().Infof("Found project: %s with %v libraries.", project.Name, len(libs))
			versionWideLibraries[projectName] = libs
		}
	}
	if err := newLibraryCSVReport(versionWideLibraries, config, utils); err != nil {
		return errors.Wrapf(err, "failed toget new libary CSV report")
	}
	return nil
}

func aggregateVersionWideVulnerabilities(config *ScanOptions, utils whitesourceUtils, sys whitesource) error {
	log.Entry().Infof("Aggregating list of vulnerabilities for all projects with version: %s", config.Version)

	projects, err := sys.GetProjectsMetaInfo(config.ProductToken)
	if err != nil {
		return errors.Wrapf(err, "failed to get projects meta info")
	}

	var versionWideAlerts []ws.Alert // all alerts for a given project version
	projectNames := ``               // holds all project tokens considered a part of the report for debugging
	for _, project := range projects {
		projectVersion := strings.Split(project.Name, " - ")[1]
		if projectVersion == config.Version {
			projectNames += project.Name + "\n"
			alerts, err := sys.GetProjectAlertsByType(project.Token, "SECURITY_VULNERABILITY")
			if err != nil {
				return errors.Wrapf(err, "failed to get project alerts by type")
			}
			log.Entry().Infof("Found project: %s with %v vulnerabilities.", project.Name, len(alerts))
			versionWideAlerts = append(versionWideAlerts, alerts...)
		}
	}

	reportPath := filepath.Join(ws.ReportsDirectory, "project-names-aggregated.txt")
	if err := utils.FileWrite(reportPath, []byte(projectNames), 0666); err != nil {
		return errors.Wrapf(err, "failed to write report: %s", reportPath)
	}
	if err := newVulnerabilityExcelReport(versionWideAlerts, config, utils); err != nil {
		return errors.Wrapf(err, "failed to create new vulnerability excel report")
	}
	return nil
}

const wsReportTimeStampLayout = "20060102-150405"

// outputs an slice of alerts to an excel file
func newVulnerabilityExcelReport(alerts []ws.Alert, config *ScanOptions, utils whitesourceUtils) error {
	file := excelize.NewFile()
	streamWriter, err := file.NewStreamWriter("Sheet1")
	if err != nil {
		return err
	}
	styleID, err := file.NewStyle(`{"font":{"color":"#777777"}}`)
	if err != nil {
		return err
	}
	if err := fillVulnerabilityExcelReport(alerts, streamWriter, styleID); err != nil {
		return err
	}
	if err := streamWriter.Flush(); err != nil {
		return err
	}

	if err := utils.MkdirAll(ws.ReportsDirectory, 0777); err != nil {
		return err
	}

	fileName := filepath.Join(ws.ReportsDirectory,
		fmt.Sprintf("vulnerabilities-%s.xlsx", utils.Now().Format(wsReportTimeStampLayout)))
	stream, err := utils.FileOpen(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	if err := file.Write(stream); err != nil {
		return err
	}
	filePath := piperutils.Path{Name: "aggregated-vulnerabilities", Target: fileName}
	piperutils.PersistReportsAndLinks("whitesourceExecuteScan", "", []piperutils.Path{filePath}, nil)
	return nil
}

func fillVulnerabilityExcelReport(alerts []ws.Alert, streamWriter *excelize.StreamWriter, styleID int) error {
	rows := []struct {
		axis  string
		title string
	}{
		{"A1", "Severity"},
		{"B1", "Library"},
		{"C1", "Vulnerability Id"},
		{"D1", "CVSS 3"},
		{"E1", "Project"},
		{"F1", "Resolution"},
	}
	for _, row := range rows {
		err := streamWriter.SetRow(row.axis, []interface{}{excelize.Cell{StyleID: styleID, Value: row.title}})
		if err != nil {
			return err
		}
	}

	for i, alert := range alerts {
		row := make([]interface{}, 6)
		vuln := alert.Vulnerability
		row[0] = vuln.CVSS3Severity
		row[1] = alert.Library.Filename
		row[2] = vuln.Name
		row[3] = vuln.CVSS3Score
		row[4] = alert.Project
		row[5] = vuln.FixResolutionText
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		if err := streamWriter.SetRow(cell, row); err != nil {
			log.Entry().Errorf("failed to write alert row: %v", err)
		}
	}
	return nil
}

// outputs an slice of libraries to an excel file based on projects with version == config.Version
func newLibraryCSVReport(libraries map[string][]ws.Library, config *ScanOptions, utils whitesourceUtils) error {
	output := "Library Name, Project Name\n"
	for projectName, libraries := range libraries {
		log.Entry().Infof("Writing %v libraries for project %s to excel report..", len(libraries), projectName)
		for _, library := range libraries {
			output += library.Name + ", " + projectName + "\n"
		}
	}

	// Ensure reporting directory exists
	if err := utils.MkdirAll(ws.ReportsDirectory, 0777); err != nil {
		return errors.Wrapf(err, "failed to create directories: %s", ws.ReportsDirectory)
	}

	// Write result to file
	fileName := fmt.Sprintf("%s/libraries-%s.csv", ws.ReportsDirectory,
		utils.Now().Format(wsReportTimeStampLayout))
	if err := utils.FileWrite(fileName, []byte(output), 0666); err != nil {
		return errors.Wrapf(err, "failed to write file: %s", fileName)
	}
	filePath := piperutils.Path{Name: "aggregated-libraries", Target: fileName}
	piperutils.PersistReportsAndLinks("whitesourceExecuteScan", "", []piperutils.Path{filePath}, nil)
	return nil
}

// persistScannedProjects writes all actually scanned WhiteSource project names as list
// into the Common Pipeline Environment, from where it can be used by sub-sequent steps.
func persistScannedProjects(config *ScanOptions, scan *ws.Scan, commonPipelineEnvironment *whitesourceExecuteScanCommonPipelineEnvironment) {
	projectNames := []string{}
	if config.ProjectName != "" {
		projectNames = []string{config.ProjectName + " - " + config.Version}
	} else {
		projectNames = scan.ScannedProjectNames()
	}
	commonPipelineEnvironment.custom.whitesourceProjectNames = projectNames
}

// create toolrecord file for whitesource
//
func createToolRecordWhitesource(workspace string, config *whitesourceExecuteScanOptions, scan *ws.Scan) (string, error) {
	record := toolrecord.New(workspace, "whitesource", config.ServiceURL)
	wsUiRoot := "https://saas.whitesourcesoftware.com"
	productURL := wsUiRoot + "/Wss/WSS.html#!product;token=" + config.ProductToken
	err := record.AddKeyData("product",
		config.ProductToken,
		config.ProductName,
		productURL)
	if err != nil {
		return "", err
	}
	max_idx := 0
	for idx, project := range scan.ScannedProjects() {
		max_idx = idx
		name := project.Name
		token := project.Token
		projectURL := ""
		if token != "" {
			projectURL = wsUiRoot + "/Wss/WSS.html#!project;token=" + token
		} else {
			// token is empty, provide a dummy to have an indication
			token = "unknown"
		}
		err = record.AddKeyData("project",
			token,
			name,
			projectURL)
		if err != nil {
			return "", err
		}
	}
	// set overall display data to product if there
	// is more than one project
	if max_idx > 1 {
		record.SetOverallDisplayData(config.ProductName, productURL)
	}
	err = record.Persist()
	if err != nil {
		return "", err
	}
	return record.GetFileName(), nil
}
