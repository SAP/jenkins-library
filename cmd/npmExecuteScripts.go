package cmd

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func npmExecuteScripts(config npmExecuteScriptsOptions, telemetryData *telemetry.CustomData) {
	npmExecutor, err := npm.NewExecutor(config.Install, config.RunScripts, []string{}, config.DefaultNpmRegistry, config.SapNpmRegistry)

	err = runNpmExecuteScripts(npmExecutor, &config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNpmExecuteScripts(npmExecutor npm.Executor, config *npmExecuteScriptsOptions) error {
	packageJSONFiles := npmExecutor.FindPackageJSONFiles()

	if config.Install {
		err := npmExecutor.InstallAllDependencies(packageJSONFiles)
		if err != nil {
			return err
		}
	}

	err := npmExecutor.ExecuteAllScripts()
	if err != nil {
		return err
	}
	return err
}
