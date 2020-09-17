package cmd

import (
	"encoding/json"
	"fmt"
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
	FileOpen(name string, flag int, perm os.FileMode) (*os.File, error)

	GetArtifactCoordinates(config *ScanOptions) (versioning.Coordinates, error)
}

type whitesourceUtilsBundle struct {
	*piperhttp.Client
	*command.Command
	*piperutils.Files
}

func (w *whitesourceUtilsBundle) GetArtifactCoordinates(config *ScanOptions) (versioning.Coordinates, error) {
	opts := &versioning.Options{
		ProjectSettingsFile: config.ProjectSettingsFile,
		GlobalSettingsFile:  config.GlobalSettingsFile,
		M2Path:              config.M2Path,
	}
	artifact, err := versioning.GetArtifact(config.BuildTool, config.BuildDescriptorFile, opts, w)
	if err != nil {
		return nil, err
	}
	return artifact.GetCoordinates()
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

func whitesourceExecuteScan(config ScanOptions, _ *telemetry.CustomData) {
	utils := newWhitesourceUtils()
	sys := ws.NewSystem(config.ServiceURL, config.OrgToken, config.UserToken)
	if err := resolveProjectIdentifiers(&config, utils, sys); err != nil {
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
		if err := runWhitesourceScan(&config, utils, sys); err != nil {
			log.Entry().WithError(err).Fatal("step execution failed on executing whitesource scan")
		}
	}
}

func runWhitesourceScan(config *ScanOptions, utils whitesourceUtils, sys whitesource) error {
	// Start the scan
	if err := executeScan(config, utils); err != nil {
		return err
	}

	// Scan finished: we need to resolve project token again if the project was just created.
	if err := resolveProjectIdentifiers(config, utils, sys); err != nil {
		return err
	}

	log.Entry().Info("-----------------------------------------------------")
	log.Entry().Infof("Project name: '%s'", config.ProjectName)
	log.Entry().Infof("Product Version: '%s'", config.ProductVersion)
	log.Entry().Infof("Project Token: %s", config.ProjectToken)
	log.Entry().Info("-----------------------------------------------------")

	if config.Reporting || config.SecurityVulnerabilities {
		// Project was scanned. We need to wait for WhiteSource backend to propagate the changes
		// before downloading any reports or check security vulnerabilities.
		if err := pollProjectStatus(config, sys); err != nil {
			return err
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
		if err := checkSecurityViolations(config, sys); err != nil {
			return err
		}
	}
	return nil
}

func resolveProjectIdentifiers(config *ScanOptions, utils whitesourceUtils, sys whitesource) error {
	if config.ProjectName == "" || config.ProductVersion == "" {
		coordinates, err := utils.GetArtifactCoordinates(config)
		if err != nil {
			return fmt.Errorf("failed to get build artifact description: %w", err)
		}

		nameTmpl := `{{list .GroupID .ArtifactID | join "-" | trimAll "-"}}`
		name, version := versioning.DetermineProjectCoordinates(nameTmpl, config.VersioningModel, coordinates)
		if config.ProjectName == "" {
			log.Entry().Infof("Resolved project name '%s' from descriptor file", name)
			config.ProjectName = name
		}
		if config.ProductVersion == "" {
			log.Entry().Infof("Resolved product version '%s' from descriptor file with versioning '%s'",
				version, config.VersioningModel)
			config.ProductVersion = version
		}
	}

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

	// Get project token  if user did not specify one at runtime
	if config.ProjectToken == "" {
		log.Entry().Infof("Attempting to resolve project token for project '%s'..", config.ProjectName)
		fullProjName := fmt.Sprintf("%s - %s", config.ProjectName, config.ProductVersion)
		projectToken, err := sys.GetProjectToken(config.ProductToken, fullProjName)
		if err != nil {
			return err
		}
		if projectToken == "" {
			return fmt.Errorf("failed to resolve project token for '%s' and product token %s",
				config.ProjectName, config.ProductToken)
		}
		log.Entry().Infof("Resolved project token: '%s'..", projectToken)
		config.ProjectToken = projectToken
	}
	return nil
}

// executeScan executes different types of scans depending on the scanType parameter.
// The default is to download the Unified Agent and use it to perform the scan.
func executeScan(config *ScanOptions, utils whitesourceUtils) error {
	if config.ScanType == "" {
		config.ScanType = config.BuildTool
	}

	switch config.ScanType {
	case "npm":
		// Execute scan with whitesource yarn plugin
		if err := executeYarnScan(config, utils); err != nil {
			return err
		}
	default:
		// Download the unified agent jar file if one does not exist
		if err := downloadAgent(config, utils); err != nil {
			return err
		}

		// Auto generate a config file based on the working directory's contents.
		// TODO/NOTE: Currently this scans the UA jar file as a dependency since it is downloaded beforehand
		if err := autoGenerateWhitesourceConfig(config, utils); err != nil {
			return err
		}

		// Execute whitesource scan with unified agent jar file
		if err := executeUAScan(config, utils); err != nil {
			return err
		}
	}
	return nil
}

// executeUAScan executes a scan with the Whitesource Unified Agent.
func executeUAScan(config *ScanOptions, utils whitesourceUtils) error {
	return utils.RunExecutable("java", "-jar", config.AgentFileName, "-d", ".", "-c", config.ConfigFilePath,
		"-apiKey", config.OrgToken, "-userKey", config.UserToken, "-project", config.ProjectName,
		"-product", config.ProductName, "-productVersion", config.ProductVersion)
}

const whiteSourceConfig = "whitesource.config.json"

func setValueAndLogChange(config map[string]interface{}, key string, value interface{}) {
	oldValue, exists := config[key]
	if exists && oldValue != value {
		log.Entry().Infof("overwriting '%s' in %s: %v -> %v", key, whiteSourceConfig, oldValue, value)
	}
	config[key] = value
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
	setValueAndLogChange(npmConfig, "projectName", config.ProjectName)
	setValueAndLogChange(npmConfig, "productVer", config.ProductVersion)
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
func checkSecurityViolations(config *ScanOptions, sys whitesource) error {
	// convert config.CvssSeverityLimit to float64
	cvssSeverityLimit, err := strconv.ParseFloat(config.CvssSeverityLimit, 64)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("failed to parse parameter cvssSeverityLimit (%s) "+
			"as floating point number: %w", config.CvssSeverityLimit, err)
	}

	// get project alerts (vulnerabilities)
	alerts, err := sys.GetProjectAlerts(config.ProjectToken)
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
			"CVSS score below threshold %s detected in project %s.", nonSevereVulnerabilities,
			config.CvssSeverityLimit, config.ProjectName)
	} else if len(alerts) == 0 {
		log.Entry().Infof("No Open Source Software Security vulnerabilities detected in project %s",
			config.ProjectName)
	}

	// https://github.com/SAP/jenkins-library/blob/master/vars/whitesourceExecuteScan.groovy#L558
	if severeVulnerabilities > 0 {
		return fmt.Errorf("%v Open Source Software Security vulnerabilities with CVSS score greater "+
			"or equal to %s detected in project %s",
			severeVulnerabilities, config.CvssSeverityLimit, config.ProjectName)
	}
	return nil
}

// pollProjectStatus polls project LastUpdateDate until it reflects the most recent scan
func pollProjectStatus(config *ScanOptions, sys whitesource) error {
	return blockUntilProjectIsUpdated(config, sys, time.Now(), 20*time.Second, 20*time.Second, 15*time.Minute)
}

const whitesourceDateTimeLayout = "2006-01-02 15:04:05 -0700"

// blockUntilProjectIsUpdated polls the project LastUpdateDate until it is newer than the given time stamp
// or no older than maxAge relative to the given time stamp.
func blockUntilProjectIsUpdated(config *ScanOptions, sys whitesource, currentTime time.Time, maxAge, timeBetweenPolls, maxWaitTime time.Duration) error {
	startTime := time.Now()
	for {
		project, err := sys.GetProjectByToken(config.ProjectToken)
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
	if !fileExists(agentFile) {
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
	cfg := fmt.Sprintf("gradle.aggregateModules=true\nmaven.aggregateModules=true\ngradle.localRepositoryPath=.gradle\nmaven.m2RepositoryPath=.m2\nexcludes=%s", config.Excludes)
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

	if err := ioutil.WriteFile("whitesource-reports/project-names-aggregated.txt", []byte(projectNames), 0777); err != nil {
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
	if err := file.SaveAs(fileName); err != nil {
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
	if err := ioutil.WriteFile(fileName, []byte(output), 0777); err != nil {
		return err
	}
	return nil
}
