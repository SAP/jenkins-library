package cmd

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func npmExecuteScripts(config npmExecuteScriptsOptions, telemetryData *telemetry.CustomData) {
	npmExecutorOptions := npm.ExecutorOptions{Install: config.Install, RunScripts: config.RunScripts, DefaultNpmRegistry: config.DefaultNpmRegistry, SapNpmRegistry: config.SapNpmRegistry, VirtualFrameBuffer: config.VirtualFrameBuffer}
	npmExecutor, err := npm.NewExecutor(npmExecutorOptions)

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

	return npmExecutor.ExecuteAllScripts()
}
