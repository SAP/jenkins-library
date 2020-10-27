package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/npm"
	"os"
	"path/filepath"
	"sort"
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
	ws.Utils

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

func (w *whitesourceUtilsBundle) GetArtifactCoordinates(buildTool, buildDescriptorFile string,
	options *versioning.Options) (versioning.Coordinates, error) {
	artifact, err := versioning.GetArtifact(buildTool, buildDescriptorFile, options, w)
	if err != nil {
		return nil, err
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

func (w *whitesourceUtilsBundle) Now() time.Time {
	return time.Now()
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

func newWhitesourceScan(config *ScanOptions) *ws.Scan {
	return &ws.Scan{
		AggregateProjectName: config.ProjectName,
		ProductVersion:       config.ProductVersion,
	}
}

func whitesourceExecuteScan(config ScanOptions, _ *telemetry.CustomData) {
	utils := newWhitesourceUtils()
	scan := newWhitesourceScan(&config)
	sys := ws.NewSystem(config.ServiceURL, config.OrgToken, config.UserToken)
	err := runWhitesourceExecuteScan(&config, scan, utils, sys)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runWhitesourceExecuteScan(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils, sys whitesource) error {
	if err := resolveProjectIdentifiers(config, scan, utils, sys); err != nil {
		return fmt.Errorf("failed to resolve project identifiers: %w", err)
	}

	if config.AggregateVersionWideReport {
		// Generate a vulnerability report for all projects with version = config.ProjectVersion
		// Note that this is not guaranteed that all projects are from the same scan.
		// For example, if a module was removed from the source code, the project may still
		// exist in the WhiteSource system.
		if err := aggregateVersionWideLibraries(config, utils, sys); err != nil {
			return fmt.Errorf("failed to aggregate version wide libraries: %w", err)
		}
		if err := aggregateVersionWideVulnerabilities(config, utils, sys); err != nil {
			return fmt.Errorf("failed to aggregate version wide vulnerabilities: %w", err)
		}
	} else {
		if err := runWhitesourceScan(config, scan, utils, sys); err != nil {
			return fmt.Errorf("failed to execute WhiteSource scan: %w", err)
		}
	}
	return nil
}

func runWhitesourceScan(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils, sys whitesource) error {
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
	for _, project := range scan.ScannedProjects() {
		log.Entry().Infof("  Name: '%s', token: %s", project.Name, project.Token)
	}
	log.Entry().Info("-----------------------------------------------------")

	if err := checkAndReportScanResults(config, scan, utils, sys); err != nil {
		return err
	}

	if err := persistScannedProjects(config, scan, utils); err != nil {
		return fmt.Errorf("failed to persist scanned WhiteSource project names: %w", err)
	}

	return nil
}

func checkAndReportScanResults(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils, sys whitesource) error {
	if !config.Reporting && !config.SecurityVulnerabilities {
		return nil
	}
	if err := blockUntilReportsAreaReady(config, scan, sys); err != nil {
		return err
	}
	if config.Reporting {
		paths, err := scan.DownloadReports(ws.ReportOptions{
			ReportDirectory:           config.ReportDirectoryName,
			VulnerabilityReportFormat: config.VulnerabilityReportFormat,
		}, utils, sys)
		if err != nil {
			return err
		}
		piperutils.PersistReportsAndLinks("whitesourceExecuteScan", "", nil, paths)
	}
	if config.SecurityVulnerabilities {
		if err := checkSecurityViolations(config, scan, sys); err != nil {
			return err
		}
	}
	return nil
}

func resolveProjectIdentifiers(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils, sys whitesource) error {
	if scan.AggregateProjectName == "" || config.ProductVersion == "" {
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
		if scan.AggregateProjectName == "" {
			log.Entry().Infof("Resolved project name '%s' from descriptor file", name)
			scan.AggregateProjectName = name
		}
		if config.ProductVersion == "" {
			log.Entry().Infof("Resolved product version '%s' from descriptor file with versioning '%s'",
				version, config.VersioningModel)
			config.ProductVersion = version
		}
	}
	scan.ProductVersion = validateProductVersion(config.ProductVersion)

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

	return scan.UpdateProjects(config.ProductToken, sys)
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
		ScanType:                   config.ScanType,
		OrgToken:                   config.OrgToken,
		UserToken:                  config.UserToken,
		ProductName:                config.ProductName,
		ProductToken:               config.ProductToken,
		ProjectName:                config.ProjectName,
		BuildDescriptorExcludeList: config.BuildDescriptorExcludeList,
		PomPath:                    config.BuildDescriptorFile,
		M2Path:                     config.M2Path,
		GlobalSettingsFile:         config.GlobalSettingsFile,
		ProjectSettingsFile:        config.ProjectSettingsFile,
		DefaultNpmRegistry:         config.DefaultNpmRegistry,
		AgentDownloadURL:           config.AgentDownloadURL,
		AgentFileName:              config.AgentFileName,
		ConfigFilePath:             config.ConfigFilePath,
		Includes:                   config.Includes,
		Excludes:                   config.Excludes,
	}
}

// executeScan executes different types of scans depending on the scanType parameter.
// The default is to download the Unified Agent and use it to perform the scan.
func executeScan(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils) error {
	if config.ScanType == "" {
		config.ScanType = config.BuildTool
	}

	options := wsScanOptions(config)

	switch config.ScanType {
	case "mta":
		// Execute scan for maven and all npm modules
		if err := scan.ExecuteMTAScan(options, utils); err != nil {
			return err
		}
	case "maven":
		// Execute scan with maven plugin goal
		if err := scan.ExecuteMavenScan(options, utils); err != nil {
			return err
		}
	case "npm":
		// Execute scan with in each npm module using npm.Executor
		if err := scan.ExecuteNpmScan(options, utils); err != nil {
			return err
		}
	case "yarn":
		// Execute scan with whitesource yarn plugin
		if err := scan.ExecuteYarnScan(options, utils); err != nil {
			return err
		}
	default:
		// Execute scan with Unified Agent jar file
		if err := scan.ExecuteUAScan(options, utils); err != nil {
			return err
		}
	}
	return nil
}

func checkSecurityViolations(config *ScanOptions, scan *ws.Scan, sys whitesource) error {
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
		if err := checkProjectSecurityViolations(cvssSeverityLimit, project, sys); err != nil {
			return err
		}
	} else {
		for _, project := range scan.ScannedProjects() {
			if err := checkProjectSecurityViolations(cvssSeverityLimit, project, sys); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkSecurityViolations checks security violations and returns an error if the configured severity limit is crossed.
func checkProjectSecurityViolations(cvssSeverityLimit float64, project ws.Project, sys whitesource) error {
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

func blockUntilReportsAreaReady(config *ScanOptions, scan *ws.Scan, sys whitesource) error {
	// Project was scanned. We need to wait for WhiteSource backend to propagate the changes
	// before downloading any reports or check security vulnerabilities.
	if config.ProjectToken != "" {
		// Poll status of aggregated project
		if err := pollProjectStatus(config.ProjectToken, time.Now(), sys); err != nil {
			return err
		}
	} else {
		// Poll status of all scanned projects
		for _, project := range scan.ScannedProjects() {
			if err := pollProjectStatus(project.Token, scan.ScanTime(project.Name), sys); err != nil {
				return err
			}
		}
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

	reportPath := filepath.Join(config.ReportDirectoryName, "project-names-aggregated.txt")
	if err := utils.FileWrite(reportPath, []byte(projectNames), 0644); err != nil {
		return err
	}
	if err := newVulnerabilityExcelReport(versionWideAlerts, config, utils); err != nil {
		return err
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

	if err := utils.MkdirAll(config.ReportDirectoryName, 0777); err != nil {
		return err
	}

	fileName := filepath.Join(config.ReportDirectoryName,
		fmt.Sprintf("vulnerabilities-%s.xlsx", utils.Now().Format(wsReportTimeStampLayout)))
	stream, err := utils.FileOpen(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	if err := file.Write(stream); err != nil {
		return err
	}
	return nil
}

func fillVulnerabilityExcelReport(alerts []ws.Alert, streamWriter *excelize.StreamWriter, styleID int) error {
	rows := []struct {
		axis  string
		title string
	}{
		{"A1", "Severity"},
		{"B1", "Library"},
		{"C1", "Vulnerability ID"},
		{"D1", "Project"},
		{"E1", "Resolution"},
	}
	for _, row := range rows {
		err := streamWriter.SetRow(row.axis, []interface{}{excelize.Cell{StyleID: styleID, Value: row.title}})
		if err != nil {
			return err
		}
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
			log.Entry().Errorf("failed to write alert row: %v", err)
		}
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
	fileName := fmt.Sprintf("%s/libraries-%s.csv", config.ReportDirectoryName,
		utils.Now().Format(wsReportTimeStampLayout))
	if err := utils.FileWrite(fileName, []byte(output), 0777); err != nil {
		return err
	}
	return nil
}

// persistScannedProjects writes all actually scanned WhiteSource project names as comma separated
// string into the Common Pipeline Environment, from where it can be used by sub-sequent steps.
func persistScannedProjects(config *ScanOptions, scan *ws.Scan, utils whitesourceUtils) error {
	var projectNames []string
	if config.ProjectName != "" {
		projectNames = []string{config.ProjectName + " - " + config.ProductVersion}
	} else {
		for _, project := range scan.ScannedProjects() {
			projectNames = append(projectNames, project.Name)
		}
		// Sorting helps the list become stable across pipeline runs (and in the unit tests),
		// as the order in which we travers map keys is not deterministic.
		sort.Strings(projectNames)
	}
	resourceDir := filepath.Join(".pipeline", "commonPipelineEnvironment", "custom")
	if err := utils.MkdirAll(resourceDir, 0755); err != nil {
		return err
	}
	fileContents := strings.Join(projectNames, ",")
	resource := filepath.Join(resourceDir, "whitesourceProjectNames")
	if err := utils.FileWrite(resource, []byte(fileContents), 0644); err != nil {
		return err
	}
	return nil
}
