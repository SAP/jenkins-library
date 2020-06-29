package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/SAP/jenkins-library/pkg/whitesource"
)

// just to make the lines less long
type ScanOptions = whitesourceExecuteScanOptions
type System = whitesource.System

func whitesourceExecuteScan(config ScanOptions, telemetry *telemetry.CustomData) {
	// reroute cmd output to logging framework
	c := command.Command{}
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	sys := whitesource.NewSystem(config.ServiceURL, config.OrgToken, config.UserToken)
	if err := resolveProjectIdentifiers(&c, sys, &config); err != nil {
		log.Entry().WithError(err).Fatal("step execution failed on resolving project identifiers")
	}

	// Generate a vulnerability report for all projects with version = config.ProjectVersion
	if config.AggregateVersionWideReport {
		if err := aggregateVersionWideLibraries(sys, &config); err != nil {
			log.Entry().WithError(err).Fatal("step execution failed on aggregating version wide libraries")
		}
		if err := aggregateVersionWideVulnerabilities(sys, &config); err != nil {
			log.Entry().WithError(err).Fatal("step execution failed on aggregating version wide vulnerabilities")
		}
	} else {
		if err := runWhitesourceScan(&config, sys, telemetry, &c); err != nil {
			log.Entry().WithError(err).Fatal("step execution failed on executing whitesource scan")
		}
	}
}

func runWhitesourceScan(config *ScanOptions, sys *System, _ *telemetry.CustomData, cmd *command.Command) error {
	// Start the scan
	if err := triggerWhitesourceScan(cmd, config); err != nil {
		return err
	}

	// Scan finished: we need to resolve project token again if the project was just created.
	if err := resolveProjectIdentifiers(cmd, sys, config); err != nil {
		return err
	}

	log.Entry().Info("-----------------------------------------------------")
	log.Entry().Infof("Project name: '%s'", config.ProjectName)
	log.Entry().Infof("Product Version: '%s'", config.ProductVersion)
	log.Entry().Infof("Project Token: %s", config.ProjectToken)
	log.Entry().Info("-----------------------------------------------------")

	if config.Reporting {
		paths, err := downloadReports(config, sys)
		if err != nil {
			return err
		}
		piperutils.PersistReportsAndLinks("whitesourceExecuteScan", "", nil, paths)
	}

	// Check for security vulnerabilities and fail the build if cvssSeverityLimit threshold is crossed
	if config.SecurityVulnerabilities {
		if err := checkSecurityViolations(config, sys); err != nil {
			return err
		}
	}
	return nil
}

func resolveProjectIdentifiers(cmd *command.Command, sys *System, config *ScanOptions) error {
	if config.ProjectName == "" || config.ProductVersion == "" {
		opts := &versioning.Options{}
		artifact, err := versioning.GetArtifact(config.ScanType, config.BuildDescriptorFile, opts, cmd)
		if err != nil {
			return err
		}
		gav, err := artifact.GetCoordinates()
		if err != nil {
			return err
		}

		nameTmpl := `{{list .GroupID .ArtifactID | join "-" | trimAll "-"}}`
		pName, pVer := versioning.DetermineProjectCoordinates(nameTmpl, config.DefaultVersioningModel, gav)
		if config.ProjectName == "" {
			log.Entry().Infof("Resolved project name '%s' from descriptor file", pName)
			config.ProjectName = pName
		}
		if config.ProductVersion == "" {
			log.Entry().Infof("Resolved project version '%s' from descriptor file", pVer)
			config.ProductVersion = pVer
		}
	}

	// Get product token if user did not specify one at runtime
	if config.ProductToken == "" {
		log.Entry().Infof("Attempting to resolve product token for product '%s'..", config.ProductName)
		product, err := sys.GetProductByName(config.ProductName)
		if err != nil {
			return err
		}
		if product != nil {
			log.Entry().Infof("Resolved product token: '%s'..", product.Token)
			config.ProductToken = product.Token
		}
	}

	// Get project token  if user did not specify one at runtime
	if config.ProjectToken == "" {
		log.Entry().Infof("Attempting to resolve project token for project '%s'..", config.ProjectName)
		fullProjName := fmt.Sprintf("%s - %s", config.ProjectName, config.ProductVersion)
		projectToken, err := sys.GetProjectToken(config.ProductToken, fullProjName)
		if err != nil {
			return err
		}
		if projectToken != "" {
			log.Entry().Infof("Resolved project token: '%s'..", projectToken)
			config.ProjectToken = projectToken
		}
	}
	return nil
}

func triggerWhitesourceScan(cmd *command.Command, config *ScanOptions) error {
	switch config.ScanType {
	case "npm":
		// Execute whitesource scan with
		if err := executeNpmScan(config, cmd); err != nil {
			return err
		}
	default:
		// Download the unified agent jar file if one does not exist
		if err := downloadAgent(config, cmd); err != nil {
			return err
		}

		// Auto generate a config file based on the working directory's contents.
		// TODO/NOTE: Currently this scans the UA jar file as a dependency since it is downloaded beforehand
		if err := autoGenerateWhitesourceConfig(config, cmd); err != nil {
			return err
		}

		// Execute whitesource scan with unified agent jar file
		if err := executeUAScan(config, cmd); err != nil {
			return err
		}
	}
	return nil
}

// executeUAScan
// Executes a scan with the Whitesource Unified Agent
// returns stdout buffer of the unified agent for token extraction in case of multi-module gradle project
func executeUAScan(config *ScanOptions, cmd *command.Command) error {
	return cmd.RunExecutable("java", "-jar", config.AgentFileName, "-d", ".", "-c", config.ConfigFilePath,
		"-apiKey", config.OrgToken, "-userKey", config.UserToken, "-project", config.ProjectName,
		"-product", config.ProductName, "-productVersion", config.ProductVersion)
}

// executeNpmScan
// generates a configuration file whitesource.config.json with appropriate values from config,
// installs whitesource yarn plugin and executes the scan
func executeNpmScan(config *ScanOptions, cmd *command.Command) error {
	npmConfig := []byte(fmt.Sprintf(`{
		"apiKey": "%s",
		"userKey": "%s",
		"checkPolicies": true,
		"productName": "%s",
		"projectName": "%s",
		"productVer": "%s",
		"devDep": true
	}`, config.OrgToken, config.UserToken, config.ProductName, config.ProjectName, config.ProductVersion))
	if err := ioutil.WriteFile("whitesource.config.json", npmConfig, 0644); err != nil {
		return err
	}
	if err := cmd.RunExecutable("yarn", "global", "add", "whitesource"); err != nil {
		return err
	}
	if err := cmd.RunExecutable("yarn", "install"); err != nil {
		return err
	}
	if err := cmd.RunExecutable("whitesource", "yarn"); err != nil {
		return err
	}
	return nil
}

// checkSecurityViolations: checks security violations and fails build is severity limit is crossed
func checkSecurityViolations(config *ScanOptions, sys *System) error {
	severeVulnerabilities := 0

	// convert config.CvssSeverityLimit to float64
	cvssSeverityLimit, err := strconv.ParseFloat(config.CvssSeverityLimit, 64)
	if err != nil {
		return err
	}

	// get project alerts (vulnerabilities)
	alerts, err := sys.GetProjectAlerts(config.ProjectToken)
	if err != nil {
		return err
	}

	// https://github.com/SAP/jenkins-library/blob/master/vars/whitesourceExecuteScan.groovy#L537
	for _, alert := range alerts {
		vuln := alert.Vulnerability
		if (vuln.Score >= cvssSeverityLimit || vuln.CVSS3Score >= cvssSeverityLimit) && cvssSeverityLimit >= 0 {
			severeVulnerabilities++
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

// pollProjectStatus polls project LastUpdateTime until it reflects the most recent scan
func pollProjectStatus(config *ScanOptions, sys *System) error {
	currentTime := time.Now()
	for {
		project, err := sys.GetProjectVitals(config.ProjectToken)
		if err != nil {
			return err
		}

		// Make sure the project was updated in whitesource backend before downloading any reports
		lastUpdatedTime, err := time.Parse("2006-01-02 15:04:05 +0000", project.LastUpdateDate)
		if currentTime.Sub(lastUpdatedTime) < 10*time.Second {
			//done polling
			break
		}
		log.Entry().Info("time since project was last updated > 10 seconds, polling status...")
		time.Sleep(5 * time.Second)
	}
	return nil
}

// downloadReports downloads a project's risk and vulnerability reports
func downloadReports(config *ScanOptions, sys *System) ([]piperutils.Path, error) {
	utils := piperutils.Files{}

	// Project was scanned, now we need to wait for Whitesource backend to propagate the changes
	if err := pollProjectStatus(config, sys); err != nil {
		return nil, err
	}

	if err := utils.MkdirAll(config.ReportDirectoryName, 0777); err != nil {
		return nil, err
	}
	vulnPath, err := downloadVulnerabilityReport(config, sys)
	if err != nil {
		return nil, err
	}
	riskPath, err := downloadRiskReport(config, sys)
	if err != nil {
		return nil, err
	}
	return []piperutils.Path{*vulnPath, *riskPath}, nil
}

func downloadVulnerabilityReport(config *ScanOptions, sys *System) (*piperutils.Path, error) {
	utils := piperutils.Files{}
	if err := utils.MkdirAll(config.ReportDirectoryName, 0777); err != nil {
		return nil, err
	}

	reportBytes, err := sys.GetProjectVulnerabilityReport(config.ProjectToken, config.VulnerabilityReportFormat)
	if err != nil {
		return nil, err
	}

	// Write report to file
	rptFileName := fmt.Sprintf("%s-vulnerability-report.%s", config.ProjectName, config.VulnerabilityReportFormat)
	rptFileName = filepath.Join(config.ReportDirectoryName, rptFileName)
	if err := ioutil.WriteFile(rptFileName, reportBytes, 0644); err != nil {
		return nil, err
	}

	log.Entry().Infof("Successfully downloaded vulnerability report to %s", rptFileName)
	pathName := fmt.Sprintf("%s Vulnerability Report", config.ProjectName)
	return &piperutils.Path{Name: pathName, Target: rptFileName}, nil
}

func downloadRiskReport(config *ScanOptions, sys *System) (*piperutils.Path, error) {
	reportBytes, err := sys.GetProjectRiskReport(config.ProjectToken)
	if err != nil {
		return nil, err
	}

	rptFileName := fmt.Sprintf("%s-risk-report.pdf", config.ProjectName)
	rptFileName = filepath.Join(config.ReportDirectoryName, rptFileName)
	if err := ioutil.WriteFile(rptFileName, reportBytes, 0644); err != nil {
		return nil, err
	}

	log.Entry().Infof("Successfully downloaded risk report to %s", rptFileName)
	pathName := fmt.Sprintf("%s PDF Risk Report", config.ProjectName)
	return &piperutils.Path{Name: pathName, Target: rptFileName}, nil
}

// downloadAgent: Downloads the unified agent jar file if one does not exist
func downloadAgent(config *ScanOptions, cmd *command.Command) error {
	agentFile := config.AgentFileName
	if !fileExists(agentFile) {
		if err := cmd.RunExecutable("curl", "-L", config.AgentDownloadURL, "-o", agentFile); err != nil {
			return err
		}
	}
	return nil
}

// autoGenerateWhitesourceConfig
// Auto generate a config file based on the current directory structure, renames it to user specified configFilePath
// Generated file name will be 'wss-generated-file.config'
func autoGenerateWhitesourceConfig(config *ScanOptions, cmd *command.Command) error {
	// TODO: Should we rely on -detect, or set the parameters manually?
	if err := cmd.RunExecutable("java", "-jar", config.AgentFileName, "-d", ".", "-detect"); err != nil {
		return err
	}

	// Rename generated config file to config.ConfigFilePath parameter
	if err := os.Rename("wss-generated-file.config", config.ConfigFilePath); err != nil {
		return err
	}

	// Append aggregateModules=true parameter to config file (consolidates multi-module projects into one)
	f, err := os.OpenFile(config.ConfigFilePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Append additional config parameters to prevent multiple projects being generated
	cfg := fmt.Sprintf("gradle.aggregateModules=true\nmaven.aggregateModules=true\ngradle.localRepositoryPath=.gradle\nmaven.m2RepositoryPath=.m2\nexcludes=%s", config.Excludes)
	if _, err = f.WriteString(cfg); err != nil {
		return err
	}

	// archiveExtractionDepth=0
	if err := cmd.RunExecutable("sed", "-ir", `s/^[#]*\s*archiveExtractionDepth=.*/archiveExtractionDepth=0/`,
		config.ConfigFilePath); err != nil {
		return err
	}

	// config.Includes defaults to "**/*.java **/*.jar **/*.py **/*.go **/*.js **/*.ts"
	regex := fmt.Sprintf(`s/^[#]*\s*includes=.*/includes="%s"/`, config.Includes)
	if err := cmd.RunExecutable("sed", "-ir", regex, config.ConfigFilePath); err != nil {
		return err
	}

	return nil
}

func aggregateVersionWideLibraries(sys *System, config *ScanOptions) error {
	log.Entry().Infof("Aggregating list of libraries used for all projects with version: %s", config.ProductVersion)

	projects, err := sys.GetProjectsMetaInfo(config.ProductToken)
	if err != nil {
		return err
	}

	versionWideLibraries := map[string][]whitesource.Library{} // maps project name to slice of libraries
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
	if err := newLibraryCSVReport(versionWideLibraries, config); err != nil {
		return err
	}
	return nil
}

func aggregateVersionWideVulnerabilities(sys *System, config *ScanOptions) error {
	log.Entry().Infof("Aggregating list of vulnerabilities for all projects with version: %s", config.ProductVersion)

	projects, err := sys.GetProjectsMetaInfo(config.ProductToken)
	if err != nil {
		return err
	}

	var versionWideAlerts []whitesource.Alert // all alerts for a given project version
	projectNames := ``                        // holds all project tokens considered a part of the report for debugging
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
	if err := newVulnerabilityExcelReport(versionWideAlerts, config); err != nil {
		return err
	}
	return nil
}

// outputs an slice of alerts to an excel file
func newVulnerabilityExcelReport(alerts []whitesource.Alert, config *ScanOptions) error {
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

	utils := piperutils.Files{}
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
func newLibraryCSVReport(libraries map[string][]whitesource.Library, config *ScanOptions) error {
	output := "Library Name, Project Name\n"
	for projectName, libraries := range libraries {
		log.Entry().Infof("Writing %v libraries for project %s to excel report..", len(libraries), projectName)
		for _, library := range libraries {
			output += library.Name + ", " + projectName + "\n"
		}
	}

	// Ensure reporting directory exists
	utils := piperutils.Files{}
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
