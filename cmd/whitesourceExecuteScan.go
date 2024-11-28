package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	piperDocker "github.com/SAP/jenkins-library/pkg/docker"
	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	ws "github.com/SAP/jenkins-library/pkg/whitesource"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/golang"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"

	"github.com/google/go-github/v68/github"
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
	GetProjectIgnoredAlertsByType(projectToken string, alertType string) ([]ws.Alert, error)
	GetProjectLibraryLocations(projectToken string) ([]ws.Library, error)
	GetProjectHierarchy(projectToken string, includeInHouse bool) ([]ws.Library, error)
}

type whitesourceUtils interface {
	ws.Utils
	piperutils.FileUtils
	GetArtifactCoordinates(buildTool, buildDescriptorFile string, options *versioning.Options) (versioning.Coordinates, error)
	Now() time.Time
	GetIssueService() *github.IssuesService
	GetSearchService() *github.SearchService
}

type whitesourceUtilsBundle struct {
	*piperhttp.Client
	*command.Command
	*piperutils.Files
	npmExecutor npm.Executor
	issues      *github.IssuesService
	search      *github.SearchService
}

func (w *whitesourceUtilsBundle) FileOpen(name string, flag int, perm os.FileMode) (ws.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (w *whitesourceUtilsBundle) GetArtifactCoordinates(buildTool, buildDescriptorFile string, options *versioning.Options) (versioning.Coordinates, error) {
	if err := validationBuildDescriptorFile(buildTool, buildDescriptorFile); err != nil {
		return versioning.Coordinates{}, err
	}
	artifact, err := versioning.GetArtifact(buildTool, buildDescriptorFile, options, w)
	if err != nil {
		return versioning.Coordinates{}, err
	}
	return artifact.GetCoordinates()
}

func validationBuildDescriptorFile(buildTool, buildDescriptorFile string) error {
	if buildDescriptorFile == "" {
		return nil
	}
	switch buildTool {
	case "dub":
		if filepath.Ext(buildDescriptorFile) != ".json" {
			return errors.New("extension of buildDescriptorFile must be in '*.json'")
		}
	case "gradle":
		if filepath.Ext(buildDescriptorFile) != ".properties" {
			return errors.New("extension of buildDescriptorFile must be in '*.properties'")
		}
	case "golang":
		if !strings.HasSuffix(buildDescriptorFile, "go.mod") &&
			!strings.HasSuffix(buildDescriptorFile, "VERSION") &&
			!strings.HasSuffix(buildDescriptorFile, "version.txt") {
			return errors.New("buildDescriptorFile must be one of  [\"go.mod\",\"VERSION\", \"version.txt\"]")
		}
	case "maven":
		if filepath.Ext(buildDescriptorFile) != ".xml" {
			return errors.New("extension of buildDescriptorFile must be in '*.xml'")
		}
	case "mta":
		if filepath.Ext(buildDescriptorFile) != ".yaml" {
			return errors.New("extension of buildDescriptorFile must be in '*.yaml'")
		}
	case "npm", "yarn":
		if filepath.Ext(buildDescriptorFile) != ".json" {
			return errors.New("extension of buildDescriptorFile must be in '*.json'")
		}
	case "pip":
		if !strings.HasSuffix(buildDescriptorFile, "setup.py") &&
			!strings.HasSuffix(buildDescriptorFile, "version.txt") &&
			!strings.HasSuffix(buildDescriptorFile, "VERSION") {
			return errors.New("buildDescriptorFile must be one of  [\"setup.py\",\"version.txt\", \"VERSION\"]")
		}
	case "sbt":
		if !strings.HasSuffix(buildDescriptorFile, "sbtDescriptor.json") &&
			!strings.HasSuffix(buildDescriptorFile, "build.sbt") {
			return errors.New("extension of buildDescriptorFile must be in '*.json' or '*sbt'")
		}
	}
	return nil
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

func (w *whitesourceUtilsBundle) GetIssueService() *github.IssuesService {
	return w.issues
}

func (w *whitesourceUtilsBundle) GetSearchService() *github.SearchService {
	return w.search
}

func newWhitesourceUtils(config *ScanOptions, client *github.Client) *whitesourceUtilsBundle {
	utils := whitesourceUtilsBundle{
		Client:  &piperhttp.Client{},
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	if client != nil {
		utils.issues = client.Issues
		utils.search = client.Search
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
		AggregateProjectName:        config.ProjectName,
		ProductVersion:              config.Version,
		BuildTool:                   config.BuildTool,
		SkipProjectsWithEmptyTokens: config.SkipProjectsWithEmptyTokens,
	}
}

func whitesourceExecuteScan(config ScanOptions, _ *telemetry.CustomData, commonPipelineEnvironment *whitesourceExecuteScanCommonPipelineEnvironment, influx *whitesourceExecuteScanInflux) {
	ctx, client, err := piperGithub.
		NewClientBuilder(config.GithubToken, config.GithubAPIURL).
		WithTrustedCerts(config.CustomTLSCertificateLinks).Build()
	if err != nil {
		log.Entry().WithError(err).Warning("Failed to get GitHub client")
	}
	if log.IsVerbose() {
		logConfigInVerboseModeForWhitesource(config)
		logWorkspaceContent()
	}
	utils := newWhitesourceUtils(&config, client)
	scan := newWhitesourceScan(&config)
	sys := ws.NewSystem(config.ServiceURL, config.OrgToken, config.UserToken, time.Duration(config.Timeout)*time.Second)
	influx.step_data.fields.whitesource = false
	if err := runWhitesourceExecuteScan(ctx, &config, scan, utils, sys, commonPipelineEnvironment, influx); err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	influx.step_data.fields.whitesource = true
}

func runWhitesourceExecuteScan(ctx context.Context, config *ScanOptions, scan *ws.Scan, utils whitesourceUtils, sys whitesource, commonPipelineEnvironment *whitesourceExecuteScanCommonPipelineEnvironment, influx *whitesourceExecuteScanInflux) error {
	if config != nil && config.PrivateModules != "" && config.PrivateModulesGitToken != "" {
		//configuring go private packages
		if err := golang.PrepareGolangPrivatePackages("WhitesourceExecuteStep", config.PrivateModules, config.PrivateModulesGitToken); err != nil {
			log.Entry().Warningf("couldn't set private packages for golang, error: %s", err.Error())
		}
	}

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
		if err := runWhitesourceScan(ctx, config, scan, utils, sys, commonPipelineEnvironment, influx); err != nil {
			return errors.Wrapf(err, "failed to execute WhiteSource scan")
		}
	}
	return nil
}

func runWhitesourceScan(ctx context.Context, config *ScanOptions, scan *ws.Scan, utils whitesourceUtils, sys whitesource, commonPipelineEnvironment *whitesourceExecuteScanCommonPipelineEnvironment, influx *whitesourceExecuteScanInflux) error {

	// Download Docker image for container scan
	// ToDo: move it to improve testability
	if config.BuildTool == "docker" {
		if len(config.ScanImages) != 0 && config.ActivateMultipleImagesScan {
			for _, image := range config.ScanImages {
				config.ScanImage = image
				err := downloadMultipleDockerImageAsTar(config, utils)
				if err != nil {
					return errors.Wrapf(err, "failed to download docker image")
				}
			}

		} else {
			err := downloadDockerImageAsTar(config, utils)
			if err != nil {
				return errors.Wrapf(err, "failed to download docker image")
			}
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

	paths, err := checkAndReportScanResults(ctx, config, scan, utils, sys, influx)
	piperutils.PersistReportsAndLinks("whitesourceExecuteScan", "", utils, paths, nil)
	persistScannedProjects(config, scan, commonPipelineEnvironment)
	if err != nil {
		return errors.Wrapf(err, "failed to check and report scan results")
	}
	return nil
}

func checkAndReportScanResults(ctx context.Context, config *ScanOptions, scan *ws.Scan, utils whitesourceUtils, sys whitesource, influx *whitesourceExecuteScanInflux) ([]piperutils.Path, error) {
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

	rPath, err := checkPolicyViolations(ctx, config, scan, sys, utils, reportPaths, influx)

	if err != nil {
		if !config.FailOnSevereVulnerabilities && log.GetErrorCategory() == log.ErrorCompliance {
			log.Entry().Infof("policy violation(s) found - step will only create data but not fail due to setting failOnSevereVulnerabilities: false")
		} else {
			checkErrors = append(checkErrors, fmt.Sprint(err))
		}
	}
	reportPaths = append(reportPaths, rPath)

	if config.SecurityVulnerabilities {
		rPaths, err := checkSecurityViolations(ctx, config, scan, sys, utils, influx)
		reportPaths = append(reportPaths, rPaths...)
		if err != nil {
			if !config.FailOnSevereVulnerabilities && log.GetErrorCategory() == log.ErrorCompliance {
				log.Entry().Infof("policy violation(s) found - step will only create data but not fail due to setting failOnSevereVulnerabilities: false")
			} else {
				checkErrors = append(checkErrors, fmt.Sprint(err))
			}
		}
	}

	// create toolrecord file
	// tbd - how to handle verifyOnly
	toolRecordFileName, err := createToolRecordWhitesource(utils, "./", config, scan)
	if err != nil {
		// do not fail until the framework is well established
		log.Entry().Warning("TR_WHITESOURCE: Failed to create toolrecord file ...", err)
	} else {
		reportPaths = append(reportPaths, piperutils.Path{Target: toolRecordFileName})
	}

	if len(checkErrors) > 0 {
		return reportPaths, errors.New(strings.Join(checkErrors, ": "))
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
		if len(config.CustomScanVersion) > 0 {
			log.Entry().Infof("Using custom version: %v", config.CustomScanVersion)
			config.Version = config.CustomScanVersion
		} else if len(config.Version) > 0 {
			log.Entry().Infof("Resolving product version from default provided '%s' with versioning '%s'", config.Version, config.VersioningModel)
			config.Version = versioning.ApplyVersioningModel(config.VersioningModel, config.Version)
			log.Entry().Infof("Resolved product version '%s'", config.Version)
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
		scan.Coordinates = coordinates

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

	if !config.SkipParentProjectResolution {
		if err := resolveAggregateProjectToken(config, sys); err != nil {
			return errors.Wrap(err, "error resolving aggregate project token")
		}
	}

	scan.ProductToken = config.ProductToken

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
		BuildTool:                       config.BuildTool,
		ScanType:                        "", // no longer provided via config
		OrgToken:                        config.OrgToken,
		UserToken:                       config.UserToken,
		ProductName:                     config.ProductName,
		ProductToken:                    config.ProductToken,
		ProductVersion:                  config.Version,
		ProjectName:                     config.ProjectName,
		BuildDescriptorFile:             config.BuildDescriptorFile,
		BuildDescriptorExcludeList:      config.BuildDescriptorExcludeList,
		PomPath:                         config.BuildDescriptorFile,
		M2Path:                          config.M2Path,
		GlobalSettingsFile:              config.GlobalSettingsFile,
		ProjectSettingsFile:             config.ProjectSettingsFile,
		InstallArtifacts:                config.InstallArtifacts,
		DefaultNpmRegistry:              config.DefaultNpmRegistry,
		NpmIncludeDevDependencies:       config.NpmIncludeDevDependencies,
		AgentDownloadURL:                config.AgentDownloadURL,
		AgentFileName:                   config.AgentFileName,
		ConfigFilePath:                  config.ConfigFilePath,
		UseGlobalConfiguration:          config.UseGlobalConfiguration,
		Includes:                        config.Includes,
		Excludes:                        config.Excludes,
		JreDownloadURL:                  config.JreDownloadURL,
		AgentURL:                        config.AgentURL,
		ServiceURL:                      config.ServiceURL,
		ScanPath:                        config.ScanPath,
		InstallCommand:                  config.InstallCommand,
		Verbose:                         GeneralConfig.Verbose,
		SkipParentProjectResolution:     config.SkipParentProjectResolution,
		DisableNpmSubmodulesAggregation: config.DisableNpmSubmodulesAggregation,
	}
}

// Unified Agent is the only supported option by WhiteSource going forward:
// The Unified Agent will be used to perform the scan.
func executeScan(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils) error {
	options := wsScanOptions(config)

	if options.InstallCommand != "" {
		installCommandTokens := strings.Split(config.InstallCommand, " ")
		if err := utils.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...); err != nil {
			log.SetErrorCategory(log.ErrorCustom)
			return errors.Wrapf(err, "failed to execute install command: %v", config.InstallCommand)
		}
	}

	// Execute scan with Unified Agent jar file
	if err := scan.ExecuteUAScan(options, utils); err != nil {
		return errors.Wrapf(err, "failed to execute Unified Agent scan")
	}
	return nil
}

func checkPolicyViolations(ctx context.Context, config *ScanOptions, scan *ws.Scan, sys whitesource, utils whitesourceUtils, reportPaths []piperutils.Path, influx *whitesourceExecuteScanInflux) (piperutils.Path, error) {
	policyViolationCount := 0
	allAlerts := []ws.Alert{}
	for _, project := range scan.ScannedProjects() {
		alerts, err := sys.GetProjectAlertsByType(project.Token, "REJECTED_BY_POLICY_RESOURCE")
		if err != nil {
			return piperutils.Path{}, fmt.Errorf("failed to retrieve project policy alerts from WhiteSource: %w", err)
		}

		policyViolationCount += len(alerts)
		allAlerts = append(allAlerts, alerts...)
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
	err = utils.FileWrite(jsonViolationReportPath, violationContent, 0o666)
	if err != nil {
		return piperutils.Path{}, fmt.Errorf("failed to write policy violation report: %w", err)
	}

	policyReport := piperutils.Path{Name: "WhiteSource Policy Violation Report", Target: jsonViolationReportPath}

	// create a json report to be used later, e.g. issue creation in GitHub
	ipReport := reporting.ScanReport{
		ReportTitle: "WhiteSource IP Report",
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
		err := utils.MkdirAll(reporting.StepReportDirectory, 0o777)
		if err != nil {
			return policyReport, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("whitesourceExecuteScan_ip_%v.json", ws.ReportSha(config.ProductName, scan))), jsonReport, 0o666); err != nil {
		return policyReport, errors.Wrap(err, "failed to write json report")
	}
	// we do not add the json report to the overall list of reports for now,
	// since it is just an intermediary report used as input for later
	// and there does not seem to be real benefit in archiving it.

	if policyViolationCount > 0 {
		influx.whitesource_data.fields.policy_violations = policyViolationCount
		log.SetErrorCategory(log.ErrorCompliance)

		if config.CreateResultIssue && policyViolationCount > 0 && len(config.GithubToken) > 0 && len(config.GithubAPIURL) > 0 && len(config.Owner) > 0 && len(config.Repository) > 0 {
			log.Entry().Debugf("Creating result issues for %v alert(s)", policyViolationCount)
			issueDetails := make([]reporting.IssueDetail, len(allAlerts))
			piperutils.CopyAtoB(allAlerts, issueDetails)
			gh := reporting.GitHub{
				Owner:         &config.Owner,
				Repository:    &config.Repository,
				Assignees:     &config.Assignees,
				IssueService:  utils.GetIssueService(),
				SearchService: utils.GetSearchService(),
			}
			if err := gh.UploadMultipleReports(ctx, &issueDetails); err != nil {
				return policyReport, fmt.Errorf("failed to upload reports to GitHub for %v policy violations: %w", policyViolationCount, err)
			}
		}
		return policyReport, fmt.Errorf("%v policy violation(s) found", policyViolationCount)
	}

	return policyReport, nil
}

func checkSecurityViolations(ctx context.Context, config *ScanOptions, scan *ws.Scan, sys whitesource, utils whitesourceUtils, influx *whitesourceExecuteScanInflux) ([]piperutils.Path, error) {
	// Check for security vulnerabilities and fail the build if cvssSeverityLimit threshold is crossed
	// convert config.CvssSeverityLimit to float64
	cvssSeverityLimit, err := strconv.ParseFloat(config.CvssSeverityLimit, 64)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return []piperutils.Path{}, fmt.Errorf("failed to parse parameter cvssSeverityLimit (%s) "+
			"as floating point number: %w", config.CvssSeverityLimit, err)
	}

	// inhale assessments from file system
	assessments := readAssessmentsFromFile(config.AssessmentFile, utils)

	vulnerabilitiesCount := 0
	var allOccurredErrors []string
	allAlerts := []ws.Alert{}
	allAssessedAlerts := []ws.Alert{}
	allLibraries := []ws.Library{}

	if config.ProjectToken != "" {
		project := ws.Project{Name: config.ProjectName, Token: config.ProjectToken}
		// ToDo: see if HTML report generation is really required here
		// we anyway need to do some refactoring here since config.ProjectToken != "" essentially indicates an aggregated project

		vulnerabilitiesCount, allAlerts, allAssessedAlerts, allLibraries, allOccurredErrors = collectVulnsAndLibsForProject(
			config,
			cvssSeverityLimit,
			project,
			sys,
			assessments,
			influx,
		)

		log.Entry().Debugf("Collected %v libraries for project %v", len(allLibraries), project.Name)

	} else {
		for _, project := range scan.ScannedProjects() {
			// collect errors and aggregate vulnerabilities from all projects
			vulCount, alerts, assessedAlerts, libraries, occurredErrors := collectVulnsAndLibsForProject(
				config,
				cvssSeverityLimit,
				project,
				sys,
				assessments,
				influx,
			)
			if len(occurredErrors) != 0 {
				allOccurredErrors = append(allOccurredErrors, occurredErrors...)
			}

			allAlerts = append(allAlerts, alerts...)
			allAssessedAlerts = append(allAssessedAlerts, assessedAlerts...)
			vulnerabilitiesCount += vulCount
			allLibraries = append(allLibraries, libraries...)
		}
		log.Entry().Debugf("Aggregated %v alerts for scanned projects", len(allAlerts))
	}

	reportPaths, e := reportGitHubIssuesAndCreateReports(
		ctx,
		config,
		utils,
		scan,
		allAlerts,
		allLibraries,
		allAssessedAlerts,
		cvssSeverityLimit,
		vulnerabilitiesCount,
	)

	allOccurredErrors = append(allOccurredErrors, e...)

	if len(allOccurredErrors) > 0 {
		if vulnerabilitiesCount > 0 {
			log.SetErrorCategory(log.ErrorCompliance)
		}
		return reportPaths, errors.New(strings.Join(allOccurredErrors, ": "))
	}

	return reportPaths, nil
}

func collectVulnsAndLibsForProject(
	config *ScanOptions,
	cvssSeverityLimit float64,
	project ws.Project,
	sys whitesource,
	assessments *[]format.Assessment,
	influx *whitesourceExecuteScanInflux,
) (
	int,
	[]ws.Alert,
	[]ws.Alert,
	[]ws.Library,
	[]string,
) {
	var errorsOccurred []string
	vulCount, alerts, assessedAlerts, err := checkProjectSecurityViolations(config, cvssSeverityLimit, project, sys, assessments, influx)
	if err != nil {
		errorsOccurred = append(errorsOccurred, fmt.Sprint(err))
	}
	log.Entry().Infof("Current influx data : minor_vulnerabilities = %v / major_vulnerabilities = %v / vulnerabilities = %v", influx.whitesource_data.fields.minor_vulnerabilities, influx.whitesource_data.fields.major_vulnerabilities, influx.whitesource_data.fields.vulnerabilities)

	// collect all libraries detected in all related projects and errors
	libraries, err := sys.GetProjectHierarchy(project.Token, true)
	if err != nil {
		errorsOccurred = append(errorsOccurred, fmt.Sprint(err))
	}
	log.Entry().Debugf("Collected %v libraries for project %v", len(libraries), project.Name)

	return vulCount, alerts, assessedAlerts, libraries, errorsOccurred
}

func reportGitHubIssuesAndCreateReports(
	ctx context.Context,
	config *ScanOptions,
	utils whitesourceUtils,
	scan *ws.Scan,
	allAlerts []ws.Alert,
	allLibraries []ws.Library,
	allAssessedAlerts []ws.Alert,
	cvssSeverityLimit float64,
	vulnerabilitiesCount int,
) ([]piperutils.Path, []string) {
	errorsOccured := make([]string, 0)
	reportPaths := make([]piperutils.Path, 0)

	if config.CreateResultIssue && vulnerabilitiesCount > 0 && len(config.GithubToken) > 0 && len(config.GithubAPIURL) > 0 && len(config.Owner) > 0 && len(config.Repository) > 0 {
		log.Entry().Debugf("Creating result issues for %v alert(s)", vulnerabilitiesCount)
		issueDetails := make([]reporting.IssueDetail, len(allAlerts))
		piperutils.CopyAtoB(allAlerts, issueDetails)
		gh := reporting.GitHub{
			Owner:         &config.Owner,
			Repository:    &config.Repository,
			Assignees:     &config.Assignees,
			IssueService:  utils.GetIssueService(),
			SearchService: utils.GetSearchService(),
		}

		if err := gh.UploadMultipleReports(ctx, &issueDetails); err != nil {
			errorsOccured = append(errorsOccured, fmt.Sprint(err))
		}
	}

	scanReport := ws.CreateCustomVulnerabilityReport(config.ProductName, scan, &allAlerts, cvssSeverityLimit)
	paths, err := ws.WriteCustomVulnerabilityReports(config.ProductName, scan, scanReport, utils)
	if err != nil {
		errorsOccured = append(errorsOccured, fmt.Sprint(err))
	}

	reportPaths = append(reportPaths, paths...)

	combinedAlerts := make([]ws.Alert, 0, len(allAlerts)+len(allAssessedAlerts))
	combinedAlerts = append(combinedAlerts, allAlerts...)
	combinedAlerts = append(combinedAlerts, allAssessedAlerts...)

	sarif := ws.CreateSarifResultFile(scan, &combinedAlerts)
	paths, err = ws.WriteSarifFile(sarif, utils)
	if err != nil {
		errorsOccured = append(errorsOccured, fmt.Sprint(err))
	}

	reportPaths = append(reportPaths, paths...)

	sbom, err := ws.CreateCycloneSBOM(scan, &allLibraries, &allAlerts, &allAssessedAlerts)
	if err != nil {
		errorsOccured = append(errorsOccured, fmt.Sprint(err))
	}

	paths, err = ws.WriteCycloneSBOM(sbom, utils)
	if err != nil {
		errorsOccured = append(errorsOccured, fmt.Sprint(err))
	}

	reportPaths = append(reportPaths, paths...)

	return reportPaths, errorsOccured
}

// read assessments from file and expose them to match alerts and filter them before processing
func readAssessmentsFromFile(assessmentFilePath string, utils whitesourceUtils) *[]format.Assessment {
	exists, err := utils.FileExists(assessmentFilePath)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		log.Entry().WithError(err).Errorf("unable to check existence of assessment file at '%s'", assessmentFilePath)
	}
	assessmentFile, err := utils.Open(assessmentFilePath)
	if exists && err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		log.Entry().WithError(err).Errorf("unable to open assessment file at '%s'", assessmentFilePath)
	}
	assessments := &[]format.Assessment{}
	if exists {
		defer assessmentFile.Close()
		assessments, err = format.ReadAssessments(assessmentFile)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			log.Entry().WithError(err).Errorf("unable to parse assessment file at '%s'", assessmentFilePath)
		}
	}
	return assessments
}

// checkSecurityViolations checks security violations and returns an error if the configured severity limit is crossed. Besides the potential error the list of unassessed and assessed alerts are being returned to allow generating reports and issues from the data.
func checkProjectSecurityViolations(config *ScanOptions, cvssSeverityLimit float64, project ws.Project, sys whitesource, assessments *[]format.Assessment, influx *whitesourceExecuteScanInflux) (int, []ws.Alert, []ws.Alert, error) {
	// get project alerts (vulnerabilities)
	alerts, err := sys.GetProjectAlertsByType(project.Token, "SECURITY_VULNERABILITY")
	if err != nil {
		return 0, alerts, []ws.Alert{}, fmt.Errorf("failed to retrieve project alerts from WhiteSource: %w", err)
	}

	assessedAlerts, err := sys.GetProjectIgnoredAlertsByType(project.Token, "SECURITY_VULNERABILITY")
	if err != nil {
		return 0, alerts, []ws.Alert{}, fmt.Errorf("failed to retrieve project ignored alerts from WhiteSource: %w", err)
	}

	// filter alerts related to existing assessments
	filteredAlerts := []ws.Alert{}
	if assessments != nil && len(*assessments) > 0 {
		for _, alert := range alerts {
			if result, err := alert.ContainedIn(assessments); err == nil && !result {
				filteredAlerts = append(filteredAlerts, alert)
			} else if alert.Assessment != nil {
				log.Entry().Debugf("Matched assessment with status %v and analysis %v to vulnerability %v affecting packages %v", alert.Assessment.Status, alert.Assessment.Analysis, alert.Assessment.Vulnerability, alert.Assessment.Purls)
				assessedAlerts = append(assessedAlerts, alert)
			}
		}
		// intentionally overwriting original list of alerts with those remaining unassessed after processing of assessments
		alerts = filteredAlerts
	}

	severeVulnerabilities, nonSevereVulnerabilities := ws.CountSecurityVulnerabilities(&alerts, cvssSeverityLimit)
	influx.whitesource_data.fields.minor_vulnerabilities += nonSevereVulnerabilities
	influx.whitesource_data.fields.major_vulnerabilities += severeVulnerabilities
	influx.whitesource_data.fields.vulnerabilities += (nonSevereVulnerabilities + severeVulnerabilities)
	log.Entry().Infof("Current influx data : minor_vulnerabilities = %v / major_vulnerabilities = %v / vulnerabilities = %v", influx.whitesource_data.fields.minor_vulnerabilities, influx.whitesource_data.fields.major_vulnerabilities, influx.whitesource_data.fields.vulnerabilities)
		
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
		log.Entry().Infof("%v Open Source Software Security vulnerabilities with CVSS score greater or equal to %.1f detected in project %s", severeVulnerabilities, cvssSeverityLimit, project.Name)
		if config.FailOnSevereVulnerabilities {
			log.SetErrorCategory(log.ErrorCompliance)
			return severeVulnerabilities, alerts, assessedAlerts, fmt.Errorf("%v Open Source Software Security vulnerabilities with CVSS score greater or equal to %.1f detected in project %s", severeVulnerabilities, cvssSeverityLimit, project.Name)
		}
		log.Entry().Info("Step will only create data but not fail due to setting failOnSevereVulnerabilities: false")
		return severeVulnerabilities, alerts, assessedAlerts, nil
	}
	return 0, alerts, assessedAlerts, nil
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
	if err := utils.FileWrite(reportPath, []byte(projectNames), 0o666); err != nil {
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

	if err := utils.MkdirAll(ws.ReportsDirectory, 0o777); err != nil {
		return err
	}

	fileName := filepath.Join(ws.ReportsDirectory,
		fmt.Sprintf("vulnerabilities-%s.xlsx", utils.Now().Format(wsReportTimeStampLayout)))
	stream, err := utils.FileOpen(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o666)
	if err != nil {
		return err
	}
	if err := file.Write(stream); err != nil {
		return err
	}
	filePath := piperutils.Path{Name: "aggregated-vulnerabilities", Target: fileName}
	piperutils.PersistReportsAndLinks("whitesourceExecuteScan", "", utils, []piperutils.Path{filePath}, nil)
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
	if err := utils.MkdirAll(ws.ReportsDirectory, 0o777); err != nil {
		return errors.Wrapf(err, "failed to create directories: %s", ws.ReportsDirectory)
	}

	// Write result to file
	fileName := fmt.Sprintf("%s/libraries-%s.csv", ws.ReportsDirectory,
		utils.Now().Format(wsReportTimeStampLayout))
	if err := utils.FileWrite(fileName, []byte(output), 0o666); err != nil {
		return errors.Wrapf(err, "failed to write file: %s", fileName)
	}
	filePath := piperutils.Path{Name: "aggregated-libraries", Target: fileName}
	piperutils.PersistReportsAndLinks("whitesourceExecuteScan", "", utils, []piperutils.Path{filePath}, nil)
	return nil
}

// persistScannedProjects writes all actually scanned WhiteSource project names as list
// into the Common Pipeline Environment, from where it can be used by sub-sequent steps.
func persistScannedProjects(config *ScanOptions, scan *ws.Scan, commonPipelineEnvironment *whitesourceExecuteScanCommonPipelineEnvironment) {
	var projectNames []string
	if config.ProjectName != "" {
		projectNames = []string{config.ProjectName + " - " + config.Version}
	} else {
		projectNames = scan.ScannedProjectNames()
	}
	commonPipelineEnvironment.custom.whitesourceProjectNames = projectNames
}

// create toolrecord file for whitesource
func createToolRecordWhitesource(utils whitesourceUtils, workspace string, config *whitesourceExecuteScanOptions, scan *ws.Scan) (string, error) {
	record := toolrecord.New(utils, workspace, "whitesource", config.ServiceURL)
	// rest api url https://.../api/v1.x
	apiUrl, err := url.Parse(config.ServiceURL)
	if err != nil {
		return "", err
	}
	wsUiRoot := "https://" + apiUrl.Hostname()
	productURL := wsUiRoot + "/Wss/WSS.html#!product;token=" + config.ProductToken
	err = record.AddKeyData("product",
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
		projectId := strconv.FormatInt(project.ID, 10)
		token := project.Token
		projectURL := ""
		if projectId != "" {
			projectURL = wsUiRoot + "/Wss/WSS.html#!project;id=" + projectId
		}
		if token == "" {
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

func downloadMultipleDockerImageAsTar(config *ScanOptions, utils whitesourceUtils) error {

	imageNameToSave := strings.Replace(config.ScanImage, "/", "-", -1)

	saveImageOptions := containerSaveImageOptions{
		ContainerImage:            config.ScanImage,
		ContainerRegistryURL:      config.ScanImageRegistryURL,
		ContainerRegistryUser:     config.ContainerRegistryUser,
		ContainerRegistryPassword: config.ContainerRegistryPassword,
		DockerConfigJSON:          config.DockerConfigJSON,
		FilePath:                  config.ScanPath + "/" + imageNameToSave, // previously was config.ProjectName
		ImageFormat:               "legacy",                                // keep the image format legacy or whitesource is not able to read layers
	}
	dClientOptions := piperDocker.ClientOptions{ImageName: saveImageOptions.ContainerImage, RegistryURL: saveImageOptions.ContainerRegistryURL, LocalPath: "", ImageFormat: "legacy"}
	dClient := &piperDocker.Client{}
	dClient.SetOptions(dClientOptions)
	tarFilePath, err := runContainerSaveImage(&saveImageOptions, &telemetry.CustomData{}, "./cache", "", dClient, utils)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "no image found") {
			log.SetErrorCategory(log.ErrorConfiguration)
		}
		return errors.Wrapf(err, "failed to download Docker image %v", config.ScanImage)
	}
	// remove contents after : in the image name
	if err := renameTarfilePath(tarFilePath); err != nil {
		return errors.Wrapf(err, "failed to rename image %v", err)
	}

	return nil
}

func downloadDockerImageAsTar(config *ScanOptions, utils whitesourceUtils) error {

	saveImageOptions := containerSaveImageOptions{
		ContainerImage:            config.ScanImage,
		ContainerRegistryURL:      config.ScanImageRegistryURL,
		ContainerRegistryUser:     config.ContainerRegistryUser,
		ContainerRegistryPassword: config.ContainerRegistryPassword,
		DockerConfigJSON:          config.DockerConfigJSON,
		FilePath:                  config.ProjectName, // consider changing this to config.ScanPath + "/" + config.ProjectName
		ImageFormat:               "legacy",           // keep the image format legacy or whitesource is not able to read layers
	}
	dClientOptions := piperDocker.ClientOptions{ImageName: saveImageOptions.ContainerImage, RegistryURL: saveImageOptions.ContainerRegistryURL, LocalPath: "", ImageFormat: "legacy"}
	dClient := &piperDocker.Client{}
	dClient.SetOptions(dClientOptions)
	if _, err := runContainerSaveImage(&saveImageOptions, &telemetry.CustomData{}, "./cache", "", dClient, utils); err != nil {
		if strings.Contains(fmt.Sprint(err), "no image found") {
			log.SetErrorCategory(log.ErrorConfiguration)
		}
		return errors.Wrapf(err, "failed to download Docker image %v", config.ScanImage)
	}

	return nil
}

// rename tarFilepath to remove all contents after :
func renameTarfilePath(tarFilepath string) error {
	if _, err := os.Stat(tarFilepath); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", tarFilepath)
	}
	newFileName := ""
	if index := strings.Index(tarFilepath, ":"); index != -1 {
		newFileName = tarFilepath[:index]
		newFileName += ".tar"
	}
	if err := os.Rename(tarFilepath, newFileName); err != nil {
		return fmt.Errorf("error renaming file %s to %s: %v", tarFilepath, newFileName, err)
	}
	return nil
}

// log config parameters
func logConfigInVerboseModeForWhitesource(config ScanOptions) {
	config.ContainerRegistryPassword = "********"
	config.ContainerRegistryUser = "********"
	config.DockerConfigJSON = "********"
	config.OrgToken = "********"
	config.UserToken = "********"
	config.GithubToken = "********"
	config.PrivateModulesGitToken = "********"
	debugLog, _ := json.Marshal(config)
	log.Entry().Debugf("Whitesource configuration: %v", string(debugLog))
}
