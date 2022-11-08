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
)

const (
	npmBomFilename = "bom-npm.xml"
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
	PublishAllPackages(packageJSONFiles []string, registry, username, password string, packBeforePublish bool) error
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
	err := execRunner.RunExecutable("npm", "config", "get", npmRegistry)
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
		err = execRunner.RunExecutable("npm", "config", "set", npmRegistry, exec.Options.DefaultNpmRegistry)
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
			log.Entry().Warnf("could not find any package.json file with script " + script)
			continue
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

	packageLockExists, yarnLockExists, err := exec.checkIfLockFilesExist()
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
func (exec *Execute) checkIfLockFilesExist() (bool, bool, error) {
	packageLockExists, err := exec.Utils.FileExists("package-lock.json")
	if err != nil {
		return false, false, err
	}

	yarnLockExists, err := exec.Utils.FileExists("yarn.lock")
	if err != nil {
		return false, false, err
	}
	return packageLockExists, yarnLockExists, nil
}

// CreateBOM generates BOM file using CycloneDX from all package.json files
func (exec *Execute) CreateBOM(packageJSONFiles []string) error {
	execRunner := exec.Utils.GetExecRunner()
	// Install CycloneDX Node.js module locally without saving in package.json
	err := execRunner.RunExecutable("npm", "install", "@cyclonedx/bom@^3.10.6", "--no-save")
	if err != nil {
		return err
	}

	if len(packageJSONFiles) > 0 {
		for _, packageJSONFile := range packageJSONFiles {
			path := filepath.Dir(packageJSONFile)
			params := []string{
				"cyclonedx-bom",
				path,
				"--output", filepath.Join(path, npmBomFilename),
			}
			err := execRunner.RunExecutable("npx", params...)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
