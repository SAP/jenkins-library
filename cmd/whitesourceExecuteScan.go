package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/SAP/jenkins-library/pkg/whitesource"
)

func whitesourceExecuteScan(config whitesourceExecuteScanOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	sys := whitesource.NewSystem(config.ServiceURL, config.OrgToken, config.UserToken)
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runWhitesourceScan(&config, sys, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runWhitesourceScan(config *whitesourceExecuteScanOptions, sys whitesource.System, _ *telemetry.CustomData,
	command *command.Command) error {

	err := resolveProjectIdentifiers(command, config)
	if err != nil {
		return err
	}

	// Start the scan
	projectsScanned, err := triggerWhitesourceScan(command, config, sys)
	if err != nil {
		return err
	}
	// Scan finished

	log.Entry().Info("-----------------------------------------------------")
	log.Entry().Infof("Project name: '%s'", config.ProjectName)
	log.Entry().Infof("Product Version: '%s'", config.ProductVersion)
	log.Entry().Infof("Project Token: %s", config.ProjectToken)
	log.Entry().Infof("BuildDescriptorFile: %s", config.BuildDescriptorFile)
	log.Entry().Infof("Number of projects scanned: %v", len(projectsScanned))
	log.Entry().Info("-----------------------------------------------------")

	if config.Reporting {
		var finalPaths []piperutils.Path
		for _, proj := range projectsScanned {
			proj.Name = strings.Split(proj.Name, " - ")[0]
			paths, err := downloadReports(proj.Name, proj.Token, config, sys)
			if err != nil {
				return err
			}
			finalPaths = append(finalPaths, paths...)
		}
		piperutils.PersistReportsAndLinks("whitesourceExecuteScan", "./", nil, finalPaths)
	}

	// Check for security vulnerabilities and fail the build if cvssSeverityLimit threshold is crossed
	if config.SecurityVulnerabilities {
		for _, proj := range projectsScanned {
			err = checkSecurityViolations(proj.Token, config, sys)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func resolveProjectIdentifiers(command *command.Command, config *whitesourceExecuteScanOptions) error {
	opts := &versioning.Options{}
	artifact, err := versioning.GetArtifact(config.ScanType, config.BuildDescriptorFile, opts, command)
	if err != nil {
		return err
	}
	gav, err := artifact.GetCoordinates()
	if err != nil {
		return err
	}
	log.Entry().Info("GAV:", gav)
	if config.ProjectName == "" || config.ProductVersion == "" {
		nameTmpl := `{{list .GroupID .ArtifactID | join "-" | trimAll "-"}}`
		projectName, projectVersion := versioning.DetermineProjectCoordinates(nameTmpl,
			config.DefaultVersioningModel, gav)

		log.Entry().Info("Determined project version:", projectVersion)
		if config.ProjectName == "" {
			config.ProjectName = projectName
		}
		if config.ProductVersion == "" {
			config.ProductVersion = projectVersion
		}
	}
	return nil
}

func triggerWhitesourceScan(command *command.Command, config *whitesourceExecuteScanOptions,
	sys whitesource.System) ([]whitesource.Project, error) {

	var projectsScanned []whitesource.Project
	switch config.ScanType {
	case "npm":
		err := executeNpmScan(*config, command)
		if err != nil {
			return nil, err
		}
		break

	default:
		// Download the unified agent jar file if one does not exist
		if !fileExists(config.AgentFileName) {
			err := command.RunExecutable("curl", "-L", config.AgentDownloadURL,
				"-o", config.AgentFileName)
			if err != nil {
				return nil, err
			}
		}

		// Auto generate a config file based on the current directory structure.
		// Generated file name will be 'wss-generated-file.config'
		err := command.RunExecutable("java", "-jar", config.AgentFileName,
			"-d", ".", "-detect")
		if err != nil {
			return nil, err
		}

		// Rename generated config file to config.ConfigFilePath parameter
		err = os.Rename("wss-generated-file.config", config.ConfigFilePath)
		if err != nil {
			return nil, err
		}

		wsOutputBuffer := &bytes.Buffer{}
		command.Stdout(io.MultiWriter(log.Writer(), wsOutputBuffer))
		err = command.RunExecutable("java", "-jar", config.AgentFileName, "-d", ".",
			"-c", config.ConfigFilePath, "-apiKey", config.OrgToken, "-userKey", config.UserToken,
			"-project", config.ProjectName, "-product", config.ProductName, "-productVersion",
			config.ProductVersion)
		if err != nil {
			return nil, err
		}

		if config.ProductToken == "" {
			product, err := sys.GetProductByName(config.ProductName)
			if err != nil {
				return nil, err
			}
			config.ProductToken = product.Token
		}

		if config.ProjectToken == "" {
			fullProjName := fmt.Sprintf("%s-%s", config.ProjectName, config.ProductVersion)
			projectToken, err := sys.GetProjectToken(config.ProductToken, fullProjName)
			if err != nil {
				return nil, err
			}
			config.ProjectToken = projectToken
		}

		// USE CASE: Handle multi-module gradle projects
		if config.ScanType == "gradle" {
			projectsScanned, err = extractProjectTokensFromStdout(wsOutputBuffer, *config, sys)
			if err != nil {
				return nil, err
			}
		} else {
			newProj := whitesource.Project{Name: config.ProjectName, Token: config.ProjectToken}
			projectsScanned = append(projectsScanned, newProj)
		}
		break
	}
	return projectsScanned, nil
}

// deal with multimodule gradle projects... there's probably a better way of doing this...
// Problem: Find all project tokens scanned that are apart of multimodule scan.
// Issue: Only have access to a single project token to begin with (config.ProjectToken)
// TODO: Find a better way of doing this instead of extracting from unified agent's stdout...
func extractProjectTokensFromStdout(wsOutput *bytes.Buffer, config whitesourceExecuteScanOptions,
	sys whitesource.System) ([]whitesource.Project, error) {
	log.Entry().Info("Extracting project tokens from whitesource stdout..")

	ids := []int64{}
	r := regexp.MustCompile(`#!project;id=(.*[0-9])`)
	projectMetaStr := wsOutput.String()
	matches := r.FindAllString(projectMetaStr, -1)
	for _, match := range matches {
		versionStr := strings.Split(match, "id=")[1]
		versionInt, err := strconv.Atoi(versionStr)
		if err != nil {
			return nil, err
		}
		ids = append(ids, int64(versionInt))
	}

	log.Entry().Info("Getting projects by ids..")
	projects, err := sys.GetProjectsByIDs(config.ProductToken, ids)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// executeNpmScan:
// generates a configuration file whitesource.config.json with appropriate values from config,
// installs whitesource yarn plugin and executes the scan
func executeNpmScan(config whitesourceExecuteScanOptions, command *command.Command) error {
	npmConfig := []byte(fmt.Sprintf(`{
		"apiKey": "%s",
		"userKey": "%s",
		"checkPolicies": true,
		"productName": "%s",
		"projectName": "%s",
		"productVer": "%s",
		"devDep": true
	}`, config.OrgToken, config.UserToken, config.ProductName, config.ProjectName, config.ProductVersion))

	err := ioutil.WriteFile("whitesource.config.json", npmConfig, 0644)
	if err != nil {
		return err
	}

	err = command.RunExecutable("yarn", "global", "add", "whitesource")
	if err != nil {
		return err
	}
	err = command.RunExecutable("yarn", "install")
	if err != nil {
		return err
	}
	err = command.RunExecutable("whitesource", "yarn")
	if err != nil {
		return err
	}
	return nil
}

// checkSecurityViolations
func checkSecurityViolations(projectToken string, config *whitesourceExecuteScanOptions, sys whitesource.System) error {
	severeVulnerabilities := 0

	// convert config parameter to float
	cvssSeverityLimit, err := strconv.ParseFloat(config.CvssSeverityLimit, 64)
	if err != nil {
		return err
	}

	// get project alerts
	alerts, err := sys.GetProjectAlerts(projectToken)
	if err != nil {
		return err
	}

	// https://github.com/SAP/jenkins-library/blob/master/vars/whitesourceExecuteScan.groovy#L537
	for _, alert := range alerts {
		vuln := alert.Vulnerability
		if vuln.Score >= cvssSeverityLimit || vuln.CVSS3Score >= cvssSeverityLimit && cvssSeverityLimit >= 0 {
			severeVulnerabilities++
		}
	}

	//https://github.com/SAP/jenkins-library/blob/master/vars/whitesourceExecuteScan.groovy#L547
	nonSevereVulnerabilities := len(alerts) - severeVulnerabilities
	if nonSevereVulnerabilities > 0 {
		log.Entry().Infof("WARNING: %v Open Source Software Security vulnerabilities with "+
			"CVSS score below %s detected.", nonSevereVulnerabilities, config.CvssSeverityLimit)
	} else if len(alerts) == 0 {
		log.Entry().Info("No Open Source Software Security vulnerabilities detected.")
	}

	// https://github.com/SAP/jenkins-library/blob/master/vars/whitesourceExecuteScan.groovy#L558
	if severeVulnerabilities > 0 {
		return errors.New(fmt.Sprintf("%v Open Source Software Security vulnerabilities with CVSS "+
			"score greater or equal to %s detected.", severeVulnerabilities, config.CvssSeverityLimit))
	}
	return nil
}

// downloadReports downloads a project's risk and vulnerability reports
func downloadReports(projectName, projectToken string, config *whitesourceExecuteScanOptions,
	sys whitesource.System) ([]piperutils.Path, error) {

	vulnPath, err := downloadVulnerabilityReport(projectName, projectToken, config, sys)
	if err != nil {
		return nil, err
	}
	riskPath, err := downloadRiskReport(projectName, projectToken, config, sys)
	if err != nil {
		return nil, err
	}
	return []piperutils.Path{*vulnPath, *riskPath}, nil
}

func downloadVulnerabilityReport(projectName, projectToken string, config *whitesourceExecuteScanOptions,
	sys whitesource.System) (*piperutils.Path, error) {

	// create report directory if it DNE
	reportDir := "whitesource-report"
	utils := piperutils.Files{}
	err := utils.MkdirAll(reportDir, 0777)
	if err != nil {
		return nil, err
	}

	// Download report from Whitesource API
	reportBytes, err := sys.GetProjectVulnerabilityReport(projectToken, config.VulnerabilityReportFormat)
	if err != nil {
		return nil, err
	}

	// Write report to file
	reportFileName := strings.Replace(config.VulnerabilityReportFileName, `${config.projectName}`,
		projectName, -1)
	reportFileName = filepath.Join(reportDir, fmt.Sprintf("%s.%s", reportFileName,
		config.VulnerabilityReportFormat))
	err = ioutil.WriteFile(reportFileName, reportBytes, 0644)
	if err != nil {
		return nil, err
	}

	log.Entry().Infof("Successfully downloaded vulnerability report to %s", reportFileName)
	pathName := fmt.Sprintf("%s Vulnerability Report", projectName)
	return &piperutils.Path{Name: pathName, Target: reportFileName}, nil
}

func downloadRiskReport(projectName, projectToken string, config *whitesourceExecuteScanOptions,
	sys whitesource.System) (*piperutils.Path, error) {

	// create report directory if dne
	reportDir := "whitesource-report"
	utils := piperutils.Files{}
	err := utils.MkdirAll(reportDir, 0777)
	if err != nil {
		return nil, err
	}

	reportBytes, err := sys.GetProjectRiskReport(projectToken)
	if err != nil {
		return nil, err
	}

	reportFileName := strings.Replace(config.RiskReportFileName, `${config.projectName}`, projectName, -1)
	reportFileName = filepath.Join(reportDir, fmt.Sprintf("%s.pdf", reportFileName))
	err = ioutil.WriteFile(reportFileName, reportBytes, 0644)
	if err != nil {
		return nil, err
	}

	log.Entry().Infof("Successfully downloaded risk report to %s", reportFileName)
	pathName := fmt.Sprintf("%s PDF Risk Report", projectName)
	return &piperutils.Path{Name: pathName, Target: reportFileName}, nil
}
