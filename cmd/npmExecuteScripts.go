package cmd

import (
	"encoding/json"
	"reflect"

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

	log.Entry().Infof("creating build settings information...")
	builSettingsErr := createNpmBuildSettingsInfo(config, commonPipelineEnvironment)

	if builSettingsErr != nil {
		log.Entry().Warnf("failed to create build settings info : ''%v", builSettingsErr)
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

func createNpmBuildSettingsInfo(config *npmExecuteScriptsOptions, commonPipelineEnvironment *npmExecuteScriptsCommonPipelineEnvironment) error {
	currentBuildSettingsInfo := buildsettings.BuildSettingsInfo{
		Publish:            config.Publish,
		CreateBOM:          config.CreateBOM,
		DefaultNpmRegistry: config.DefaultNpmRegistry,
	}
	var jsonMap map[string][]interface{}

	if len(config.BuildSettingsInfo) > 0 {

		err := json.Unmarshal([]byte(config.BuildSettingsInfo), &jsonMap)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal existing build settings json '%v'", config.BuildSettingsInfo)
		}

		if npmBuild, exist := jsonMap["npmBuild"]; exist {
			if reflect.TypeOf(npmBuild).Kind() == reflect.Slice {
				jsonMap["npmBuild"] = append(npmBuild, currentBuildSettingsInfo)
			}
		} else {
			var settings []interface{}
			settings = append(settings, currentBuildSettingsInfo)
			jsonMap["npmBuild"] = settings
		}

		newJsonMap, err := json.Marshal(&jsonMap)
		if err != nil {
			return errors.Wrapf(err, "Creating build settings failed with json marshalling")
		}

		commonPipelineEnvironment.custom.buildSettingsInfo = string(newJsonMap)

	} else {
		var settings []buildsettings.BuildSettingsInfo
		settings = append(settings, currentBuildSettingsInfo)
		jsonResult, err := json.Marshal(buildsettings.BuildSettings{
			NpmBuild: settings,
		})
		if err != nil {
			return errors.Wrapf(err, "Creating build settings failed with json marshalling")
		} else {
			commonPipelineEnvironment.custom.buildSettingsInfo = string(jsonResult)
		}
	}

	log.Entry().Infof("build settings infomration successfully created with '%v", commonPipelineEnvironment.custom.buildSettingsInfo)

	return nil

}
