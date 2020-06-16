package npm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type execute struct {
	utils utils
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

// ExecutorOptions holds parameters to pass to NewExecutor()
type ExecutorOptions struct {
	Install            bool
	RunScripts         []string
	RunOptions         []string
	DefaultNpmRegistry string
	SapNpmRegistry     string
	ExecRunner         execRunner
}

// NewExecutor instantiates execute struct and sets executeOptions
func NewExecutor(executorOptions ExecutorOptions) (*execute, error) {
	utils := utilsBundle{Files: &piperutils.Files{}, execRunner: executorOptions.ExecRunner}
	executeOptions := executeOptions{
		install:            executorOptions.Install,
		runScripts:         executorOptions.RunScripts,
		runOptions:         executorOptions.RunOptions,
		defaultNpmRegistry: executorOptions.DefaultNpmRegistry,
		sapNpmRegistry:     executorOptions.SapNpmRegistry,
	}
	exec := &execute{
		utils:   &utils,
		options: executeOptions,
	}
	return exec, nil
}

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
	Stderr(out io.Writer)
	RunExecutable(executable string, params ...string) error
}

type utils interface {
	Chdir(path string) error
	FileExists(filename string) (bool, error)
	FileRead(path string) ([]byte, error)
	Getwd() (string, error)
	Glob(pattern string) (matches []string, err error)

	getExecRunner() execRunner
}

type utilsBundle struct {
	*piperutils.Files
	execRunner execRunner
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

// ExecuteAllScripts runs all scripts defined in ExecuteOptions.RunScripts
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
	oldWorkingDirectory, err := exec.utils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory before executing npm scripts: %w", err)
	}

	dir := filepath.Dir(packageJSON)
	err = exec.utils.Chdir(dir)
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

	err = exec.utils.Chdir(oldWorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to change back into original directory: %w", err)
	}
	return nil
}

// FindPackageJSONFiles returns a list of all package.json fileUtils of the project excluding node_modules and gen/ directories
func (exec *execute) FindPackageJSONFiles() []string {
	unfilteredListOfPackageJSONFiles, _ := exec.utils.Glob("**/package.json")

	var packageJSONFiles []string

	for _, file := range unfilteredListOfPackageJSONFiles {
		if strings.Contains(file, "node_modules") {
			continue
		}

		if strings.HasPrefix(file, "gen" + string(os.PathSeparator)) || strings.Contains(file, string(os.PathSeparator) + "gen" + string(os.PathSeparator)) {
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

		packageRaw, err := exec.utils.FileRead(file)
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
func (exec *execute) InstallAllDependencies(packageJSONFiles []string) error {
	for _, packageJSON := range packageJSONFiles {
		err := exec.install(packageJSON)
		if err != nil {
			return err
		}
	}
	return nil
}

// install executes npm or yarn Install for package.json
func (exec *execute) install(packageJSON string) error {
	execRunner := exec.utils.getExecRunner()

	oldWorkingDirectory, err := exec.utils.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory before executing npm scripts: %w", err)
	}

	dir := filepath.Dir(packageJSON)
	err = exec.utils.Chdir(dir)
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
		err = execRunner.RunExecutable("yarn", "Install", "--frozen-lockfile")
		if err != nil {
			return err
		}
	} else {
		log.Entry().Warn("No package lock file found. " +
			"It is recommended to create a `package-lock.json` file by running `npm Install` locally." +
			" Add this file to your version control. " +
			"By doing so, the builds of your application become more reliable.")
		err = execRunner.RunExecutable("npm", "Install")
		if err != nil {
			return err
		}
	}

	err = exec.utils.Chdir(oldWorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to change back into original directory: %w", err)
	}
	return nil
}

// checkIfLockFilesExist checks if yarn/package lock fileUtils exist
func (exec *execute) checkIfLockFilesExist() (bool, bool, error) {
	packageLockExists, err := exec.utils.FileExists("package-lock.json")
	if err != nil {
		return false, false, err
	}

	yarnLockExists, err := exec.utils.FileExists("yarn.lock")
	if err != nil {
		return false, false, err
	}
	return packageLockExists, yarnLockExists, nil
}
