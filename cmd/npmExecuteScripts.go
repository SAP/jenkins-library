package cmd

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/buildsettings"
	conf "github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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

	log.Entry().Debugf("creating build settings information...")
	stepName := "npmExecuteScripts"
	var dockerImage string
	var dataParametersJSON map[string]interface{}
	var errUnmarshal = json.Unmarshal([]byte(GeneralConfig.ParametersJSON), &dataParametersJSON)
	if errUnmarshal != nil {
		log.Entry().Infof("Reading ParametersJSON is failed")
	}
	if value, ok := dataParametersJSON["dockerImage"]; ok {
		dockerImage = value.(string)
	} else {
		metadata, err := conf.ResolveMetadata(GeneralConfig.GitHubAccessTokens, GetAllStepMetadata, GeneralConfig.StepMetadata, stepName)
		if err != nil {
			log.Entry().Warnf("failed to resolve metadata: %v", err)
		}
		containers := metadata.Spec.Containers
		if len(containers) > 0 {
			dockerImage = containers[0].Image
		}
	}
	npmConfig := buildsettings.BuildOptions{
		Publish:            config.Publish,
		CreateBOM:          config.CreateBOM,
		DefaultNpmRegistry: config.DefaultNpmRegistry,
		BuildSettingsInfo:  config.BuildSettingsInfo,
		DockerImage:        dockerImage,
	}
	buildSettingsInfo, err := buildsettings.CreateBuildSettingsInfo(&npmConfig, stepName)
	if err != nil {
		log.Entry().Warnf("failed to create build settings info: %v", err)
	}
	commonPipelineEnvironment.custom.buildSettingsInfo = buildSettingsInfo

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
