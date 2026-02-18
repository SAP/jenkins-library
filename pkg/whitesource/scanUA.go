package whitesource

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

const jvmTarGz = "jvm.tar.gz"
const jvmDir = "./jvm"
const projectRegEx = `Project name: ([^,]*), URL: (.*)`

// ExecuteUAScan executes a scan with the Whitesource Unified Agent.
func (s *Scan) ExecuteUAScan(config *ScanOptions, utils Utils) error {
	s.AgentName = "WhiteSource Unified Agent"

	switch config.BuildTool {
	case "mta":
		return s.ExecuteUAScanForMTA(config, utils)
	case "npm":
		if config.DisableNpmSubmodulesAggregation {
			return s.ExecuteUAScanForMultiModuleNPM(config, utils)
		} else {
			return s.ExecuteUAScanInPath(config, utils, config.ScanPath)
		}
	default:
		return s.ExecuteUAScanInPath(config, utils, config.ScanPath)
	}

}

func (s *Scan) ExecuteUAScanForMTA(config *ScanOptions, utils Utils) error {
	log.Entry().Infof("Executing WhiteSource UA scan for MTA project")

	log.Entry().Infof("Executing WhiteSource UA scan for Maven part")
	pomExists, _ := utils.FileExists("pom.xml")
	if pomExists {
		mavenConfig := *config
		mavenConfig.BuildTool = "maven"
		if err := s.ExecuteUAScanInPath(&mavenConfig, utils, config.ScanPath); err != nil {
			return fmt.Errorf("failed to run scan for maven modules of mta: %w", err)
		}
	} else {
		if pomFiles, _ := utils.Glob("**/pom.xml"); len(pomFiles) > 0 {
			log.SetErrorCategory(log.ErrorCustom)
			return fmt.Errorf("mta project with java modules does not contain an aggregator pom.xml in the root - this is mandatory")
		}
	}

	log.Entry().Infof("Executing WhiteSource UA scan for NPM part")
	return s.ExecuteUAScanForMultiModuleNPM(config, utils)
}

func (s *Scan) ExecuteUAScanForMultiModuleNPM(config *ScanOptions, utils Utils) error {
	log.Entry().Infof("Executing WhiteSource UA scan for multi-module NPM projects")

	packageJSONFiles, err := utils.FindPackageJSONFiles(config)
	if err != nil {
		return fmt.Errorf("failed to find package.json files: %w", err)
	}
	if len(packageJSONFiles) > 0 {
		npmConfig := *config
		npmConfig.BuildTool = "npm"
		for _, packageJSONFile := range packageJSONFiles {
			// we only need the path here
			modulePath, _ := filepath.Split(packageJSONFile)
			projectName, err := getProjectNameFromPackageJSON(packageJSONFile, utils)
			if err != nil {
				return fmt.Errorf("failed retrieve project name: %w", err)
			}
			npmConfig.ProjectName = projectName
			// ToDo: likely needs to be refactored, AggregateProjectName should only be available if we want to force aggregation?
			s.AggregateProjectName = projectName
			if err := s.ExecuteUAScanInPath(&npmConfig, utils, modulePath); err != nil {
				return fmt.Errorf("failed to run scan for npm module %v: %w", modulePath, err)
			}
		}
	}

	_ = removeJre(filepath.Join(jvmDir, "bin", "java"), utils)

	return nil
}

// ExecuteUAScanInPath executes a scan with the Whitesource Unified Agent in a dedicated scanPath.
func (s *Scan) ExecuteUAScanInPath(config *ScanOptions, utils Utils, scanPath string) error {
	// Download the unified agent jar file if one does not exist
	err := downloadAgent(config, utils)
	if err != nil {
		return err
	}

	// Download JRE in case none is available
	javaPath, err := downloadJre(config, utils)
	if err != nil {
		return err
	}

	// Fetch version of UA
	versionBuffer := bytes.Buffer{}
	utils.Stdout(&versionBuffer)
	err = utils.RunExecutable(javaPath, "-jar", config.AgentFileName, "-v")
	if err != nil {
		return fmt.Errorf("Failed to determine UA version: %w", err)
	}
	s.AgentVersion = strings.TrimSpace(versionBuffer.String())
	log.Entry().Debugf("Read UA version %v from Stdout", s.AgentVersion)
	utils.Stdout(log.Writer())

	// ToDo: Check if Download of Docker/container image should be done here instead of in cmd/whitesourceExecuteScan.go

	// ToDo: check if this is required
	if !config.SkipParentProjectResolution {
		if err := s.AppendScannedProject(s.AggregateProjectName); err != nil {
			return err
		}
	}

	if config.UseGlobalConfiguration {
		config.ConfigFilePath, err = filepath.Abs(config.ConfigFilePath)
		if err != nil {
			return err
		}
	}

	configPath, err := config.RewriteUAConfigurationFile(utils, s.AggregateProjectName, config.Verbose)
	if err != nil {
		return err
	}

	if len(scanPath) == 0 {
		scanPath = "."
	}

	// log parsing in order to identify the projects WhiteSource really scanned
	// we may refactor this in case there is a safer way to identify the projects e.g. via REST API

	//ToDO: we only need stdOut or stdErr, let's see where UA writes to ...
	prOut, stdOut := io.Pipe()
	trOut := io.TeeReader(prOut, os.Stderr)
	utils.Stdout(stdOut)

	prErr, stdErr := io.Pipe()
	trErr := io.TeeReader(prErr, os.Stderr)
	utils.Stdout(stdErr)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		scanLog(trOut, s)
	}()

	go func() {
		defer wg.Done()
		scanLog(trErr, s)
	}()
	err = utils.RunExecutable(javaPath, "-jar", config.AgentFileName, "-d", scanPath, "-c", configPath, "-wss.url", config.AgentURL)

	if err := removeJre(javaPath, utils); err != nil {
		log.Entry().Warning(err)
	}

	if err != nil {
		if err := removeJre(javaPath, utils); err != nil {
			log.Entry().Warning(err)
		}
		exitCode := utils.GetExitCode()
		log.Entry().Infof("WhiteSource scan failed with exit code %v", exitCode)
		evaluateExitCode(exitCode)
		return fmt.Errorf("failed to execute WhiteSource scan with exit code %v: %w", exitCode, err)
	}
	return nil
}

func evaluateExitCode(exitCode int) {
	switch exitCode {
	case 255:
		log.Entry().Info("General error has occurred.")
		log.SetErrorCategory(log.ErrorUndefined)
	case 254:
		log.Entry().Info("Whitesource found one or multiple policy violations.")
		log.SetErrorCategory(log.ErrorCompliance)
	case 253:
		log.Entry().Info("The local scan client failed to execute the scan.")
		log.SetErrorCategory(log.ErrorUndefined)
	case 252:
		log.Entry().Info("There was a failure in the connection to the WhiteSource servers.")
		log.SetErrorCategory(log.ErrorInfrastructure)
	case 251:
		log.Entry().Info("The server failed to analyze the scan.")
		log.SetErrorCategory(log.ErrorService)
	case 250:
		log.Entry().Info("One of the package manager's prerequisite steps (e.g. npm install) failed.")
		log.SetErrorCategory(log.ErrorCustom)
	default:
		log.Entry().Info("Whitesource scan failed with unknown error code")
		log.SetErrorCategory(log.ErrorUndefined)
	}
}

// downloadAgent downloads the unified agent jar file if one does not exist
func downloadAgent(config *ScanOptions, utils Utils) error {
	agentFile := config.AgentFileName
	exists, err := utils.FileExists(agentFile)
	if err != nil {
		return fmt.Errorf("failed to check if file '%s' exists: %w", agentFile, err)
	}
	if !exists {
		err := utils.DownloadFile(config.AgentDownloadURL, agentFile, nil, nil)
		if err != nil {
			// we check if the copy and the unauthorized error occurs and retry the download
			// if the copy error did not happen, we rerun the whole download mechanism once
			if strings.Contains(err.Error(), "unable to copy content from url to file") || strings.Contains(err.Error(), "returned with response 404 Not Found") || strings.Contains(err.Error(), "returned with response 403 Forbidden") {
				// retry the download once again
				log.Entry().Warnf("[Retry] Previous download failed due to %v", err)
				err = utils.DownloadFile(config.AgentDownloadURL, agentFile, nil, nil)
			}
		}

		if err != nil {
			return fmt.Errorf("failed to download unified agent from URL '%s' to file '%s': %w", config.AgentDownloadURL, agentFile, err)
		}
	}
	return nil
}

// downloadJre downloads the a JRE in case no java command can be executed
func downloadJre(config *ScanOptions, utils Utils) (string, error) {
	// cater for multiple executions
	if exists, _ := utils.FileExists(filepath.Join(jvmDir, "bin", "java")); exists {
		return filepath.Join(jvmDir, "bin", "java"), nil
	}
	err := utils.RunExecutable("java", "-version")
	javaPath := "java"
	if err != nil {
		log.Entry().Infof("No Java installation found, downloading JVM from %v", config.JreDownloadURL)
		const maxRetries = 3
		retries := 0
		for retries < maxRetries {
			err = utils.DownloadFile(config.JreDownloadURL, jvmTarGz, nil, nil)
			if err == nil {
				break
			}
			log.Entry().Warnf("Attempt %d: Download failed due to %v", retries+1, err)
			retries++
			if retries >= maxRetries {
				log.Entry().Errorf("Download failed after %d attempts", retries)
				return "", fmt.Errorf("failed to download jre from URL '%s': %w", config.JreDownloadURL, err)
			}
			time.Sleep(1 * time.Second)
		}

		if err != nil {
			return "", fmt.Errorf("Even after retry failed to download jre from URL '%s': %w", config.JreDownloadURL, err)
		}

		// ToDo: replace tar call with go library call
		err = utils.MkdirAll(jvmDir, 0755)

		err = utils.RunExecutable("tar", fmt.Sprintf("--directory=%v", jvmDir), "--strip-components=1", "-xzf", jvmTarGz)
		if err != nil {
			return "", fmt.Errorf("failed to extract %v: %w", jvmTarGz, err)
		}
		log.Entry().Info("Java successfully installed")
		javaPath = filepath.Join(jvmDir, "bin", "java")
	}
	return javaPath, nil
}

func removeJre(javaPath string, utils Utils) error {
	if javaPath == "java" {
		return nil
	}
	if err := utils.RemoveAll(jvmDir); err != nil {
		return fmt.Errorf("failed to remove downloaded and extracted jvm from %v", jvmDir)
	}
	log.Entry().Debugf("Java successfully removed from %v", jvmDir)
	if err := utils.FileRemove(jvmTarGz); err != nil {
		return fmt.Errorf("failed to remove downloaded %v", jvmTarGz)
	}
	log.Entry().Debugf("%v successfully removed", jvmTarGz)
	return nil
}

func getProjectNameFromPackageJSON(packageJSONPath string, utils Utils) (string, error) {
	fileContents, err := utils.FileRead(packageJSONPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %v: %w", packageJSONPath, err)
	}
	var packageJSON = make(map[string]any)
	if err := json.Unmarshal(fileContents, &packageJSON); err != nil {
		return "", fmt.Errorf("failed to read file content of %v: %w", packageJSONPath, err)
	}

	projectNameEntry, exists := packageJSON["name"]
	if !exists {
		return "", fmt.Errorf("the file '%s' must configure a name", packageJSONPath)
	}

	projectName, isString := projectNameEntry.(string)
	if !isString {
		return "", fmt.Errorf("the file '%s' must configure a name as string", packageJSONPath)
	}

	return projectName, nil
}

func scanLog(in io.Reader, scan *Scan) {
	scanner := bufio.NewScanner(in)
	scanner.Split(scanShortLines)
	for scanner.Scan() {
		line := scanner.Text()
		parseForProjects(line, scan)
	}
	if err := scanner.Err(); err != nil {
		log.Entry().WithError(err).Info("failed to scan log file")
	}
}

func parseForProjects(logLine string, scan *Scan) {
	compile := regexp.MustCompile(projectRegEx)
	values := compile.FindStringSubmatch(logLine)

	if len(values) > 0 && scan.scannedProjects != nil && len(scan.scannedProjects[values[1]].Name) == 0 {
		scan.scannedProjects[values[1]] = Project{Name: values[1]}
	}

}

func scanShortLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	lenData := len(data)
	if atEOF && lenData == 0 {
		return 0, nil, nil
	}
	if lenData > 32767 && !bytes.Contains(data[0:lenData], []byte("\n")) {
		// we will neglect long output
		// no use cases known where this would be relevant
		return lenData, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 && i < 32767 {
		// We have a full newline-terminated line with a size limit
		// Size limit is required since otherwise scanner would stall
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
