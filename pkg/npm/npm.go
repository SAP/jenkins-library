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

type execute struct {
	utils   utils
	options executeOptions
}

// Executor interface to enable mocking for testing
type Executor interface {
	FindPackageJSONFiles() []string
	FindPackageJSONFilesWithScript(packageJSONFiles []string, script string) ([]string, error)
	ExecuteAllScripts() error
	InstallAllDependencies(packageJSONFiles []string) error
	SetNpmRegistries() error
}

// NewExecutor instantiates execute struct and sets executeOptions
func NewExecutor(installDeps bool, runScripts []string, runOptions []string, defaultNpmRegistry string, sapNpmRegistry string) (*execute, error) {
	utils := utilsBundle{}
	options := executeOptions{
		install:            installDeps,
		runScripts:         runScripts,
		runOptions:         runOptions,
		defaultNpmRegistry: defaultNpmRegistry,
		sapNpmRegistry:     sapNpmRegistry,
	}
	exec := &execute{
		utils:   &utils,
		options: options,
	}
	return exec, nil
}

// ExecuteOptions holds the list of scripts to be executed by ExecuteAllScripts, options for npm run and the npm registry configuration
type executeOptions struct {
	install            bool
	runScripts         []string
	runOptions         []string
	defaultNpmRegistry string
	sapNpmRegistry     string
}

// execRunner interface to enable mocking for testing
type execRunner interface {
	Stdout(out io.Writer)
	RunExecutable(executable string, params ...string) error
}

type utils interface {
	fileExists(path string) (bool, error)
	fileRead(path string) ([]byte, error)
	glob(pattern string) (matches []string, err error)
	getwd() (dir string, err error)
	chdir(dir string) error
	getExecRunner() execRunner
}

type utilsBundle struct {
	fileUtils  FileUtils.Files
	execRunner *command.Command
}

func (u *utilsBundle) fileExists(path string) (bool, error) {
	return u.fileUtils.FileExists(path)
}

func (u *utilsBundle) fileRead(path string) ([]byte, error) {
	return u.fileUtils.FileRead(path)
}

func (u *utilsBundle) glob(pattern string) (matches []string, err error) {
	return doublestar.Glob(pattern)
}

func (u *utilsBundle) getwd() (dir string, err error) {
	return os.Getwd()
}

func (u *utilsBundle) chdir(dir string) error {
	return os.Chdir(dir)
}

func (u *utilsBundle) getExecRunner() execRunner {
	if u.execRunner == nil {
		u.execRunner = &command.Command{}
		u.execRunner.Stdout(log.Writer())
		u.execRunner.Stderr(log.Writer())
	}
	return u.execRunner
}

// SetNpmRegistries configures the given npm registries.
// CAUTION: This will change the npm configuration in the user's home directory.
func (exec *execute) SetNpmRegistries() error {
	execRunner := exec.utils.getExecRunner()
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

		if registry == npmRegistry && exec.options.defaultNpmRegistry != "" && registryRequiresConfiguration(preConfiguredRegistry, "https://registry.npmjs.org") {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + exec.options.defaultNpmRegistry)
			err = execRunner.RunExecutable("npm", "config", "set", registry, exec.options.defaultNpmRegistry)
			if err != nil {
				return err
			}
		}

		if registry == sapRegistry && exec.options.sapNpmRegistry != "" && registryRequiresConfiguration(preConfiguredRegistry, "https://npm.sap.com") {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + exec.options.sapNpmRegistry)
			err = execRunner.RunExecutable("npm", "config", "set", registry, exec.options.sapNpmRegistry)
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

// ExecuteAllScripts runs all scripts defined in ExecuteOptions.runScripts
func (exec *execute) ExecuteAllScripts() error {
	packageJSONFiles := exec.FindPackageJSONFiles()

	for _, script := range exec.options.runScripts {
		packagesWithScript, err := exec.FindPackageJSONFilesWithScript(packageJSONFiles, script)
		if err != nil {
			return err
		}

		if len(packagesWithScript) == 0 {
			log.Entry().Warnf("could not find any package.json file with script " + script)
			continue
		}

		for _, packageJSON := range packagesWithScript {
			err = exec.executeScript(packageJSON, script)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (exec *execute) executeScript(packageJSON string, script string) error {
	execRunner := exec.utils.getExecRunner()
	oldWorkingDirectory, err := exec.utils.getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory before executing npm scripts: %w", err)
	}

	dir := path.Dir(packageJSON)
	err = exec.utils.chdir(dir)
	if err != nil {
		return fmt.Errorf("failed to change into directory for executing script: %w", err)
	}

	if !(exec.options.install) {
		// set in each directory to respect existing config in rc fileUtils
		err = exec.SetNpmRegistries()
		if err != nil {
			return err
		}
	}

	log.Entry().WithField("WorkingDirectory", dir).Info("run-script " + script)

	npmRunArgs := []string{"run", script}
	if len(exec.options.runOptions) > 0 {
		npmRunArgs = append(npmRunArgs, exec.options.runOptions...)
	}

	err = execRunner.RunExecutable("npm", npmRunArgs...)
	if err != nil {
		return fmt.Errorf("failed to run npm script %s: %w", script, err)
	}

	err = exec.utils.chdir(oldWorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to change back into original directory: %w", err)
	}
	return nil
}

// FindPackageJSONFiles returns a list of all package.json fileUtils of the project excluding node_modules and gen/ directories
func (exec *execute) FindPackageJSONFiles() []string {
	unfilteredListOfPackageJSONFiles, _ := exec.utils.glob("**/package.json")

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

// FindPackageJSONFilesWithScript returns a list of package.json fileUtils that contain the script
func (exec *execute) FindPackageJSONFilesWithScript(packageJSONFiles []string, script string) ([]string, error) {
	var packagesWithScript []string

	for _, file := range packageJSONFiles {
		var packageJSON map[string]interface{}

		packageRaw, err := exec.utils.fileRead(file)
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

// InstallAllDependencies executes npm or yarn install for all package.json fileUtils defined in packageJSONFiles
func (exec *execute) InstallAllDependencies(packageJSONFiles []string) error {
	for _, packageJSON := range packageJSONFiles {
		err := exec.install(packageJSON)
		if err != nil {
			return err
		}
	}
	return nil
}

// install executes npm or yarn install for package.json
func (exec *execute) install(packageJSON string) error {
	execRunner := exec.utils.getExecRunner()

	oldWorkingDirectory, err := exec.utils.getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory before executing npm scripts: %w", err)
	}

	dir := path.Dir(packageJSON)
	err = exec.utils.chdir(dir)
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

	err = exec.utils.chdir(oldWorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to change back into original directory: %w", err)
	}
	return nil
}

// checkIfLockFilesExist checks if yarn/package lock fileUtils exist
func (exec *execute) checkIfLockFilesExist() (bool, bool, error) {
	packageLockExists, err := exec.utils.fileExists("package-lock.json")
	if err != nil {
		return false, false, err
	}

	yarnLockExists, err := exec.utils.fileExists("yarn.lock")
	if err != nil {
		return false, false, err
	}
	return packageLockExists, yarnLockExists, nil
}
