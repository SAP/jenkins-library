package cmd

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func npmExecuteScripts(config npmExecuteScriptsOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *npmExecuteScriptsCommonPipelineEnvironment) {
	npmExecutorOptions := npm.ExecutorOptions{DefaultNpmRegistry: config.DefaultNpmRegistry}
	npmExecutor := npm.NewExecutor(npmExecutorOptions)

	err := runNpmExecuteScripts(npmExecutor, &config, commonPipelineEnvironment)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNpmExecuteScripts(npmExecutor npm.Executor, config *npmExecuteScriptsOptions, commonPipelineEnvironment *npmExecuteScriptsCommonPipelineEnvironment) error {
	if config.Install {
		packageJSONFiles, err := npmExecutor.FindPackageJSONFilesWithExcludes(config.BuildDescriptorExcludeList)
		if err != nil {
			return err
		}

		err = npmExecutor.InstallAllDependencies(packageJSONFiles)
		if err != nil {
			return err
		}
	}

	if config.CreateBOM {
		packageJSONFiles, err := npmExecutor.FindPackageJSONFilesWithExcludes(config.BuildDescriptorExcludeList)
		if err != nil {
			return err
		}

		if err := npmExecutor.CreateBOM(packageJSONFiles); err != nil {
			return err
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

	log.Entry().Info("Creating files for build-setting.json")
	commonPipelineEnvironment.custom.createBom = config.CreateBOM
	commonPipelineEnvironment.custom.publish = config.Publish
	commonPipelineEnvironment.custom.defaultNpmRegistry = config.DefaultNpmRegistry

	var dataParametersJSON map[string]interface{}
	err = json.Unmarshal([]byte(GeneralConfig.ParametersJSON), &dataParametersJSON)
	if err != nil {
		return errors.Wrap(err, "Reading ParametersJSON is failed")
	}
	if value, ok := dataParametersJSON["dockerImage"]; ok {
		commonPipelineEnvironment.custom.dockerImage = value.(string)
	} else {
		log.Entry().Info("Value dockerImage doesn;t exist")
	}

	return nil
}
