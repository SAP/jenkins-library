package cmd

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func npmExecuteScripts(config npmExecuteScriptsOptions, telemetryData *telemetry.CustomData) {
	utils := npm.UtilsBundle{}

	err := runNpmExecuteScripts(&utils, &config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNpmExecuteScripts(utils npm.Utils, config *npmExecuteScriptsOptions) error {
	options := npm.ExecuteOptions{
		Install:            config.Install,
		RunScripts:         config.RunScripts,
		DefaultNpmRegistry: config.DefaultNpmRegistry,
		SapNpmRegistry:     config.SapNpmRegistry,
	}

	packageJSONFiles := npm.FindPackageJSONFiles(utils)

	if options.Install {
		err := npm.InstallAllDependencies(packageJSONFiles, utils, &options)
		if err != nil {
			return err
		}
	}

	err := npm.ExecuteAllScripts(utils, options)
	if err != nil {
		return err
	}
	return err
}
