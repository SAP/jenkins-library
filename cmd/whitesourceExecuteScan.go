package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/npm"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
	ws "github.com/SAP/jenkins-library/pkg/whitesource"
)

// just to make the lines less long
type ScanOptions = whitesourceExecuteScanOptions

// whitesource defines the functions that are expected by the step implementation to
// be available from the whitesource system.
type whitesource interface {
	GetProductByName(productName string) (ws.Product, error)
	GetProjectsMetaInfo(productToken string) ([]ws.Project, error)
	GetProjectToken(productToken, projectName string) (string, error)
	GetProjectByToken(projectToken string) (ws.Project, error)
	GetProjectRiskReport(projectToken string) ([]byte, error)
	GetProjectVulnerabilityReport(projectToken string, format string) ([]byte, error)
	GetProjectAlerts(projectToken string) ([]ws.Alert, error)
	GetProjectLibraryLocations(projectToken string) ([]ws.Library, error)
}

// wsFile defines the method subset we use from os.File
type wsFile interface {
	io.Writer
	io.StringWriter
	io.Closer
}

type whitesourceUtils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(executable string, params ...string) error

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error

	Chdir(path string) error
	Getwd() (string, error)
	MkdirAll(path string, perm os.FileMode) error
	FileExists(path string) (bool, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRemove(path string) error
	FileRename(oldPath, newPath string) error
	RemoveAll(path string) error
	FileOpen(name string, flag int, perm os.FileMode) (wsFile, error)

	GetArtifactCoordinates(buildTool, buildDescriptorFile string,
		options *versioning.Options) (versioning.Coordinates, error)

	FindPackageJSONFiles(config *ScanOptions) ([]string, error)
	InstallAllNPMDependencies(config *ScanOptions, packageJSONFiles []string) error
}

type whitesourceUtilsBundle struct {
	*piperhttp.Client
	*command.Command
	*piperutils.Files
	npmExecutor npm.Executor
}

func (w *whitesourceUtilsBundle) FileOpen(name string, flag int, perm os.FileMode) (wsFile, error) {
	return os.OpenFile(name, flag, perm)
}

func (w *whitesourceUtilsBundle) GetArtifactCoordinates(buildTool, buildDescriptorFile string,
	options *versioning.Options) (versioning.Coordinates, error) {
	artifact, err := versioning.GetArtifact(buildTool, buildDescriptorFile, options, w)
	if err != nil {
		return nil, err
	}
	return artifact.GetCoordinates()
}

func (w *whitesourceUtilsBundle) getNpmExecutor(config *ScanOptions) npm.Executor {
	if w.npmExecutor == nil {
		w.npmExecutor = npm.NewExecutor(npm.ExecutorOptions{DefaultNpmRegistry: config.DefaultNpmRegistry})
	}
	return w.npmExecutor
}

func (w *whitesourceUtilsBundle) FindPackageJSONFiles(config *ScanOptions) ([]string, error) {
	return w.getNpmExecutor(config).FindPackageJSONFilesWithExcludes(config.BuildDescriptorExcludeList)
}

func (w *whitesourceUtilsBundle) InstallAllNPMDependencies(config *ScanOptions, packageJSONFiles []string) error {
	return w.getNpmExecutor(config).InstallAllDependencies(packageJSONFiles)
}

func newWhitesourceUtils() *whitesourceUtilsBundle {
	utils := whitesourceUtilsBundle{
		Client:  &piperhttp.Client{},
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute cmd output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

// whitesourceScan stores information about scanned projects
type whitesourceScan struct {
	productToken         string
	aggregateProjectName string
	projectVersion       string
	scannedProjects      map[string]ws.Project
	scanTimes            map[string]time.Time
}

func (s *whitesourceScan) init() {
	if s.scannedProjects == nil {
		s.scannedProjects = make(map[string]ws.Project)
	}
	if s.scanTimes == nil {
		s.scanTimes = make(map[string]time.Time)
	}
}

// appendScannedProject checks that no whitesource.Project is already contained in the list of scanned projects,
// and appends a new whitesource.Project with the given name.
func (s *whitesourceScan) appendScannedProject(projectName string) error {
	s.init()
	_, exists := s.scannedProjects[projectName]
	if exists {
		log.Entry().Errorf("A module with the name '%s' was already scanned. "+
			"Your project's modules must have unique names.", projectName)
		return fmt.Errorf("project with name '%s' was already scanned", projectName)
	}
	s.scannedProjects[projectName] = ws.Project{Name: projectName}
	s.scanTimes[projectName] = time.Now()
	return nil
}

func (s *whitesourceScan) updateProjects(sys whitesource) error {
	s.init()
	projects, err := sys.GetProjectsMetaInfo(s.productToken)
	if err != nil {
		return fmt.Errorf("failed to retrieve WhiteSource projects meta info: %w", err)
	}

	var projectsToUpdate []string
	for projectName := range s.scannedProjects {
		projectsToUpdate = append(projectsToUpdate, projectName)
	}

	for _, project := range projects {
		_, exists := s.scannedProjects[project.Name]
		if exists {
			s.scannedProjects[project.Name] = project
			projectsToUpdate, _ = piperutils.RemoveAll(projectsToUpdate, project.Name)
		}
	}
	if len(projectsToUpdate) != 0 {
		log.Entry().Warnf("Could not fetch metadata for projects %v", projectsToUpdate)
	}
	return nil
}

func whitesourceExecuteScan(config ScanOptions, _ *telemetry.CustomData) {
	utils := newWhitesourceUtils()
	scan := whitesourceScan{}
	sys := ws.NewSystem(config.ServiceURL, config.OrgToken, config.UserToken)
	if err := resolveProjectIdentifiers(&config, &scan, utils, sys); err != nil {
		log.Entry().WithError(err).Fatal("step execution failed on resolving project identifiers")
	}

	// Generate a vulnerability report for all projects with version = config.ProjectVersion
	if config.AggregateVersionWideReport {
		if err := aggregateVersionWideLibraries(&config, utils, sys); err != nil {
			log.Entry().WithError(err).Fatal("step execution failed on aggregating version wide libraries")
		}
		if err := aggregateVersionWideVulnerabilities(&config, utils, sys); err != nil {
			log.Entry().WithError(err).Fatal("step execution failed on aggregating version wide vulnerabilities")
		}
	} else {
		if err := runWhitesourceScan(&config, &scan, utils, sys); err != nil {
			log.Entry().WithError(err).Fatal("step execution failed on executing whitesource scan")
		}
	}
}

func runWhitesourceScan(config *ScanOptions, scan *whitesourceScan, utils whitesourceUtils, sys whitesource) error {
	// Start the scan
	if err := executeScan(config, scan, utils); err != nil {
		return err
	}

	// Could perhaps use scan.updateProjects(sys) directly... have not investigated what could break
	if err := resolveProjectIdentifiers(config, scan, utils, sys); err != nil {
		return err
	}

	log.Entry().Info("-----------------------------------------------------")
	log.Entry().Infof("Product Version: '%s'", config.ProductVersion)
	log.Entry().Info("Scanned projects:")
	for _, project := range scan.scannedProjects {
		log.Entry().Infof("  Name: '%s', token: %s", project.Name, project.Token)
	}
	log.Entry().Info("-----------------------------------------------------")

	if config.Reporting || config.SecurityVulnerabilities {
		// Project was scanned. We need to wait for WhiteSource backend to propagate the changes
		// before downloading any reports or check security vulnerabilities.
		if config.ProjectToken != "" {
			// Poll status of aggregated project
			if err := pollProjectStatus(config.ProjectToken, time.Now(), sys); err != nil {
				return err
			}
		} else {
			// Poll status of all scanned projects
			for key, project := range scan.scannedProjects {
				if err := pollProjectStatus(project.Token, scan.scanTimes[key], sys); err != nil {
					return err
				}
			}
		}
	}

	if config.Reporting {
		paths, err := downloadReports(config, utils, sys)
		if err != nil {
			return err
		}
		piperutils.PersistReportsAndLinks("whitesourceExecuteScan", "", nil, paths)
	}

	if config.SecurityVulnerabilities {
		// Check for security vulnerabilities and fail the build if cvssSeverityLimit threshold is crossed
		// convert config.CvssSeverityLimit to float64
		cvssSeverityLimit, err := strconv.ParseFloat(config.CvssSeverityLimit, 64)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("failed to parse parameter cvssSeverityLimit (%s) "+
				"as floating point number: %w", config.CvssSeverityLimit, err)
		}
		if config.ProjectToken != "" {
			project := ws.Project{Name: config.ProjectName, Token: config.ProjectToken}
			if err := checkSecurityViolations(cvssSeverityLimit, project, sys); err != nil {
				return err
			}
		} else {
			for _, project := range scan.scannedProjects {
				if err := checkSecurityViolations(cvssSeverityLimit, project, sys); err != nil {
					return err
				}
			}
		}

	}
	return nil
}

func resolveProjectIdentifiers(config *ScanOptions, scan *whitesourceScan, utils whitesourceUtils, sys whitesource) error {
	if scan.aggregateProjectName == "" {
		scan.aggregateProjectName = config.ProjectName
	}

	if scan.aggregateProjectName == "" || config.ProductVersion == "" {
		options := &versioning.Options{
			ProjectSettingsFile: config.ProjectSettingsFile,
			GlobalSettingsFile:  config.GlobalSettingsFile,
			M2Path:              config.M2Path,
		}
		coordinates, err := utils.GetArtifactCoordinates(config.BuildTool, config.BuildDescriptorFile, options)
		if err != nil {
			return fmt.Errorf("failed to get build artifact description: %w", err)
		}

		nameTmpl := `{{list .GroupID .ArtifactID | join "-" | trimAll "-"}}`
		name, version := versioning.DetermineProjectCoordinates(nameTmpl, config.VersioningModel, coordinates)
		if scan.aggregateProjectName == "" {
			log.Entry().Infof("Resolved project name '%s' from descriptor file", name)
			scan.aggregateProjectName = name
		}
		if config.ProductVersion == "" {
			log.Entry().Infof("Resolved product version '%s' from descriptor file with versioning '%s'",
				version, config.VersioningModel)
			config.ProductVersion = version
		}
	}
	scan.projectVersion = config.ProductVersion

	// Get product token if user did not specify one at runtime
	if config.ProductToken == "" {
		log.Entry().Infof("Attempting to resolve product token for product '%s'..", config.ProductName)
		product, err := sys.GetProductByName(config.ProductName)
		if err != nil {
			return err
		}
		log.Entry().Infof("Resolved product token: '%s'..", product.Token)
		config.ProductToken = product.Token
	}
	scan.productToken = config.ProductToken

	// Get project token if user did not specify one at runtime
	if config.ProjectToken == "" && config.ProjectName != "" {
		log.Entry().Infof("Attempting to resolve project token for project '%s'..", config.ProjectName)
		fullProjName := fmt.Sprintf("%s - %s", config.ProjectName, config.ProductVersion)
		projectToken, err := sys.GetProjectToken(config.ProductToken, fullProjName)
		if err != nil {
			return err
		}
		// A project may not yet exist for this project name-version combo
		// It will be created by the scan, we retrieve the token again after scanning.
		if projectToken != "" {
			log.Entry().Infof("Resolved project token: '%s'..", projectToken)
			config.ProjectToken = projectToken
		} else {
			log.Entry().Infof("Project '%s' not yet present in WhiteSource", fullProjName)
		}
	}

	if err := scan.updateProjects(sys); err != nil {
		return err
	}

	return nil
}

// executeScan executes different types of scans depending on the scanType parameter.
// The default is to download the Unified Agent and use it to perform the scan.
func executeScan(config *ScanOptions, scan *whitesourceScan, utils whitesourceUtils) error {
	if config.ScanType == "" {
		config.ScanType = config.BuildTool
	}

	switch config.ScanType {
	case "mta":
		// Execute scan for maven and all npm modules
		if err := executeMTAScan(config, scan, utils); err != nil {
			return err
		}
	case "maven":
		// Execute scan with maven plugin goal
		if err := executeMavenScan(config, scan, utils); err != nil {
			return err
		}
	case "npm":
		// Execute scan with in each npm module using npm.Executor
		if err := executeNpmScan(config, scan, utils); err != nil {
			return err
		}
	case "yarn":
		// Execute scan with whitesource yarn plugin
		if err := executeYarnScan(config, utils); err != nil {
			return err
		}
	default:
		// Execute scan with Unified Agent jar file
		if err := executeUAScan(config, utils); err != nil {
			return err
		}
	}
	return nil
}

// executeUAScan executes a scan with the Whitesource Unified Agent.
func executeUAScan(config *ScanOptions, utils whitesourceUtils) error {
	// Download the unified agent jar file if one does not exist
	if err := downloadAgent(config, utils); err != nil {
		return err
	}

	// Auto generate a config file based on the working directory's contents.
	// TODO/NOTE: Currently this scans the UA jar file as a dependency since it is downloaded beforehand
	if err := autoGenerateWhitesourceConfig(config, utils); err != nil {
		return err
	}

	return utils.RunExecutable("java", "-jar", config.AgentFileName, "-d", ".", "-c", config.ConfigFilePath,
		"-apiKey", config.OrgToken, "-userKey", config.UserToken, "-project", config.ProjectName,
		"-product", config.ProductName, "-productVersion", config.ProductVersion)
}

// executeMTAScan executes a scan for the Java part with maven, and performs a scan for each NPM module.
func executeMTAScan(config *ScanOptions, scan *whitesourceScan, utils whitesourceUtils) error {
	log.Entry().Infof("Executing Whitesource scan for MTA project")
	err := executeMavenScanForPomFile(config, scan, utils, "pom.xml")
	if err != nil {
		return err
	}

	return executeNpmScan(config, scan, utils)
}

// executeMavenScan constructs maven parameters from the given configuration, and executes the maven goal
// "org.whitesource:whitesource-maven-plugin:19.5.1:update".
func executeMavenScan(config *ScanOptions, scan *whitesourceScan, utils whitesourceUtils) error {
	log.Entry().Infof("Using Whitesource scan for Maven project")
	pomPath := config.BuildDescriptorFile
	if pomPath == "" {
		pomPath = "pom.xml"
	}
	return executeMavenScanForPomFile(config, scan, utils, pomPath)
}

func executeMavenScanForPomFile(config *ScanOptions, scan *whitesourceScan, utils whitesourceUtils, pomPath string) error {
	pomExists, _ := utils.FileExists(pomPath)
	if !pomExists {
		return fmt.Errorf("for scanning with type '%s', the file '%s' must exist in the project root",
			config.ScanType, pomPath)
	}

	defines := []string{
		"-Dorg.whitesource.orgToken=" + config.OrgToken,
		"-Dorg.whitesource.product=" + config.ProductName,
		"-Dorg.whitesource.checkPolicies=true",
		"-Dorg.whitesource.failOnError=true",
	}

	// Aggregate all modules into one WhiteSource project, if user specified the 'projectName' parameter.
	if config.ProjectName != "" {
		defines = append(defines, "-Dorg.whitesource.aggregateProjectName="+config.ProjectName)
		defines = append(defines, "-Dorg.whitesource.aggregateModules=true")
	}

	if config.UserToken != "" {
		defines = append(defines, "-Dorg.whitesource.userKey="+config.UserToken)
	}

	if config.ProductVersion != "" {
		defines = append(defines, "-Dorg.whitesource.productVersion="+config.ProductVersion)
	}

	var flags []string
	excludes := config.BuildDescriptorExcludeList
	if len(excludes) == 0 {
		excludes = []string{
			filepath.Join("unit-tests", "pom.xml"),
			filepath.Join("integration-tests", "pom.xml"),
			filepath.Join("performance-tests", "pom.xml"),
		}
	}
	// From the documentation, these are file paths to a module's pom.xml.
	// For MTA projects, we want to support mixing paths to package.json files and pom.xml files.
	for _, exclude := range excludes {
		if !strings.HasSuffix(exclude, "pom.xml") {
			continue
		}
		moduleName := filepath.Dir(exclude)
		if moduleName != "" {
			flags = append(flags, "-pl", "!"+moduleName)
		}
	}

	err := maven.VisitAllMavenModules(".", utils, excludes, func(info maven.ModuleInfo) error {
		project := info.Project
		if project.Packaging != "pom" {
			if project.ArtifactId == "" {
				return fmt.Errorf("artifactId missing from '%s'", info.PomXMLPath)
			}

			err := scan.appendScannedProject(project.ArtifactId + " - " + scan.projectVersion)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	_, err = maven.Execute(&maven.ExecuteOptions{
		PomPath:             pomPath,
		M2Path:              config.M2Path,
		GlobalSettingsFile:  config.GlobalSettingsFile,
		ProjectSettingsFile: config.ProjectSettingsFile,
		Defines:             defines,
		Flags:               flags,
		Goals:               []string{"org.whitesource:whitesource-maven-plugin:19.5.1:update"},
	}, utils)

	return err
}

const whiteSourceConfig = "whitesource.config.json"

func setValueAndLogChange(config map[string]interface{}, key string, value interface{}) {
	oldValue, exists := config[key]
	if exists && oldValue != value {
		log.Entry().Infof("overwriting '%s' in %s: %v -> %v", key, whiteSourceConfig, oldValue, value)
	}
	config[key] = value
}

func setValueOmitIfPresent(config map[string]interface{}, key, omitIfPresent string, value interface{}) {
	_, exists := config[omitIfPresent]
	if exists {
		return
	}
	setValueAndLogChange(config, key, value)
}

func writeWhitesourceConfigJSON(config *ScanOptions, utils whitesourceUtils, devDep, ignoreLsErrors bool) error {
	var npmConfig = make(map[string]interface{})

	exists, _ := utils.FileExists(whiteSourceConfig)
	if exists {
		fileContents, err := utils.FileRead(whiteSourceConfig)
		if err != nil {
			return fmt.Errorf("file '%s' already exists, but could not be read: %w", whiteSourceConfig, err)
		}
		err = json.Unmarshal(fileContents, &npmConfig)
		if err != nil {
			return fmt.Errorf("file '%s' already exists, but could not be parsed: %w", whiteSourceConfig, err)
		}
		log.Entry().Infof("The file '%s' already exists in the project. Changed config details will be logged.",
			whiteSourceConfig)
	}

	npmConfig["apiKey"] = config.OrgToken
	npmConfig["userKey"] = config.UserToken
	setValueAndLogChange(npmConfig, "checkPolicies", true)
	setValueAndLogChange(npmConfig, "productName", config.ProductName)
	setValueAndLogChange(npmConfig, "productVer", config.ProductVersion)
	setValueOmitIfPresent(npmConfig, "productToken", "projectToken", config.ProductToken)
	if config.ProjectName != "" {
		// In case there are other modules (i.e. maven modules in MTA projects),
		// or more than one NPM module, setting the project name will lead to
		// overwriting any previous scan results with the one from this module!
		// If this is not provided, the WhiteSource project name will be generated
		// from "name" in package.json plus " - " plus productVersion.
		setValueAndLogChange(npmConfig, "projectName", config.ProjectName)
	}
	setValueAndLogChange(npmConfig, "devDep", devDep)
	setValueAndLogChange(npmConfig, "ignoreNpmLsErrors", ignoreLsErrors)

	jsonBuffer, err := json.Marshal(npmConfig)
	if err != nil {
		return fmt.Errorf("failed to generate '%s': %w", whiteSourceConfig, err)
	}

	err = utils.FileWrite(whiteSourceConfig, jsonBuffer, 0644)
	if err != nil {
		return fmt.Errorf("failed to write '%s': %w", whiteSourceConfig, err)
	}
	return nil
}

// executeNpmScan iterates over all found npm modules and performs a scan in each one.
func executeNpmScan(config *ScanOptions, scan *whitesourceScan, utils whitesourceUtils) error {
	modules, err := utils.FindPackageJSONFiles(config)
	if err != nil {
		return fmt.Errorf("failed to find package.json files with excludes: %w", err)
	}
	if len(modules) == 0 {
		return fmt.Errorf("found no NPM modules to scan. Configured excludes: %v",
			config.BuildDescriptorExcludeList)
	}
	for _, module := range modules {
		err := executeNpmScanForModule(module, config, scan, utils)
		if err != nil {
			return fmt.Errorf("failed to scan NPM module '%s': %w", module, err)
		}
	}
	return nil
}

// executeNpmScanForModule generates a configuration file whitesource.config.json with appropriate values from config,
// installs all dependencies if necessary, and executes the scan via "npx whitesource run".
func executeNpmScanForModule(modulePath string, config *ScanOptions, scan *whitesourceScan, utils whitesourceUtils) error {
	log.Entry().Infof("Executing Whitesource scan for NPM module '%s'", modulePath)

	resetDir, err := utils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to obtain current directory: %w", err)
	}

	dir := filepath.Dir(modulePath)
	if err := utils.Chdir(dir); err != nil {
		return fmt.Errorf("failed to change into directory '%s': %w", dir, err)
	}
	defer func() {
		err = utils.Chdir(resetDir)
		if err != nil {
			log.Entry().Errorf("Failed to reset into directory '%s': %v", resetDir, err)
		}
	}()

	if err := writeWhitesourceConfigJSON(config, utils, false, true); err != nil {
		return err
	}
	defer func() { _ = utils.FileRemove(whiteSourceConfig) }()

	projectName, err := getNpmProjectName(modulePath, utils)
	if err != nil {
		return err
	}

	if err := reinstallNodeModulesIfLsFails(modulePath, config, utils); err != nil {
		return err
	}

	projectName = projectName + " - " + config.ProductVersion
	if err := scan.appendScannedProject(projectName); err != nil {
		return err
	}

	return utils.RunExecutable("npx", "whitesource", "run")
}

func getNpmProjectName(modulePath string, utils whitesourceUtils) (string, error) {
	fileContents, err := utils.FileRead("package.json")
	if err != nil {
		return "", fmt.Errorf("could not read package.json: %w", err)
	}
	var packageJson = make(map[string]interface{})
	err = json.Unmarshal(fileContents, &packageJson)

	projectNameEntry, exists := packageJson["name"]
	if !exists {
		return "", fmt.Errorf("the file '%s' must configure a name",
			filepath.Join(modulePath, "package.json"))
	}

	projectName, isString := projectNameEntry.(string)
	if !isString {
		return "", fmt.Errorf("the file '%s' must configure a name",
			filepath.Join(modulePath, "package.json"))
	}

	return projectName, nil
}

func reinstallNodeModulesIfLsFails(modulePath string, config *ScanOptions, utils whitesourceUtils) error {
	// No need to have output from "npm ls" in the log
	utils.Stdout(ioutil.Discard)
	defer utils.Stdout(log.Writer())

	err := utils.RunExecutable("npm", "ls")
	if err == nil {
		return nil
	}
	log.Entry().Warnf("'npm ls' failed. Re-installing NPM Node Modules")
	err = utils.RemoveAll("node_modules")
	if err != nil {
		return fmt.Errorf("failed to remove node_modules directory: %w", err)
	}
	err = utils.MkdirAll("node_modules", os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to recreate node_modules directory: %w", err)
	}
	exists, _ := utils.FileExists("package-lock.json")
	if exists {
		err = utils.FileRemove("package-lock.json")
		if err != nil {
			return fmt.Errorf("failed to remove package-lock.json: %w", err)
		}
	}
	return utils.InstallAllNPMDependencies(config, []string{modulePath})
}

// executeYarnScan generates a configuration file whitesource.config.json with appropriate values from config,
// installs whitesource yarn plugin and executes the scan.
func executeYarnScan(config *ScanOptions, utils whitesourceUtils) error {
	if err := writeWhitesourceConfigJSON(config, utils, true, false); err != nil {
		return err
	}
	defer func() { _ = utils.FileRemove(whiteSourceConfig) }()
	if err := utils.RunExecutable("yarn", "global", "add", "whitesource"); err != nil {
		return err
	}
	if err := utils.RunExecutable("yarn", "install"); err != nil {
		return err
	}
	if err := utils.RunExecutable("whitesource", "yarn"); err != nil {
		return err
	}
	return nil
}

// checkSecurityViolations checks security violations and returns an error if the configured severity limit is crossed.
func checkSecurityViolations(cvssSeverityLimit float64, project ws.Project, sys whitesource) error {
	// get project alerts (vulnerabilities)
	alerts, err := sys.GetProjectAlerts(project.Token)
	if err != nil {
		return fmt.Errorf("failed to retrieve project alerts from Whitesource: %w", err)
	}

	severeVulnerabilities := 0
	// https://github.com/SAP/jenkins-library/blob/master/vars/whitesourceExecuteScan.groovy#L537
	for _, alert := range alerts {
		vuln := alert.Vulnerability
		if (vuln.Score >= cvssSeverityLimit || vuln.CVSS3Score >= cvssSeverityLimit) && cvssSeverityLimit >= 0 {
			log.Entry().Infof("Vulnerability with Score %v / CVSS3Score %v treated as severe",
				vuln.Score, vuln.CVSS3Score)
			severeVulnerabilities++
		} else {
			log.Entry().Infof("Ignoring vulnerability with Score %v / CVSS3Score %v",
				vuln.Score, vuln.CVSS3Score)
		}
	}

	//https://github.com/SAP/jenkins-library/blob/master/vars/whitesourceExecuteScan.groovy#L547
	nonSevereVulnerabilities := len(alerts) - severeVulnerabilities
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
		return fmt.Errorf("%v Open Source Software Security vulnerabilities with CVSS score greater "+
			"or equal to %.1f detected in project %s",
			severeVulnerabilities, cvssSeverityLimit, project.Name)
	}
	return nil
}

// pollProjectStatus polls project LastUpdateDate until it reflects the most recent scan
func pollProjectStatus(projectToken string, scanTime time.Time, sys whitesource) error {
	return blockUntilProjectIsUpdated(projectToken, sys, scanTime, 20*time.Second, 20*time.Second, 15*time.Minute)
}

const whitesourceDateTimeLayout = "2006-01-02 15:04:05 -0700"

// blockUntilProjectIsUpdated polls the project LastUpdateDate until it is newer than the given time stamp
// or no older than maxAge relative to the given time stamp.
func blockUntilProjectIsUpdated(projectToken string, sys whitesource, currentTime time.Time, maxAge, timeBetweenPolls, maxWaitTime time.Duration) error {
	startTime := time.Now()
	for {
		project, err := sys.GetProjectByToken(projectToken)
		if err != nil {
			return err
		}

		if project.LastUpdateDate == "" {
			log.Entry().Infof("last updated time missing from project metadata, retrying")
		} else {
			lastUpdatedTime, err := time.Parse(whitesourceDateTimeLayout, project.LastUpdateDate)
			if err != nil {
				return fmt.Errorf("failed to parse last updated time (%s) of Whitesource project: %w",
					project.LastUpdateDate, err)
			}
			age := currentTime.Sub(lastUpdatedTime)
			if age < maxAge {
				//done polling
				break
			}
			log.Entry().Infof("time since project was last updated %v > %v, polling status...", age, maxAge)
		}

		if time.Now().Sub(startTime) > maxWaitTime {
			return fmt.Errorf("timeout while waiting for Whitesource scan results to be reflected in service")
		}

		time.Sleep(timeBetweenPolls)
	}
	return nil
}

// downloadReports downloads a project's risk and vulnerability reports
func downloadReports(config *ScanOptions, utils whitesourceUtils, sys whitesource) ([]piperutils.Path, error) {
	if err := utils.MkdirAll(config.ReportDirectoryName, os.ModePerm); err != nil {
		return nil, err
	}
	vulnPath, err := downloadVulnerabilityReport(config, utils, sys)
	if err != nil {
		return nil, err
	}
	riskPath, err := downloadRiskReport(config, utils, sys)
	if err != nil {
		return nil, err
	}
	return []piperutils.Path{*vulnPath, *riskPath}, nil
}

func downloadVulnerabilityReport(config *ScanOptions, utils whitesourceUtils, sys whitesource) (*piperutils.Path, error) {
	reportBytes, err := sys.GetProjectVulnerabilityReport(config.ProjectToken, config.VulnerabilityReportFormat)
	if err != nil {
		return nil, err
	}

	// Write report to file
	rptFileName := fmt.Sprintf("%s-vulnerability-report.%s", config.ProjectName, config.VulnerabilityReportFormat)
	rptFileName = filepath.Join(config.ReportDirectoryName, rptFileName)
	if err := utils.FileWrite(rptFileName, reportBytes, 0644); err != nil {
		return nil, err
	}

	log.Entry().Infof("Successfully downloaded vulnerability report to %s", rptFileName)
	pathName := fmt.Sprintf("%s Vulnerability Report", config.ProjectName)
	return &piperutils.Path{Name: pathName, Target: rptFileName}, nil
}

func downloadRiskReport(config *ScanOptions, utils whitesourceUtils, sys whitesource) (*piperutils.Path, error) {
	reportBytes, err := sys.GetProjectRiskReport(config.ProjectToken)
	if err != nil {
		return nil, err
	}

	rptFileName := fmt.Sprintf("%s-risk-report.pdf", config.ProjectName)
	rptFileName = filepath.Join(config.ReportDirectoryName, rptFileName)
	if err := utils.FileWrite(rptFileName, reportBytes, 0644); err != nil {
		return nil, err
	}

	log.Entry().Infof("Successfully downloaded risk report to %s", rptFileName)
	pathName := fmt.Sprintf("%s PDF Risk Report", config.ProjectName)
	return &piperutils.Path{Name: pathName, Target: rptFileName}, nil
}

// downloadAgent downloads the unified agent jar file if one does not exist
func downloadAgent(config *ScanOptions, utils whitesourceUtils) error {
	agentFile := config.AgentFileName
	exists, err := utils.FileExists(agentFile)
	if err != nil {
		return fmt.Errorf("could not check whether the file '%s' exists: %w", agentFile, err)
	}
	if !exists {
		err := utils.DownloadFile(config.AgentDownloadURL, agentFile, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to download unified agent from URL '%s' to file '%s': %w",
				config.AgentDownloadURL, agentFile, err)
		}
	}
	return nil
}

// autoGenerateWhitesourceConfig
// Auto generate a config file based on the current directory structure, renames it to user specified configFilePath
// Generated file name will be 'wss-generated-file.config'
func autoGenerateWhitesourceConfig(config *ScanOptions, utils whitesourceUtils) error {
	// TODO: Should we rely on -detect, or set the parameters manually?
	if err := utils.RunExecutable("java", "-jar", config.AgentFileName, "-d", ".", "-detect"); err != nil {
		return err
	}

	// Rename generated config file to config.ConfigFilePath parameter
	if err := utils.FileRename("wss-generated-file.config", config.ConfigFilePath); err != nil {
		return err
	}

	// Append aggregateModules=true parameter to config file (consolidates multi-module projects into one)
	f, err := utils.FileOpen(config.ConfigFilePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	// Append additional config parameters to prevent multiple projects being generated
	m2Path := config.M2Path
	if m2Path == "" {
		m2Path = ".m2"
	}
	cfg := fmt.Sprintf("\ngradle.aggregateModules=true\nmaven.aggregateModules=true\ngradle.localRepositoryPath=.gradle\nmaven.m2RepositoryPath=%s\nexcludes=%s",
		m2Path,
		config.Excludes)
	if _, err = f.WriteString(cfg); err != nil {
		return err
	}

	// archiveExtractionDepth=0
	if err := utils.RunExecutable("sed", "-ir", `s/^[#]*\s*archiveExtractionDepth=.*/archiveExtractionDepth=0/`,
		config.ConfigFilePath); err != nil {
		return err
	}

	// config.Includes defaults to "**/*.java **/*.jar **/*.py **/*.go **/*.js **/*.ts"
	regex := fmt.Sprintf(`s/^[#]*\s*includes=.*/includes="%s"/`, config.Includes)
	if err := utils.RunExecutable("sed", "-ir", regex, config.ConfigFilePath); err != nil {
		return err
	}

	return nil
}

func aggregateVersionWideLibraries(config *ScanOptions, utils whitesourceUtils, sys whitesource) error {
	log.Entry().Infof("Aggregating list of libraries used for all projects with version: %s", config.ProductVersion)

	projects, err := sys.GetProjectsMetaInfo(config.ProductToken)
	if err != nil {
		return err
	}

	versionWideLibraries := map[string][]ws.Library{} // maps project name to slice of libraries
	for _, project := range projects {
		projectVersion := strings.Split(project.Name, " - ")[1]
		projectName := strings.Split(project.Name, " - ")[0]
		if projectVersion == config.ProductVersion {
			libs, err := sys.GetProjectLibraryLocations(project.Token)
			if err != nil {
				return err
			}
			log.Entry().Infof("Found project: %s with %v libraries.", project.Name, len(libs))
			versionWideLibraries[projectName] = libs
		}
	}
	if err := newLibraryCSVReport(versionWideLibraries, config, utils); err != nil {
		return err
	}
	return nil
}

func aggregateVersionWideVulnerabilities(config *ScanOptions, utils whitesourceUtils, sys whitesource) error {
	log.Entry().Infof("Aggregating list of vulnerabilities for all projects with version: %s", config.ProductVersion)

	projects, err := sys.GetProjectsMetaInfo(config.ProductToken)
	if err != nil {
		return err
	}

	var versionWideAlerts []ws.Alert // all alerts for a given project version
	projectNames := ``               // holds all project tokens considered a part of the report for debugging
	for _, project := range projects {
		projectVersion := strings.Split(project.Name, " - ")[1]
		if projectVersion == config.ProductVersion {
			projectNames += project.Name + "\n"
			alerts, err := sys.GetProjectAlerts(project.Token)
			if err != nil {
				return err
			}
			log.Entry().Infof("Found project: %s with %v vulnerabilities.", project.Name, len(alerts))
			versionWideAlerts = append(versionWideAlerts, alerts...)
		}
	}

	if err := utils.FileWrite("whitesource-reports/project-names-aggregated.txt", []byte(projectNames), 0777); err != nil {
		return err
	}
	if err := newVulnerabilityExcelReport(versionWideAlerts, config, utils); err != nil {
		return err
	}
	return nil
}

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
	if err := streamWriter.SetRow("A1", []interface{}{excelize.Cell{StyleID: styleID, Value: "Severity"}}); err != nil {
		return err
	}
	if err := streamWriter.SetRow("B1", []interface{}{excelize.Cell{StyleID: styleID, Value: "Library"}}); err != nil {
		return err
	}
	if err := streamWriter.SetRow("C1", []interface{}{excelize.Cell{StyleID: styleID, Value: "Vulnerability ID"}}); err != nil {
		return err
	}
	if err := streamWriter.SetRow("D1", []interface{}{excelize.Cell{StyleID: styleID, Value: "Project"}}); err != nil {
		return err
	}
	if err := streamWriter.SetRow("E1", []interface{}{excelize.Cell{StyleID: styleID, Value: "Resolution"}}); err != nil {
		return err
	}

	for i, alert := range alerts {
		row := make([]interface{}, 5)
		vuln := alert.Vulnerability
		row[0] = vuln.Severity
		row[1] = alert.Library.Filename
		row[2] = vuln.Level
		row[3] = alert.Project
		row[4] = vuln.FixResolutionText
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		if err := streamWriter.SetRow(cell, row); err != nil {
			fmt.Println(err)
		}
	}
	if err := streamWriter.Flush(); err != nil {
		return err
	}

	if err := utils.MkdirAll(config.ReportDirectoryName, 0777); err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s/vulnerabilities-%s.xlsx", config.ReportDirectoryName, time.Now().Format("2006-01-01 15:00:00"))
	stream, err := utils.FileOpen(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	if err := file.Write(stream); err != nil {
		return err
	}
	return nil
}

// outputs an slice of libraries to an excel file based on projects with version == config.ProductVersion
func newLibraryCSVReport(libraries map[string][]ws.Library, config *ScanOptions, utils whitesourceUtils) error {
	output := "Library Name, Project Name\n"
	for projectName, libraries := range libraries {
		log.Entry().Infof("Writing %v libraries for project %s to excel report..", len(libraries), projectName)
		for _, library := range libraries {
			output += library.Name + ", " + projectName + "\n"
		}
	}

	// Ensure reporting directory exists
	if err := utils.MkdirAll(config.ReportDirectoryName, 0777); err != nil {
		return err
	}

	// Write result to file
	fileName := fmt.Sprintf("%s/libraries-%s.csv", config.ReportDirectoryName, time.Now().Format("2006-01-01 15:00:00"))
	if err := utils.FileWrite(fileName, []byte(output), 0777); err != nil {
		return err
	}
	return nil
}
