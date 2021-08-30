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

func findPackageDescriptors(npmExecutor npm.Executor, config *npmExecuteScriptsOptions) ([]string, error) {
	if len(config.BuildDescriptorList) > 0 {
		return config.BuildDescriptorList, nil
	} else {
		return npmExecutor.FindPackageJSONFilesWithExcludes(config.BuildDescriptorExcludeList)
	}
}

func runNpmExecuteScripts(npmExecutor npm.Executor, config *npmExecuteScriptsOptions) error {
	packageJSONFiles, err := findPackageDescriptors(npmExecutor, config)

        if err != nil {
		return err
	}

	if config.Install {
		err = npmExecutor.InstallAllDependencies(packageJSONFiles)

		if err != nil {
			return err
		}
	}

	if config.CreateBOM {
		if err = npmExecutor.CreateBOM(packageJSONFiles); err != nil {
			return err
		}
	}

	err = npmExecutor.RunScriptsInAllPackages(config.RunScripts, nil, config.ScriptOptions, config.VirtualFrameBuffer, config.BuildDescriptorExcludeList, config.BuildDescriptorList)
	if err != nil {
		return err
	}

	if config.Publish {
		err = npmExecutor.PublishAllPackages(packageJSONFiles, config.RepositoryURL, config.RepositoryUsername, config.RepositoryPassword)
		if err != nil {
			return err
		}
	}

	return nil
}
