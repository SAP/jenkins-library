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

	"github.com/SAP/jenkins-library/pkg/whitesource"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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
	fmt.Println("Config.ProjectToken: ", config.ProjectToken)
	fmt.Println("Config.ProductVersion: ", config.ProductVersion)
	if config.Reporting {
		if config.ScanType == "gradle" && len(projectsScanned) > 1 {
			// handle multi-module  gradle projects with multiple scan reports to download
			for _, proj := range projectsScanned {
				proj.Name = strings.Split(proj.Name, " - ")[0]
				downloadRiskReport(proj.Token, proj.Name, sys)
			}
		} else {
			downloadRiskReport(config.ProjectToken, config.ProjectName, sys)
		}
	}
	return nil
}

func downloadRiskReport(projectToken string, projectName string, sys whitesource.System) {
	log.Entry().Debug("Downloading risk report for project name:", projectName, " project token:", projectToken)
	log.Entry().Debug("Downloading risk report")
	reportBytes, err := sys.GetProjectRiskReport(projectToken)
	if err != nil {
		log.Entry().Warn(fmt.Sprintf("Failed to generate report for project name %s", projectName))
	}
	reportFileName := fmt.Sprintf("whitesource-report/%s-risk-report.pdf", projectName)

	// create report directory
	err = os.Mkdir("whitesource-report", 0777)
	if err != nil {
		log.Entry().Warn("Failed to create whitesource-report directory: ", err)
	}

	err = ioutil.WriteFile(reportFileName, reportBytes, 0777)
	if err != nil {
		log.Entry().Warn("Failed to write to ./whitesource-report/", projectName, "-risk-report.pdf:", err)
	}
}

// translated from Groovy DSL: https://github.com/SAP/jenkins-library/blob/4c97231ff94d9f9fbeb1b58c3625c84b434a7de6/vars/whitesourceExecuteScan.groovy#L304
func triggerWhitesourceScan(command *command.Command, config whitesourceExecuteScanOptions, sys whitesource.System) []whitesource.Project {
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

		outBuffer := &bytes.Buffer{}
		cmd := exec.Command("java", "-jar", config.AgentFileName, "-d", ".", "-c", "wss-generated-file.config",
			"-apiKey", config.OrgToken, "-userKey", config.UserToken, "-project", config.ProjectName,
			"-product", config.ProductName, "-productVersion", config.ProductVersion)
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			log.Entry().WithError(err).Fatal("Failed to run Whitesource unified agent, exit..")
		}
		log.Entry().Info(outBuffer.String())

		if config.ScanType == "gradle" {
			// deal with multimodule gradle projects... there's probably a better way of doing this...
			// Problem: Find all project tokens scanned that are apart of multimodule scan.
			// Issue: Only have access to a single project token
			cmd = exec.Command("grep", "URL: ")
			projectsInfoBuffer := &bytes.Buffer{}
			cmd.Stdout = projectsInfoBuffer
			cmd.Stdin = os.Stdout
			err = cmd.Run()
			if err != nil {
				log.Entry().WithError(err).Fatal("Failed to run 'grep URL: ', exit..")
			}

			ids := []int64{}
			projectMetaStr := projectsInfoBuffer.String()
			projectMetas := strings.Split(projectMetaStr, "id=")
			for _, id := range projectMetas {
				id = strings.Split(id, "\n")[0]
				if !strings.HasPrefix(id, "[INFO]") {
					idInt, err := strconv.Atoi(id)
					if err != nil {
						log.Entry().Warnf("Could not convert string ID to integer: %v", err)
					}
					ids = append(ids, int64(idInt))
				}
			}

			projectTokens, err := sys.GetProjectTokensByIds(config.ProductToken, ids)
			if err != nil {
				log.Entry().Errorf("Could not get project tokens by IDs:", err)
			}
			return projectTokens
		}
		break
	}
	return nil
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

	err := ioutil.WriteFile("./whitesource.config.json", npmConfig, 0644)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to write to whitesource config")
	}
	err = command.RunExecutable("yarn", "global", "add", "whitesource")
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
			log.Entry().Warn("resolveProjectIdentifiers: ScanType not implemented")
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
			config.ProductVersion = "0.0.1"
		}
	}

	if config.ProductToken == "" && config.ProductName != "" {
		product, err := sys.GetProductByName(config.ProductName)
		if err != nil {
			log.Entry().Info(fmt.Sprintf("Product %s does not yet exist", config.ProductName))
		}
		config.ProductToken = product.Token
	}

	if config.ProjectToken == "" && config.ProjectName != "" {
		fullProjectName := config.ProjectName + " - " + config.ProductVersion
		project, err := sys.GetProjectByName(config.ProductToken, fullProjectName)
		if err != nil {
			log.Entry().Info(fmt.Sprintf("Project %s does not yet exist...", fullProjectName))
		}
		config.ProjectToken = project.Token
	}
}

func getNpmGAV() map[string]string {
	result := map[string]string{}
	byteValue, err := ioutil.ReadFile("package.json")
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to read package.json")
	}
	packageJson := map[string]interface{}{}
	json.Unmarshal(byteValue, &packageJson)

	projectName := packageJson["name"].(string)
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
	result["version"] = packageJson["version"].(string)
	log.Entry().Info("Resolved NPM project version: ", result["version"])
	return result
}

func getGoGAV() map[string]string {
	result := map[string]string{}
	if fileExists("Gopkg.toml") { // Godep
		log.Entry().Fatal("Gopkg.toml parsing not implemented. exit...")
	} else if fileExists("go.mod") { // Go modules
		bytes, err := ioutil.ReadFile("go.mod")
		if err != nil {
			log.Entry().WithError(err).Fatal("Failed to read go.mod file")
		}

		m, err := modfile.Parse("go.mod", bytes, nil)
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
	if _, err := os.Stat(projectName + "-application"); !os.IsNotExist(err) {
		projectName += "-application"
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
	log.Entry().Info("Resolved gradle project version: ", result["version"])

	return result
}

// func fileExists(filename string) bool {
// 	info, err := os.Stat(filename)
// 	if os.IsNotExist(err) {
// 		return false
// 	}
// 	return !info.IsDir()
// }
