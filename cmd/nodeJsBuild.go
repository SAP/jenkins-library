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

func nodeJsBuild(config nodeJsBuildOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	err := runNodeJsBuild(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNodeJsBuild(config *nodeJsBuildOptions, telemetryData *telemetry.CustomData, command execRunner) error {

	environment := []string{"npm_config_@sap:registry=" + config.SapNpmRegistry}
	if config.DefaultNpmRegistry != "" {
		environment = append(environment, "npm_config_registry=" + config.DefaultNpmRegistry)
	}
	command.SetEnv(environment)

	unfilteredListOfPackageJsonFiles, err := doublestar.Glob("**/package.json")
	if err != nil {
		return err
	}

	var packageJsonFiles []string

	for _, file := range unfilteredListOfPackageJsonFiles {
		if strings.Contains(file, "node_modules") {
			continue
		}
		packageJsonFiles = append(packageJsonFiles, file)
		log.Entry().Info("Discovered package.json file " + file)
	}

	oldWorkingDirectory, err := os.Getwd()

	for _, file := range packageJsonFiles {
		base := path.Base(file)
		_ = os.Chdir(base)
		packageLockExists, err := FileUtils.FileExists("package-lock.json")

		if err != nil {
			return err
		}
		yarnLockExists, err := FileUtils.FileExists("yarn.lock")
		if err != nil {
			return err
		}
		if config.Install {
			log.Entry().WithField("WorkingDirectory", base).Info("Running install")
			if packageLockExists {
				err = command.RunExecutable("npm", "ci")
				if err != nil {
					return err
				}
			} else if yarnLockExists {
				err = command.RunExecutable("yarn", "install", "--frozen-lockfile")
				if err != nil {
					return err
				}
			} else {
				log.Entry().Warn("No package lock file found. " +
					"It is recommended to create a `package-lock.json` file by running `npm install` locally." +
					" Add this file to your version control. " +
					"By doing so, the builds of your application become more reliable.")
				err = command.RunExecutable("npm", "install")
				if err != nil {
					return err
				}
			}
		}

		for _, v := range config.RunScripts {
			log.Entry().WithField("WorkingDirectory", base).Info("run-script " + v)
			err = command.RunExecutable("npm", "run-script", v, "--if-present")
			if err != nil {
				return err
			}
		}
		_ = os.Chdir(oldWorkingDirectory)
	}

	return err
}
