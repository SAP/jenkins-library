package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/bmatcuk/doublestar"
	"os"
	"path"
	"strings"
)

type npmExecuteScriptsUtilsInterface interface {
	fileExists(path string) (bool, error)
	glob(pattern string) (matches []string, err error)
	getwd() (dir string, err error)
	chdir(dir string) error
	getExecRunner() execRunner
}

type npmExecuteScriptsUtilsBundle struct {
	projectStructure FileUtils.ProjectStructure
	fileUtils        FileUtils.Files
	execRunner       *command.Command
}

func (u *npmExecuteScriptsUtilsBundle) fileExists(path string) (bool, error) {
	return u.fileUtils.FileExists(path)
}

func (u *npmExecuteScriptsUtilsBundle) glob(pattern string) (matches []string, err error) {
	return doublestar.Glob(pattern)
}

func (u *npmExecuteScriptsUtilsBundle) getwd() (dir string, err error) {
	return os.Getwd()
}

func (u *npmExecuteScriptsUtilsBundle) chdir(dir string) error {
	return os.Chdir(dir)
}

func (u *npmExecuteScriptsUtilsBundle) getExecRunner() execRunner {
	if u.execRunner == nil {
		u.execRunner = &command.Command{}
		u.execRunner.Stdout(log.Writer())
		u.execRunner.Stderr(log.Writer())
	}
	return u.execRunner
}

func npmExecuteScripts(config npmExecuteScriptsOptions, telemetryData *telemetry.CustomData) {
	utils := npmExecuteScriptsUtilsBundle{}

	err := runNpmExecuteScripts(&utils, &config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}
func runNpmExecuteScripts(utils npmExecuteScriptsUtilsInterface, options *npmExecuteScriptsOptions) error {
	execRunner := utils.getExecRunner()
	packageJSONFiles, err := findPackageJSONFiles(utils)
	if err != nil {
		return err
	}

	oldWorkingDirectory, err := utils.getwd()
	if err != nil {
		return err
	}

	for _, file := range packageJSONFiles {
		dir := path.Dir(file)
		err = utils.chdir(dir)
		if err != nil {
			return err
		}

		// set in each directory to respect existing config in rc files
		err = npm.SetNpmRegistries(
			&npm.RegistryOptions{
				DefaultNpmRegistry: options.DefaultNpmRegistry,
				SapNpmRegistry:     options.SapNpmRegistry,
			}, execRunner)

		if err != nil {
			return err
		}

		packageLockExists, yarnLockExists, err := checkIfLockFilesExist(utils)
		if err != nil {
			return err
		}
		if options.Install {
			err = installDependencies(dir, packageLockExists, yarnLockExists, execRunner)
			if err != nil {
				return err
			}
		}

		for _, v := range options.RunScripts {
			log.Entry().WithField("WorkingDirectory", dir).Info("run-script " + v)
			err = execRunner.RunExecutable("npm", "run-script", v, "--if-present")
			if err != nil {
				return err
			}
		}
		err = utils.chdir(oldWorkingDirectory)
		if err != nil {
			return err
		}
	}

	return err
}

func findPackageJSONFiles(utils npmExecuteScriptsUtilsInterface) ([]string, error) {
	unfilteredListOfPackageJSONFiles, err := utils.glob("**/package.json")
	if err != nil {
		return nil, err
	}

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
	return packageJSONFiles, nil
}
func checkIfLockFilesExist(utils npmExecuteScriptsUtilsInterface) (bool, bool, error) {
	packageLockExists, err := utils.fileExists("package-lock.json")

	if err != nil {
		return false, false, err
	}
	yarnLockExists, err := utils.fileExists("yarn.lock")
	if err != nil {
		return false, false, err
	}
	return packageLockExists, yarnLockExists, nil
}

func installDependencies(dir string, packageLockExists bool, yarnLockExists bool, execRunner execRunner) (err error) {
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
	return nil
}
