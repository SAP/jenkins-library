package cmd

import (
	"encoding/json"
	"os"

	"github.com/SAP/jenkins-library/pkg/build"
	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

func npmExecuteScripts(config npmExecuteScriptsOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *npmExecuteScriptsCommonPipelineEnvironment) {
	npmExecutorOptions := npm.ExecutorOptions{
		DefaultNpmRegistry: config.DefaultNpmRegistry,
		PnpmVersion:        config.PnpmVersion,
	}
	npmExecutor := npm.NewExecutor(npmExecutorOptions)

	err := runNpmExecuteScripts(npmExecutor, &config, commonPipelineEnvironment)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runNpmExecuteScripts(npmExecutor npm.Executor, config *npmExecuteScriptsOptions, commonPipelineEnvironment *npmExecuteScriptsCommonPipelineEnvironment) error {
	// setting env. variable to omit installation of dev. dependencies
	if config.Production {
		os.Setenv("NODE_ENV", "production")
	}

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

	log.Entry().Debugf("creating build settings information...")
	stepName := "npmExecuteScripts"
	dockerImage, err := GetDockerImageValue(stepName)
	if err != nil {
		return err
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

	buildCoordinates := []versioning.Coordinates{}

	if config.Publish {
		if len(config.BuildDescriptorList) > 0 {
			err = npmExecutor.PublishAllPackages(config.BuildDescriptorList, config.RepositoryURL, config.RepositoryUsername, config.RepositoryPassword, config.PublishTag, config.PackBeforePublish, &buildCoordinates)
			if err != nil {
				return err
			}
		} else {
			packageJSONFiles, err := npmExecutor.FindPackageJSONFilesWithExcludes(config.BuildDescriptorExcludeList)
			if err != nil {
				return err
			}

			err = npmExecutor.PublishAllPackages(packageJSONFiles, config.RepositoryURL, config.RepositoryUsername, config.RepositoryPassword, config.PublishTag, config.PackBeforePublish, &buildCoordinates)
			if err != nil {
				return err
			}
		}
	}

	if config.CreateBuildArtifactsMetadata {
		if len(buildCoordinates) == 0 {
			log.Entry().Warnf("unable to identify artifact coordinates for the npm packages published")
			return nil
		}

		var buildArtifacts build.BuildArtifacts

		buildArtifacts.Coordinates = buildCoordinates
		jsonResult, _ := json.Marshal(buildArtifacts)
		commonPipelineEnvironment.custom.npmBuildArtifacts = string(jsonResult)
	}

	return nil
}
