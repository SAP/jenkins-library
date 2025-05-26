package npm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

const (
	npmBomFilename                 = "bom-npm.xml"
	cycloneDxNpmPackageVersion     = "@cyclonedx/cyclonedx-npm@1.11.0"
	cycloneDxBomPackageVersion     = "@cyclonedx/bom@^3.10.6"
	cycloneDxNpmInstallationFolder = "./tmp" // This folder is also added to npmignore in publish.go.Any changes to this folder needs a change in publish.go publish()
	cycloneDxSchemaVersion         = "1.4"
)

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
	utils := utilsBundle{Files: &piperutils.Files{}, execRunner: executorOptions.ExecRunner}
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
}

type utilsBundle struct {
	*piperutils.Files
	execRunner ExecRunner
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
	excludeList = append(excludeList, nodeModulesExclude, genExclude)

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

// install executes npm or yarn Install for package.json
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

	packageLockExists, yarnLockExists, pnpmLockExists, err := exec.checkIfLockFilesExist()
	if err != nil {
		return err
	}

	log.Entry().WithField("WorkingDirectory", dir).Info("Running Install")
	if packageLockExists {
		err = execRunner.RunExecutable("npm", "ci")
		if err != nil {
			return err
		}
	} else if yarnLockExists {
		err = execRunner.RunExecutable("yarn", "install", "--frozen-lockfile")
		if err != nil {
			return err
		}
	} else if pnpmLockExists {
		err = execRunner.RunExecutable("pnpm", "install", "--frozen-lockfile")
		if err != nil {
			return err
		}
	} else {
		log.Entry().Warn("No package lock file found. " +
			"It is recommended to create a `package-lock.json` file by running `npm Install` locally." +
			" Add this file to your version control. " +
			"By doing so, the builds of your application become more reliable.")
		err = execRunner.RunExecutable("npm", "install")
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

// checkIfLockFilesExist checks if yarn/package lock fileUtils exist
func (exec *Execute) checkIfLockFilesExist() (bool, bool, bool, error) {
	packageLockExists, err := exec.Utils.FileExists("package-lock.json")
	if err != nil {
		return false, false, false, err
	}

	yarnLockExists, err := exec.Utils.FileExists("yarn.lock")
	if err != nil {
		return false, false, false, err
	}

	pnpmLockExists, err := exec.Utils.FileExists("pnpm-lock.yaml")
	if err != nil {
		return false, false, false, err
	}

	return packageLockExists, yarnLockExists, pnpmLockExists, nil

}

// CreateBOM generates BOM file using CycloneDX from all package.json files
func (exec *Execute) CreateBOM(packageJSONFiles []string) error {
	// Install cyclonedx-npm in a new folder (to avoid extraneous errors) and generate BOM
	cycloneDxNpmInstallParams := []string{"install", "--no-save", cycloneDxNpmPackageVersion, "--prefix", cycloneDxNpmInstallationFolder}
	cycloneDxNpmRunParams := []string{"--output-format", "XML", "--spec-version", cycloneDxSchemaVersion, "--omit", "dev", "--output-file"}

	// Install cyclonedx/bom with --nosave and generate BOM.
	cycloneDxBomInstallParams := []string{"install", cycloneDxBomPackageVersion, "--no-save"}
	cycloneDxBomRunParams := []string{"cyclonedx-bom", "--output"}

	// Attempt#1, generate BOM via cyclonedx-npm
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
