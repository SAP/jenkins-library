package npm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/bmatcuk/doublestar"
	"io"
	"os"
	"path"
	"strings"
)

// ExecuteOptions holds the list of scripts to be executed by ExecuteAllScripts, options for npm run and the npm registry configuration
type ExecuteOptions struct {
	Install            bool     `json:"install,omitempty"`
	RunScripts         []string `json:"runScripts,omitempty"`
	RunOptions         []string `json:"runOptions,omitempty"`
	DefaultNpmRegistry string   `json:"defaultNpmRegistry,omitempty"`
	SapNpmRegistry     string   `json:"sapNpmRegistry,omitempty"`
}

// ExecRunner interface to enable mocking for testing
type ExecRunner interface {
	Stdout(out io.Writer)
	RunExecutable(executable string, params ...string) error
}

// Utils interface for functions that need to be mocked for testing
type Utils interface {
	FileExists(path string) (bool, error)
	FileRead(path string) ([]byte, error)
	Glob(pattern string) (matches []string, err error)
	Getwd() (dir string, err error)
	Chdir(dir string) error
	GetExecRunner() ExecRunner
}

// UtilsBundle holds utils for mocking used by npmExecuteScripts/-Lint
type UtilsBundle struct {
	projectStructure FileUtils.ProjectStructure
	fileUtils        FileUtils.Files
	execRunner       *command.Command
}

// FileExists function for mock interface
func (u *UtilsBundle) FileExists(path string) (bool, error) {
	return u.fileUtils.FileExists(path)
}

// FileRead function for mock interface
func (u *UtilsBundle) FileRead(path string) ([]byte, error) {
	return u.fileUtils.FileRead(path)
}

// Glob function for mock interface
func (u *UtilsBundle) Glob(pattern string) (matches []string, err error) {
	return doublestar.Glob(pattern)
}

// Getwd function for mock interface
func (u *UtilsBundle) Getwd() (dir string, err error) {
	return os.Getwd()
}

// Chdir function for mock interface
func (u *UtilsBundle) Chdir(dir string) error {
	return os.Chdir(dir)
}

// GetExecRunner returns/creates an ExecRunner that redirects Stdout and Stderr to logging framework
func (u *UtilsBundle) GetExecRunner() ExecRunner {
	if u.execRunner == nil {
		u.execRunner = &command.Command{}
		u.execRunner.Stdout(log.Writer())
		u.execRunner.Stderr(log.Writer())
	}
	return u.execRunner
}

// SetNpmRegistries configures the given npm registries.
// CAUTION: This will change the npm configuration in the user's home directory.
func SetNpmRegistries(execRunner ExecRunner, options *ExecuteOptions) error {
	const sapRegistry = "@sap:registry"
	const npmRegistry = "registry"
	configurableRegistries := []string{npmRegistry, sapRegistry}
	for _, registry := range configurableRegistries {
		var buffer bytes.Buffer
		execRunner.Stdout(&buffer)
		err := execRunner.RunExecutable("npm", "config", "get", registry)
		execRunner.Stdout(log.Writer())
		if err != nil {
			return err
		}
		preConfiguredRegistry := buffer.String()

		if registryIsNonEmpty(preConfiguredRegistry) {
			log.Entry().Info("Discovered pre-configured npm registry " + registry + " with value " + preConfiguredRegistry)
		}

		if registry == npmRegistry && options.DefaultNpmRegistry != "" && registryRequiresConfiguration(preConfiguredRegistry, "https://registry.npmjs.org") {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + options.DefaultNpmRegistry)
			err = execRunner.RunExecutable("npm", "config", "set", registry, options.DefaultNpmRegistry)
			if err != nil {
				return err
			}
		}

		if registry == sapRegistry && options.SapNpmRegistry != "" && registryRequiresConfiguration(preConfiguredRegistry, "https://npm.sap.com") {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + options.SapNpmRegistry)
			err = execRunner.RunExecutable("npm", "config", "set", registry, options.SapNpmRegistry)
			if err != nil {
				return err
			}
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

// ExecuteAllScripts runs all scripts defined in ExecuteOptions.RunScripts
func ExecuteAllScripts(utils Utils, options ExecuteOptions) error {
	packageJSONFiles := FindPackageJSONFiles(utils)

	for _, script := range options.RunScripts {
		packagesWithScript, err := FindPackageJSONFilesWithScript(utils, packageJSONFiles, script)
		if err != nil {
			return err
		}

		if len(packagesWithScript) == 0 {
			log.Entry().Warnf("could not find any package.json file with script " + script)
			continue
		}

		for _, packageJSON := range packagesWithScript {
			err = executeScript(utils, options, packageJSON, script)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func executeScript(utils Utils, options ExecuteOptions, packageJSON string, script string) error {
	execRunner := utils.GetExecRunner()
	oldWorkingDirectory, err := utils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory before executing npm scripts: %w", err)
	}

	dir := path.Dir(packageJSON)
	err = utils.Chdir(dir)
	if err != nil {
		return fmt.Errorf("failed to change into directory for executing script: %w", err)
	}

	if !(options.Install) {
		// set in each directory to respect existing config in rc files
		err = SetNpmRegistries(execRunner, &options)
		if err != nil {
			return err
		}
	}

	log.Entry().WithField("WorkingDirectory", dir).Info("run-script " + script)

	npmRunArgs := []string{"run", script}
	if len(options.RunOptions) > 0 {
		npmRunArgs = append(npmRunArgs, options.RunOptions...)
	}

	err = execRunner.RunExecutable("npm", npmRunArgs...)
	if err != nil {
		return fmt.Errorf("failed to run npm script %s: %w", script, err)
	}

	err = utils.Chdir(oldWorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to change back into original directory: %w", err)
	}
	return nil
}

// FindPackageJSONFiles returns a list of all package.json files of the project excluding node_modules and gen/ directories
func FindPackageJSONFiles(utils Utils) []string {
	unfilteredListOfPackageJSONFiles, _ := utils.Glob("**/package.json")

	var packageJSONFiles []string

	for _, file := range unfilteredListOfPackageJSONFiles {
		if strings.Contains(file, "node_modules") {
			continue
		}

		if strings.HasPrefix(file, "gen/") || strings.Contains(file, "/gen/") {
			continue
		}

		packageJSONFiles = append(packageJSONFiles, file)
		log.Entry().Info("Discovered package.json file " + file)
	}
	return packageJSONFiles
}

// FindPackageJSONFilesWithScript returns a list of package.json files that contain the script
func FindPackageJSONFilesWithScript(utils Utils, packageJSONFiles []string, script string) ([]string, error) {
	var packagesWithScript []string

	for _, file := range packageJSONFiles {
		var packageJSON map[string]interface{}

		packageRaw, err := utils.FileRead(file)
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

// InstallAllDependencies executes npm or yarn install for all package.json files defined in packageJSONFiles
func InstallAllDependencies(packageJSONFiles []string, utils Utils, options *ExecuteOptions) error {
	for _, packageJSON := range packageJSONFiles {
		err := Install(utils, packageJSON, options)
		if err != nil {
			return err
		}
	}
	return nil
}

// Install executes npm or yarn install for package.json
func Install(utils Utils, packageJSON string, options *ExecuteOptions) error {
	execRunner := utils.GetExecRunner()

	oldWorkingDirectory, err := utils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory before executing npm scripts: %w", err)
	}

	dir := path.Dir(packageJSON)
	err = utils.Chdir(dir)
	if err != nil {
		return fmt.Errorf("failed to change into directory for executing script: %w", err)
	}

	err = SetNpmRegistries(utils.GetExecRunner(), options)
	if err != nil {
		return err
	}

	packageLockExists, yarnLockExists, err := checkIfLockFilesExist(utils)
	if err != nil {
		return err
	}

	log.Entry().WithField("WorkingDirectory", dir).Info("Running install")
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
			"It is recommended to create a `package-lock.json` file by running `npm install` locally." +
			" Add this file to your version control. " +
			"By doing so, the builds of your application become more reliable.")
		err = execRunner.RunExecutable("npm", "install")
		if err != nil {
			return err
		}
	}

	err = utils.Chdir(oldWorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to change back into original directory: %w", err)
	}
	return nil
}

// checkIfLockFilesExist checks if yarn/package lock files exist
func checkIfLockFilesExist(utils Utils) (bool, bool, error) {
	packageLockExists, err := utils.FileExists("package-lock.json")
	if err != nil {
		return false, false, err
	}

	yarnLockExists, err := utils.FileExists("yarn.lock")
	if err != nil {
		return false, false, err
	}
	return packageLockExists, yarnLockExists, nil
}
