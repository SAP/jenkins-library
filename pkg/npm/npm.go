package npm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

const (
	npmBomFilename                 = "bom-npm.xml"
	tempBomFilename                = "bom-npm.json"
	cycloneDxNpmPackageVersion     = "@cyclonedx/cyclonedx-npm@1.11.0"
	cycloneDxBomPackageVersion     = "@cyclonedx/bom@^3.10.6"
	cdxgenPackageVersion           = "@cyclonedx/cdxgen@^11.3.2"
	cycloneDxCliVersion            = "v0.27.2"
	cycloneDxNpmInstallationFolder = "./tmp" // This folder is also added to npmignore in publish.go.Any changes to this folder needs a change in publish.go publish()
	cycloneDxSchemaVersion         = "1.4"
)

var cycloneDxCliUrl = map[struct{ os, arch string }]string{
	{"windows", "amd64"}: "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-win-x64.exe",
	{"windows", "arm64"}: "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-win-arm64.exe",
	{"darwin", "amd64"}:  "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-osx-x64",
	{"darwin", "arm64"}:  "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-osx-arm64",
	{"linux", "amd64"}:   "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-linux-x64",
	{"linux", "arm64"}:   "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-linux-arm64",
	{"linux", "arm"}:     "https://github.com/CycloneDX/cyclonedx-cli/releases/download/%s/cyclonedx-linux-arm",
}

// Execute struct holds utils to enable mocking and common parameters
type Execute struct {
	Utils   Utils
	Options ExecutorOptions
}

// Executor interface to enable mocking for testing
type Executor interface {
	FindPackageJSONFiles() []string
	FindPackageJSONFilesWithExcludes(excludeList []string) ([]string, error)
	FindPackageJSONFilesWithScript(packageJSONFiles []string, script string) ([]string, error)
	RunScriptsInAllPackages(runScripts []string, runOptions []string, scriptOptions []string, virtualFrameBuffer bool, excludeList []string, packagesList []string) error
	InstallAllDependencies(packageJSONFiles []string) error
	PublishAllPackages(packageJSONFiles []string, registry, username, password string, packBeforePublish bool, buildCoordinates *[]versioning.Coordinates) error
	SetNpmRegistries() error
	CreateBOM(packageJSONFiles []string) error
}

// ExecutorOptions holds common parameters for functions of Executor
type ExecutorOptions struct {
	DefaultNpmRegistry string
	ExecRunner         ExecRunner
}

// NewExecutor instantiates Execute struct and sets executeOptions
func NewExecutor(executorOptions ExecutorOptions) Executor {
	utils := utilsBundle{
		Files:          &piperutils.Files{},
		execRunner:     executorOptions.ExecRunner,
		downloadClient: &piperhttp.Client{},
	}
	return &Execute{
		Utils:   &utils,
		Options: executorOptions,
	}
}

// ExecRunner interface to enable mocking for testing
type ExecRunner interface {
	SetEnv(e []string)
	Stdout(out io.Writer)
	Stderr(out io.Writer)
	RunExecutable(executable string, params ...string) error
	RunExecutableInBackground(executable string, params ...string) (command.Execution, error)
}

// Utils interface for mocking
type Utils interface {
	piperutils.FileUtils

	GetExecRunner() ExecRunner
	GetFileUtils() piperutils.FileUtils
	GetDownloadUtils() piperhttp.Downloader
}

type utilsBundle struct {
	*piperutils.Files
	execRunner     ExecRunner
	downloadClient piperhttp.Downloader
}

func (u *utilsBundle) GetFileUtils() piperutils.FileUtils {
	return u.Files
}

func (u *utilsBundle) GetDownloadUtils() piperhttp.Downloader {
	if u.downloadClient == nil {
		u.downloadClient = &piperhttp.Client{}
	}
	return u.downloadClient
}

// GetExecRunner returns an execRunner if it's not yet initialized
func (u *utilsBundle) GetExecRunner() ExecRunner {
	if u.execRunner == nil {
		u.execRunner = &command.Command{
			StepName: "npmExecuteScripts",
		}
		u.execRunner.Stdout(log.Writer())
		u.execRunner.Stderr(log.Writer())
	}
	return u.execRunner
}

// SetNpmRegistries configures the given npm registries.
// CAUTION: This will change the npm configuration in the user's home directory.
func (exec *Execute) SetNpmRegistries() error {
	execRunner := exec.Utils.GetExecRunner()
	const npmRegistry = "registry"

	var buffer bytes.Buffer
	execRunner.Stdout(&buffer)
	err := execRunner.RunExecutable("npm", "config", "get", npmRegistry, "-ws=false", "-iwr")
	execRunner.Stdout(log.Writer())
	if err != nil {
		return err
	}
	preConfiguredRegistry := buffer.String()

	if registryIsNonEmpty(preConfiguredRegistry) {
		log.Entry().Info("Discovered pre-configured npm registry " + npmRegistry + " with value " + preConfiguredRegistry)
	}

	if exec.Options.DefaultNpmRegistry != "" && registryRequiresConfiguration(preConfiguredRegistry, "https://registry.npmjs.org") {
		log.Entry().Info("npm registry " + npmRegistry + " was not configured, setting it to " + exec.Options.DefaultNpmRegistry)
		err = execRunner.RunExecutable("npm", "config", "set", npmRegistry, exec.Options.DefaultNpmRegistry, "-ws=false", "-iwr")
		if err != nil {
			return err
		}
	}

	return nil
}

func registryIsNonEmpty(preConfiguredRegistry string) bool {
	return !strings.HasPrefix(preConfiguredRegistry, "undefined") && len(preConfiguredRegistry) > 0
}

func registryRequiresConfiguration(preConfiguredRegistry, url string) bool {
	return strings.HasPrefix(preConfiguredRegistry, "undefined") || strings.HasPrefix(preConfiguredRegistry, url)
}

// RunScriptsInAllPackages runs all scripts defined in ExecuteOptions.RunScripts
func (exec *Execute) RunScriptsInAllPackages(runScripts []string, runOptions []string, scriptOptions []string, virtualFrameBuffer bool, excludeList []string, packagesList []string) error {
	var packageJSONFiles []string
	var err error

	if len(packagesList) > 0 {
		packageJSONFiles = packagesList
	} else {
		packageJSONFiles, err = exec.FindPackageJSONFilesWithExcludes(excludeList)
		if err != nil {
			return err
		}
	}

	execRunner := exec.Utils.GetExecRunner()

	if virtualFrameBuffer {
		cmd, err := execRunner.RunExecutableInBackground("Xvfb", "-ac", ":99", "-screen", "0", "1280x1024x16")
		if err != nil {
			return fmt.Errorf("failed to start virtual frame buffer%w", err)
		}
		defer cmd.Kill()
		execRunner.SetEnv([]string{"DISPLAY=:99"})
	}

	for _, script := range runScripts {
		packagesWithScript, err := exec.FindPackageJSONFilesWithScript(packageJSONFiles, script)
		if err != nil {
			return err
		}

		if len(packagesWithScript) == 0 {
			return fmt.Errorf("could not find any package.json file with script : %s ", script)

		}

		for _, packageJSON := range packagesWithScript {
			err = exec.executeScript(packageJSON, script, runOptions, scriptOptions)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (exec *Execute) executeScript(packageJSON string, script string, runOptions []string, scriptOptions []string) error {
	execRunner := exec.Utils.GetExecRunner()
	oldWorkingDirectory, err := exec.Utils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory before executing npm scripts: %w", err)
	}

	dir := filepath.Dir(packageJSON)
	err = exec.Utils.Chdir(dir)
	if err != nil {
		return fmt.Errorf("failed to change into directory for executing script: %w", err)
	}

	// set in each directory to respect existing config in rc fileUtils
	err = exec.SetNpmRegistries()
	if err != nil {
		return err
	}

	log.Entry().WithField("WorkingDirectory", dir).Info("run-script " + script)

	npmRunArgs := []string{"run", script}
	if len(runOptions) > 0 {
		npmRunArgs = append(npmRunArgs, runOptions...)
	}

	if len(scriptOptions) > 0 {
		npmRunArgs = append(npmRunArgs, "--")
		npmRunArgs = append(npmRunArgs, scriptOptions...)
	}

	err = execRunner.RunExecutable("npm", npmRunArgs...)
	if err != nil {
		return fmt.Errorf("failed to run npm script %s: %w", script, err)
	}

	err = exec.Utils.Chdir(oldWorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to change back into original directory: %w", err)
	}
	return nil
}

// FindPackageJSONFiles returns a list of all package.json files of the project excluding node_modules and gen/ directories
func (exec *Execute) FindPackageJSONFiles() []string {
	packageJSONFiles, _ := exec.FindPackageJSONFilesWithExcludes([]string{})
	return packageJSONFiles
}

// FindPackageJSONFilesWithExcludes returns a list of all package.json files of the project excluding node_modules, gen/ and directories/patterns defined by excludeList
func (exec *Execute) FindPackageJSONFilesWithExcludes(excludeList []string) ([]string, error) {
	unfilteredListOfPackageJSONFiles, _ := exec.Utils.Glob("**/package.json")

	nodeModulesExclude := "**/node_modules/**"
	genExclude := "**/gen/**"
	tmpExclude := "**/tmp/**"
	excludeList = append(excludeList, nodeModulesExclude, genExclude, tmpExclude)

	packageJSONFiles, err := piperutils.ExcludeFiles(unfilteredListOfPackageJSONFiles, excludeList)
	if err != nil {
		return nil, err
	}

	for _, file := range packageJSONFiles {
		log.Entry().Info("Discovered package.json file " + file)
	}
	return packageJSONFiles, nil
}

// FindPackageJSONFilesWithScript returns a list of package.json fileUtils that contain the script
func (exec *Execute) FindPackageJSONFilesWithScript(packageJSONFiles []string, script string) ([]string, error) {
	var packagesWithScript []string

	for _, file := range packageJSONFiles {
		var packageJSON map[string]interface{}

		packageRaw, err := exec.Utils.FileRead(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s to check for existence of %s script: %w", file, script, err)
		}

		err = json.Unmarshal(packageRaw, &packageJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s to check for existence of %s script: %w", file, script, err)
		}

		scripts, ok := packageJSON["scripts"].(map[string]interface{})
		if ok {
			_, ok := scripts[script].(string)
			if ok {
				packagesWithScript = append(packagesWithScript, file)
				log.Entry().Info("Discovered " + script + " script in " + file)
			}
		}
	}
	return packagesWithScript, nil
}

// InstallAllDependencies executes npm or yarn Install for all package.json fileUtils defined in packageJSONFiles
func (exec *Execute) InstallAllDependencies(packageJSONFiles []string) error {
	for _, packageJSON := range packageJSONFiles {
		fileExists, err := exec.Utils.FileExists(packageJSON)
		if err != nil {
			return fmt.Errorf("cannot check if '%s' exists: %w", packageJSON, err)
		}
		if !fileExists {
			return fmt.Errorf("package.json file '%s' not found: %w", packageJSON, err)
		}

		err = exec.install(packageJSON)
		if err != nil {
			return err
		}
	}
	return nil
}

// install executes the appropriate package manager install command for package.json
func (exec *Execute) install(packageJSON string) error {
	execRunner := exec.Utils.GetExecRunner()

	oldWorkingDirectory, err := exec.Utils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory before executing npm scripts: %w", err)
	}

	dir := filepath.Dir(packageJSON)
	err = exec.Utils.Chdir(dir)
	if err != nil {
		return fmt.Errorf("failed to change into directory for executing script: %w", err)
	}

	err = exec.SetNpmRegistries()
	if err != nil {
		return err
	}

	pm, err := exec.detectPackageManager()
	if err != nil {
		return err
	}

	log.Entry().WithField("WorkingDirectory", dir).Info("Running Install")

	if pm.Name == "pnpm" {
		// Special handling for pnpm since it requires local installation
		commands := [][]string{
			{"mkdir", "-p", "./tmp/pnpm-bin"},
			{"npm", "install", "pnpm", "--prefix", "./tmp/pnpm-bin"},
			{"./tmp/pnpm-bin/node_modules/.bin/pnpm", "install", "--frozen-lockfile"},
		}

		for _, cmd := range commands {
			if err := execRunner.RunExecutable(cmd[0], cmd[1:]...); err != nil {
				return err
			}
		}
	} else {
		if !strings.HasPrefix(pm.LockFile, "package-lock.json") {
			log.Entry().Info("Using " + pm.Name + " package manager")
		}

		err = execRunner.RunExecutable(pm.InstallCommand, pm.InstallArgs...)
		if err != nil {
			return err
		}
	}

	err = exec.Utils.Chdir(oldWorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to change back into original directory: %w", err)
	}
	return nil
}

// CreateBOM generates BOM file using CycloneDX from all package.json files
func (exec *Execute) CreateBOM(packageJSONFiles []string) error {
	// Detect package manager to determine which BOM generator to use
	pm, err := exec.detectPackageManager()
	if err != nil {
		return err
	}

	// Install cyclonedx-npm in a new folder (to avoid extraneous errors) and generate BOM
	cycloneDxNpmInstallParams := []string{"install", "--no-save", cycloneDxNpmPackageVersion, "--prefix", cycloneDxNpmInstallationFolder}
	cycloneDxNpmRunParams := []string{"--output-format", "XML", "--spec-version", cycloneDxSchemaVersion, "--omit", "dev", "--output-file"}

	// Install cyclonedx/bom with --nosave and generate BOM.
	cycloneDxBomInstallParams := []string{"install", cycloneDxBomPackageVersion, "--no-save"}
	cycloneDxBomRunParams := []string{"cyclonedx-bom", "--output"}

	// Attempt#1, generate BOM via cyclonedx-npm
	// Check for pnpm first since it uses cdxgen
	if pm != nil && pm.Name == "pnpm" {
		// Download cyclonedx-cli
		osArch := struct{ os, arch string }{runtime.GOOS, runtime.GOARCH}
		var cliPath string
		var err error
		if urlTemplate, ok := cycloneDxCliUrl[osArch]; ok {
			url := fmt.Sprintf(urlTemplate, cycloneDxCliVersion)
			cliPath, err = piperhttp.DownloadExecutable("", exec.Utils.GetFileUtils(), exec.Utils.GetDownloadUtils(), url)
			if err != nil {
				return fmt.Errorf("failed to download cyclonedx-cli: %w", err)
			}
		} else {
			return fmt.Errorf("cyclonedx-cli not available for OS/architecture combination: %s/%s", runtime.GOOS, runtime.GOARCH)
		}

		// Install cdxgen for pnpm projects
		execRunner := exec.Utils.GetExecRunner()
		err = execRunner.RunExecutable("npm", "install", cdxgenPackageVersion, "--prefix", cycloneDxNpmInstallationFolder)
		if err != nil {
			return fmt.Errorf("failed to install cdxgen: %w", err)
		}

		for _, packageJSONFile := range packageJSONFiles {
			path := filepath.Dir(packageJSONFile)
			jsonBomPath := filepath.Join(path, tempBomFilename)
			xmlBomPath := filepath.Join(path, npmBomFilename)

			// Generate JSON SBOM with cdxgen
			cdxgenExecutable := cycloneDxNpmInstallationFolder + "/node_modules/.bin/cdxgen"
			params := []string{
				"-r",
				"-o", jsonBomPath,
				"--spec-version", cycloneDxSchemaVersion,
			}
			err := execRunner.RunExecutable(cdxgenExecutable, params...)
			if err != nil {
				return fmt.Errorf("failed to generate CycloneDX BOM with cdxgen: %w", err)
			}

			log.Entry().Infof("Generated CycloneDX BOM in JSON format at %s", jsonBomPath)

			// Convert JSON to XML using cyclonedx-cli
			err = execRunner.RunExecutable(cliPath, "convert", "--input-file", jsonBomPath, "--output-format", "xml", "--output-file", xmlBomPath)
			if err != nil {
				return fmt.Errorf("failed to convert BOM to XML format: %w", err)
			}

			log.Entry().Infof("Converted CycloneDX BOM to XML format at %s", xmlBomPath)
		}
	} else {
		// For non-pnpm projects, use existing CycloneDX approach
		err := exec.createBOMWithParams(cycloneDxNpmInstallParams, cycloneDxNpmRunParams, packageJSONFiles, false)
		if err != nil {
			log.Entry().Infof("Failed to generate BOM CycloneDX BOM with cyclonedx-npm ,fallback to cyclonedx/bom")

			// Attempt #2, generate BOM via cyclonedx/bom@^3.10.6
			err = exec.createBOMWithParams(cycloneDxBomInstallParams, cycloneDxBomRunParams, packageJSONFiles, true)
			if err != nil {
				log.Entry().Infof("Failed to generate BOM CycloneDX BOM with fallback package cyclonedx/bom ")
				return err
			}
		}
	}
	return nil
}

// Facilitates BOM generation with different packages
func (exec *Execute) createBOMWithParams(packageInstallParams []string, packageRunParams []string, packageJSONFiles []string, fallback bool) error {
	execRunner := exec.Utils.GetExecRunner()

	// Install package
	err := execRunner.RunExecutable("npm", packageInstallParams...)
	if err != nil {
		return fmt.Errorf("failed to install CycloneDX BOM %w", err)
	}

	// Run package for all package JSON files
	if len(packageJSONFiles) > 0 {
		for _, packageJSONFile := range packageJSONFiles {
			path := filepath.Dir(packageJSONFile)
			executable := "npx"
			params := append(packageRunParams, filepath.Join(path, npmBomFilename))

			// Below code needed as to adjust according to needs of cyclonedx-npm and fallback cyclonedx/bom@^3.10.6
			if !fallback {
				params = append(params, packageJSONFile)
				executable = cycloneDxNpmInstallationFolder + "/node_modules/.bin/cyclonedx-npm"
			} else {
				params = append(params, path)
			}

			err := execRunner.RunExecutable(executable, params...)
			if err != nil {
				return fmt.Errorf("failed to generate CycloneDX BOM :%w", err)
			}
		}
	}

	return nil
}
