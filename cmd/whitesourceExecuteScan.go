package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
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

func runWhitesourceScan(config *whitesourceExecuteScanOptions, sys whitesource.System, telemetryData *telemetry.CustomData, command *command.Command) error {
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
	log.Entry().Infof("Number of projects scanned: %v", len(projectsScanned))
	log.Entry().Info("-----------------------------------------------------")

	if config.Reporting {
		var links []piperutils.Path
		for _, proj := range projectsScanned {
			proj.Name = strings.Split(proj.Name, " - ")[0]
			link, err := downloadRiskReport(proj.Token, proj.Name, sys)
			if err != nil {
				return err
			}
			links = append(links, *link)
		}

		// publish pdf file locations
		piperutils.PersistReportsAndLinks("whitesourceExecuteScan", "./", nil, links)
	}
	return nil
}

func resolveProjectIdentifiers(command *command.Command, config *whitesourceExecuteScanOptions) error {
	artifact, err := versioning.GetArtifact(config.ScanType, config.BuildDescriptorFile, &versioning.Options{}, command)
	if err != nil {
		return err
	}
	gav, err := artifact.GetCoordinates()
	if err != nil {
		return err
	}
	if config.ProjectName == "" || config.ProductVersion == "" {
		projectName, projectVersion := versioning.DetermineProjectCoordinates(config.ProjectName, config.DefaultVersioningModel, gav)
		if config.ProjectName == "" {
			config.ProjectName = projectName
		}
		if config.ProductVersion == "" {
			config.ProductVersion = projectVersion
		}
	}
	return nil
}

func triggerWhitesourceScan(command *command.Command, config *whitesourceExecuteScanOptions, sys whitesource.System) ([]whitesource.Project, error) {
	var projectsScanned []whitesource.Project

	switch config.ScanType {
	case "npm":
		err := executeNpmScan(*config, command)
		if err != nil {
			return nil, err
		}
		break

	default:
		// Auto generate a config file based on the current directory structure.
		err := command.RunExecutable("java", "-jar", config.AgentFileName, "-d", ".", "-detect")
		if err != nil {
			return nil, err
		}

		wsOutputBuffer := &bytes.Buffer{}
		cmd := exec.Command("java", "-jar", config.AgentFileName, "-d", ".", "-c", "wss-generated-file.config",
			"-apiKey", config.OrgToken, "-userKey", config.UserToken, "-project", config.ProjectName,
			"-product", config.ProductName, "-productVersion", config.ProductVersion)
		cmd.Stdout = wsOutputBuffer
		err = cmd.Run()
		log.Entry().Info(wsOutputBuffer.String())
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
			projectToken, err := sys.GetProjectToken(config.ProductToken, config.ProjectName+" - "+config.ProductVersion)
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
func extractProjectTokensFromStdout(wsOutput *bytes.Buffer, config whitesourceExecuteScanOptions, sys whitesource.System) ([]whitesource.Project, error) {
	log.Entry().Info("Extracting project tokens from whitesource stdout..")

	// TODO: Use regexp
	cmd := exec.Command("grep", "URL: ")
	projectsInfoBuffer := &bytes.Buffer{}
	cmd.Stdout = projectsInfoBuffer
	cmd.Stdin = wsOutput
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	ids := []int64{}
	projectMetaStr := projectsInfoBuffer.String()
	projectMetas := strings.Split(projectMetaStr, "id=")
	for _, idStr := range projectMetas {
		idStr = strings.Split(idStr, `\n`)[0]
		if !strings.HasPrefix(idStr, "[INFO]") {
			idInt, err := strconv.Atoi(idStr)
			if err != nil {
				return nil, err
			}
			ids = append(ids, int64(idInt))
		}
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

// downloadRiskReport downloads a project's risk report and returns a piperutils.Path which link to the file
func downloadRiskReport(projectToken string, projectName string, sys whitesource.System) (*piperutils.Path, error) {
	reportBytes, err := sys.GetProjectRiskReport(projectToken)
	if err != nil {
		return nil, err
	}

	// create report directory if dne
	reportDir := "whitesource-report"
	utils := piperutils.Files{}
	err = utils.MkdirAll(reportDir, 0777)
	if err != nil {
		return nil, err
	}

	reportFileName := filepath.Join(reportDir, projectName+"-risk-report.pdf")
	err = ioutil.WriteFile(reportFileName, reportBytes, 0644)
	if err != nil {
		return nil, err
	}

	log.Entry().Infof("Successfully downloaded risk report to %s", reportFileName)
	return &piperutils.Path{Name: fmt.Sprintf("%s PDF Risk Report", projectName), Target: reportFileName}, nil
}
