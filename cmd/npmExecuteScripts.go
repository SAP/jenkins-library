package cmd

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func npmExecuteScripts(config npmExecuteScriptsOptions, telemetryData *telemetry.CustomData) {
	npmExecutorOptions := npm.ExecutorOptions{DefaultNpmRegistry: config.DefaultNpmRegistry}
	npmExecutor := npm.NewExecutor(npmExecutorOptions)

	err := runNpmExecuteScripts(npmExecutor, &config)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNpmExecuteScripts(npmExecutor npm.Executor, config *npmExecuteScriptsOptions) error {

	if config.Install {
		if len(config.BuildDescriptorList) > 0 {
			if err := npmExecutor.InstallAllDependencies(config.BuildDescriptorList); err != nil {
				return err
			}
		} else {
			packageJSONFiles, err := npmExecutor.FindPackageJSONFilesWithExcludes(config.BuildDescriptorExcludeList)
			if err != nil {
				return err
			}

			if err := npmExecutor.InstallAllDependencies(packageJSONFiles); err != nil {
				return err
			}
		}
	}

	if config.CreateBOM {
		if len(config.BuildDescriptorList) > 0 {
			if err := npmExecutor.CreateBOM(config.BuildDescriptorList); err != nil {
				return err
			}
		} else {
			packageJSONFiles, err := npmExecutor.FindPackageJSONFilesWithExcludes(config.BuildDescriptorExcludeList)
			if err != nil {
				return err
			}

			if err := npmExecutor.CreateBOM(packageJSONFiles); err != nil {
				return err
			}
		}
	}

	err := npmExecutor.RunScriptsInAllPackages(config.RunScripts, nil, config.ScriptOptions, config.VirtualFrameBuffer, config.BuildDescriptorExcludeList, config.BuildDescriptorList)
	if err != nil {
		return err
	}

	if config.Publish {
		packageJSONFiles, err := npmExecutor.FindPackageJSONFilesWithExcludes(config.BuildDescriptorExcludeList)
		if err != nil {
			return err
		}

		err = npmExecutor.PublishAllPackages(packageJSONFiles, config.RepositoryURL, config.RepositoryUsername, config.RepositoryPassword)
		if err != nil {
			return err
		}
	}

	return nil
}
