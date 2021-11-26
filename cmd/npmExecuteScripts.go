package cmd

import (
	"github.com/SAP/jenkins-library/pkg/buildsettings"
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

	log.Entry().Debugf("creating build settings information...")
	stepName := "npmExecuteScripts"
	configOptions.contextConfig = true
	configOptions.stepName = stepName
	stepConfig, err := getConfig()
	if err != nil {
		return err
	}

	dockerImage, ok := stepConfig.Config["dockerImage"].(string)
	if !ok {
		return errors.Errorf("error: config value of %v to compare with is not a string", stepConfig.Config["dockerImage"])
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
