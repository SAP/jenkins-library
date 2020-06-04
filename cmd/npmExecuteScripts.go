package cmd

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"path"
)

func npmExecuteScripts(config npmExecuteScriptsOptions, telemetryData *telemetry.CustomData) {
	utils := npm.NpmUtilsBundle{}

	err := runNpmExecuteScripts(&utils, &config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}
func runNpmExecuteScripts(utils npm.NpmUtils, options *npmExecuteScriptsOptions) error {
	execRunner := utils.GetExecRunner()
	packageJSONFiles, err := npm.FindPackageJSONFiles(utils)
	if err != nil {
		return err
	}

	oldWorkingDirectory, err := utils.Getwd()
	if err != nil {
		return err
	}

	for _, file := range packageJSONFiles {
		dir := path.Dir(file)
		err = utils.Chdir(dir)
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

		packageLockExists, yarnLockExists, err := npm.CheckIfLockFilesExist(utils)
		if err != nil {
			return err
		}
		if options.Install {
			err = npm.InstallDependencies(dir, packageLockExists, yarnLockExists, execRunner)
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
		err = utils.Chdir(oldWorkingDirectory)
		if err != nil {
			return err
		}
	}

	return err
}

