package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/bmatcuk/doublestar"
	"os"
	"path"
	"strings"
)

type nodeJsBuildUtilsInterface interface {
	fileExists(path string) (bool, error)
	glob(pattern string) (matches []string, err error)
	getwd() (dir string, err error)
	dir(path string) string
	chdir(dir string) error
	getExecRunner() execRunner
}

type nodeJsBuildUtilsBundle struct {
	fileUtils  FileUtils.Files
	execRunner *command.Command
}

func (u *nodeJsBuildUtilsBundle) fileExists(path string) (bool, error) {
	return u.fileUtils.FileExists(path)
}

func (u *nodeJsBuildUtilsBundle) glob(pattern string) (matches []string, err error) {
	return doublestar.Glob(pattern)
}

func (u *nodeJsBuildUtilsBundle) getwd() (dir string, err error) {
	return os.Getwd()
}

func (u *nodeJsBuildUtilsBundle) dir(fileName string) string {
	return path.Dir(fileName)
}

func (u *nodeJsBuildUtilsBundle) chdir(dir string) error {
	return os.Chdir(dir)
}

func (u *nodeJsBuildUtilsBundle) getExecRunner() execRunner {
	if u.execRunner == nil {
		u.execRunner = &command.Command{}
		u.execRunner.Stdout(log.Entry().Writer())
		u.execRunner.Stderr(log.Entry().Writer())
	}
	return u.execRunner
}

func nodeJsBuild(config nodeJsBuildOptions, telemetryData *telemetry.CustomData) {
	utils := nodeJsBuildUtilsBundle{}

	err := runNodeJsBuild(&utils, &config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}
func runNodeJsBuild(utils nodeJsBuildUtilsInterface, options *nodeJsBuildOptions) error {
	execRunner := utils.getExecRunner()
	log.Entry().Infof("NPM Registry configuration: defaultNpmRegistry %s, sapNpmRegistry %s",
		options.DefaultNpmRegistry, options.SapNpmRegistry)
	environment := []string{"npm_config_@sap:registry=" + options.SapNpmRegistry}
	if options.DefaultNpmRegistry != "" {
		environment = append(environment, "npm_config_registry="+options.DefaultNpmRegistry)
	}
	execRunner.SetEnv(environment)

	packageJsonFiles, err := findPackageJSONFiles(utils)
	if err != nil {
		return err
	}

	oldWorkingDirectory, err := utils.getwd()

	for _, file := range packageJsonFiles {
		dir := utils.dir(file)
		_ = utils.chdir(dir)
		packageLockExists, err := utils.fileExists("package-lock.json")

		if err != nil {
			return err
		}
		yarnLockExists, err := utils.fileExists("yarn.lock")
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
		_ = utils.chdir(oldWorkingDirectory)
	}

	return err
}

func findPackageJSONFiles(utils nodeJsBuildUtilsInterface) ([]string, error) {
	unfilteredListOfPackageJSONFiles, err := utils.glob("**/package.json")
	if err != nil {
		return nil, err
	}

	var packageJsonFiles []string

	for _, file := range unfilteredListOfPackageJSONFiles {
		if strings.Contains(file, "node_modules") {
			continue
		}
		packageJsonFiles = append(packageJsonFiles, file)
		log.Entry().Info("Discovered package.json file " + file)
	}
	return packageJsonFiles, nil
}

func installDependencies(dir string, packageLockExists bool, yarnLockExists bool, execRunner execRunner) (err error) {
	log.Entry().WithField("WorkingDirectory", dir).Info("Running install")
	if packageLockExists {
		log.Entry().Info("npm ci")
		err = execRunner.RunExecutable("npm", "ci", "-ddd")
		if err != nil {
			return err
		}
	} else if yarnLockExists {
		log.Entry().Info("yarn install --frozen-lockfile")
		err = execRunner.RunExecutable("yarn", "install", "--frozen-lockfile")
		if err != nil {
			return err
		}
	} else {
		log.Entry().Warn("No package lock file found. " +
			"It is recommended to create a `package-lock.json` file by running `npm install` locally." +
			" Add this file to your version control. " +
			"By doing so, the builds of your application become more reliable.")
		err = execRunner.RunExecutable("npm", "install", "-ddd")
		if err != nil {
			return err
		}
	}
	return nil
}
