package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/whitesource"
	"golang.org/x/mod/modfile"
)

func whitesourceExecuteScan(config whitesourceExecuteScanOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	sys := whitesource.NewSystem(config.ServiceURL, config.OrgToken, config.UserToken)
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runWhitesourceExecuteScan(config, sys, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runWhitesourceExecuteScan(config whitesourceExecuteScanOptions, sys whitesource.System, telemetryData *telemetry.CustomData, command *command.Command) error {
	resolveProjectIdentifiers(&config, command, sys)
	projectsScanned := triggerWhitesourceScan(command, config, sys)

	if config.ProjectToken == "" {
		// This state occurs when the project was created during the triggerWhitesourceScan step above,
		resolveProjectIdentifiers(&config, command, sys) // we need to resolve the ProjectToken for pdf report download stage below
	}

	log.Entry().Info("Config.ProductVersion: " + config.ProductVersion)
	log.Entry().Info("Config.ProjectToken: " + config.ProjectToken)
	log.Entry().Infof("Number of projects scanned: %v", len(projectsScanned))

	if config.Reporting {
		for _, proj := range projectsScanned {
			proj.Name = strings.Split(proj.Name, " - ")[0]
			log.Entry().Infof("Attempting to download risk report for project name: %s", proj.Name)
			downloadRiskReport(proj.Token, proj.Name, sys)
		}
	}
	return nil
}

// translated from Groovy DSL: https://github.com/SAP/jenkins-library/blob/4c97231ff94d9f9fbeb1b58c3625c84b434a7de6/vars/whitesourceExecuteScan.groovy#L304
func triggerWhitesourceScan(command *command.Command, config whitesourceExecuteScanOptions, sys whitesource.System) []whitesource.Project {
	projectsScanned := []whitesource.Project{}
	newProj := whitesource.Project{Name: config.ProjectName, Token: config.ProjectToken}
	projectsScanned = append(projectsScanned, newProj)

	switch config.ScanType {
	case "npm":
		executeNpmScan(config, command)
		break

	default:
		// Auto generate a config file based on the current directory structure.
		err := command.RunExecutable("java", "-jar", config.AgentFileName, "-d", ".", "-detect")
		if err != nil {
			log.Entry().WithError(err).Fatal("Failed to autogenerate Whitesource unified agent config file")
		}

		err = command.RunExecutable("java", "-jar", config.AgentFileName, "-d", ".", "-c", "wss-generated-file.config",
			"-apiKey", config.OrgToken, "-userKey", config.UserToken, "-project", config.ProjectName,
			"-product", config.ProductName, "-productVersion", config.ProductVersion)
		if err != nil {
			log.Entry().WithError(err).Fatal("Failed to run Whitesource unified agent, exit..")
		}

		// USE CASE: Handle multi-module gradle projects
		if config.ScanType == "gradle" {
			projectsScanned = extractProjectTokensFromStdout(os.Stdout, config, sys)
		}
		break
	}
	return projectsScanned
}

// deal with multimodule gradle projects... there's probably a better way of doing this...
// Problem: Find all project tokens scanned that are apart of multimodule scan.
// Issue: Only have access to a single project token to begin with (config.ProjectToken)
func extractProjectTokensFromStdout(stdout *os.File, config whitesourceExecuteScanOptions, sys whitesource.System) []whitesource.Project {

	log.Entry().Info("Running grep command on whitesource stdout...")
	cmd := exec.Command("grep", "URL: ")
	projectsInfoBuffer := &bytes.Buffer{}
	cmd.Stdout = projectsInfoBuffer
	cmd.Stdin = stdout
	err := cmd.Run()
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to run 'grep URL: ', exit..")
	}

	log.Entry().Info("Extracting project tokens from whitesource scan stdout...")
	ids := []int64{}
	projectMetaStr := projectsInfoBuffer.String()
	projectMetas := strings.Split(projectMetaStr, "id=")
	for _, idStr := range projectMetas {
		idStr = strings.Split(idStr, "\n")[0]
		if !strings.HasPrefix(idStr, "[INFO]") {
			idInt := int64(Atoi(idStr))
			ids = append(ids, idInt)
		}
	}

	projects, err := sys.GetProjectsByIds(config.ProductToken, ids)
	if err != nil {
		log.Entry().WithError(err).Errorf("Could not get project by IDs: %s", ids)
	}

	return projects
}

// executeNpmScan:
// generates a configuration file whitesource.config.json with appropriate values from config,
// installs whitesource yarn plugin and executes the scan
func executeNpmScan(config whitesourceExecuteScanOptions, command *command.Command) {
	npmConfig := []byte(fmt.Sprintf(`{
		"apiKey": "%s",
		"userKey": "%s",
		"checkPolicies": true,
		"productName": "%s",
		"projectName": "%s",
		"productVer": "%s",
		"devDep": true
	}`, config.OrgToken, config.UserToken, config.ProductName, config.ProjectName, config.ProductVersion))

	writeToFile("./whitesource.config.json", npmConfig, 0644)

	err := command.RunExecutable("yarn", "global", "add", "whitesource")
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to run 'yarn global add whitesource'")
	}
	err = command.RunExecutable("yarn", "install")
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to run 'yarn install'")
	}
	err = command.RunExecutable("whitesource", "yarn")
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to run 'whitesource yarn'")
	}
}

// translated from Groovy DSL: https://github.com/SAP/jenkins-library/blob/4c97231ff94d9f9fbeb1b58c3625c84b434a7de6/vars/whitesourceExecuteScan.groovy#L443
func resolveProjectIdentifiers(config *whitesourceExecuteScanOptions, command *command.Command, sys whitesource.System) {
	if config.ProjectName == "" || config.ProductVersion == "" {
		var gav map[string]string

		switch config.ScanType {
		case "npm":
			gav = getNpmGAV()
			break
		case "go":
			gav = getGoGAV()
			break
		case "gradle":
			gav = getGradleGAV(command)
			break

		default:
			log.Entry().Warnf("resolveProjectIdentifiers: ScanType %s not implemented", config.ScanType)
			break
		}

		if config.ProjectName == "" {
			if gav["group"] != "" {
				config.ProjectName = fmt.Sprintf("%s.%s", gav["group"], gav["artifact"])
			} else {
				config.ProjectName = gav["artifact"]
			}
		}

		if gav["version"] != "" && config.ProductVersion == "" {
			config.ProductVersion = gav["version"]
		} else if gav["version"] == "" && config.ProductVersion == "" {
			// set default version if one could not be resolved
			config.ProductVersion = "unspecified"
		}
	}

	if config.ProductToken == "" && config.ProductName != "" {
		product, err := sys.GetProductByName(config.ProductName)
		if err != nil {
			log.Entry().Info(fmt.Sprintf("Product %s does not yet exist", config.ProductName))
		} else {
			config.ProductToken = product.Token
		}
	}

	if config.ProjectToken == "" && config.ProjectName != "" {
		fullProjectName := config.ProjectName + " - " + config.ProductVersion
		project, err := sys.GetProjectByName(config.ProductToken, fullProjectName)
		if err != nil {
			log.Entry().Info(fmt.Sprintf("Project %s does not yet exist...", fullProjectName))
		} else {
			config.ProjectToken = project.Token
		}
	}
}

// Project identifier resolvers
func getNpmGAV() map[string]string {
	result := map[string]string{}
	bytes := readFile("package.json")
	packageJSON := map[string]interface{}{}
	err := json.Unmarshal(bytes, &packageJSON)
	if err != nil {
		log.Entry().WithError(err).Fatal("Unable to unmarshal bytes into map")
	}

	projectName := packageJSON["name"].(string)
	if strings.HasPrefix(projectName, "@") {
		packageNameArray := strings.Split(projectName, "/")
		if len(packageNameArray) != 2 {
			log.Entry().Warn("Failed to parse package name:", projectName)
		}
		result["group"] = packageNameArray[0]
		result["artifact"] = packageNameArray[1]
	} else {
		result["group"] = ""
		result["artifact"] = projectName
	}
	result["version"] = packageJSON["version"].(string)
	log.Entry().Info("Resolved NPM project version: ", result["version"])
	return result
}

func getGoGAV() map[string]string {
	result := map[string]string{}
	if fileExists("Gopkg.toml") { // Godep
		log.Entry().Fatal("Gopkg.toml parsing not implemented. exit...")
	} else if fileExists("go.mod") { // Go modules
		bytes := readFile("go.mod")

		m, err := modfile.Parse("go.mod", bytes, nil)
		if err != nil {
			log.Entry().WithError(err).Fatal("Failed to read go.mod file %s: %v")
		}
		artifactSplit := strings.Split(m.Module.Mod.Path, "/")
		artifact := artifactSplit[len(artifactSplit)-1]
		result["artifact"] = artifact
		result["version"] = m.Module.Mod.Version
		log.Entry().Info("Resolved golang project version using go.mod parser: ", result["version"])
	} else {
		log.Entry().Fatal("Could not find suitable dependency file (go.mod, gopkg.toml, etc..) in working directory")
	}
	return result
}

func getGradleGAV(command *command.Command) map[string]string {
	result := map[string]string{}
	gradlePropsOut := &bytes.Buffer{}

	log.Entry().Info("Attempting to resolve gradle project name and version with gradle properties...")
	cmd := exec.Command("gradle", "properties")
	cmd.Stdout = gradlePropsOut
	err := cmd.Run()
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to run 'gradle properties', exit..")
	}
	gradlePropsOutput := *gradlePropsOut // needed for future command

	cmd = exec.Command("grep", "^rootProject")
	projectNameOut := &bytes.Buffer{}
	cmd.Stdout = projectNameOut
	cmd.Stdin = gradlePropsOut
	err = cmd.Run()
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to run 'grep ^rootProject', exit..")
	}

	cmd = exec.Command("grep", "^version:")
	projectVersionOut := &bytes.Buffer{}
	cmd.Stdout = projectVersionOut
	cmd.Stdin = &gradlePropsOutput
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error:", err.Error())
		log.Entry().WithError(err).Fatal("Failed to run 'grep ^version', exit..")
	}

	// Extract project name
	projectNameFragments := strings.Split(string(projectNameOut.Bytes()), "'")
	projectName := ""
	if len(projectNameFragments) > 1 {
		projectName = projectNameFragments[1]
	}
	result["artifact"] = projectName

	// Extract project version
	projectVersionFragments := strings.Split(string(projectVersionOut.Bytes()), "version: ")
	productVersion := ""
	result["version"] = ""
	if len(projectNameFragments) > 1 {
		productVersion = projectVersionFragments[1]
	}
	if !strings.Contains(productVersion, "unspecified") {
		result["version"] = productVersion
	}
	log.Entry().Infof("Resolved gradle project version: %s and name: %s ", result["version"], result["artifact"])

	return result
}

// Download PDF Risk report for a given projectToken and projectName
func downloadRiskReport(projectToken string, projectName string, sys whitesource.System) {
	reportBytes, err := sys.GetProjectRiskReport(projectToken)
	if err != nil {
		log.Entry().Warn(fmt.Sprintf("Failed to generate report for project name %s", projectName))
	}

	// create report directory if dne
	reportDir := "whitesource-report"
	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		err = os.Mkdir(reportDir, 0777)
		if err != nil {
			log.Entry().Warn("Failed to create whitesource-report directory: ", err)
		}
	}
	reportFileName := fmt.Sprintf("%s/%s-risk-report.pdf", reportDir, projectName)
	writeToFile(reportFileName, reportBytes, 0777)

	log.Entry().Infof("Successfully downloaded risk report to ./%s", reportFileName)
}

/******* Utility functions *******/
func readFile(filename string) []byte {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to read file: %s", filename)
		return nil
	}
	return bytes
}

func writeToFile(filename string, bytes []byte, mode os.FileMode) {
	err := ioutil.WriteFile(filename, bytes, mode)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to write to whitesource config")
	}
}

func Atoi(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		log.Entry().WithError(err).Warn("Could not convert string ID to integer")
	}
	return num
}

// func fileExists(filename string) bool {
// 	info, err := os.Stat(filename)
// 	if os.IsNotExist(err) {
// 		return false
// 	}
// 	return !info.IsDir()
// }
