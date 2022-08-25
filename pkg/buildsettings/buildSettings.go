package buildsettings

import (
	"encoding/json"
	"reflect"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

type BuildSettings struct {
	GolangBuild       []BuildOptions `json:"golangBuild,omitempty"`
	GradleBuild       []BuildOptions `json:"gradleBuild,omitempty"`
	HelmExecute       []BuildOptions `json:"helmExecute,omitempty"`
	KanikoExecute     []BuildOptions `json:"kanikoExecute,omitempty"`
	MavenBuild        []BuildOptions `json:"mavenBuild,omitempty"`
	MtaBuild          []BuildOptions `json:"mtaBuild,omitempty"`
	PythonBuild       []BuildOptions `json:"pythonBuild,omitempty"`
	NpmExecuteScripts []BuildOptions `json:"npmExecuteScripts,omitempty"`
	CnbBuild          []BuildOptions `json:"cnbBuild,omitempty"`
}

type BuildOptions struct {
	Profiles                    []string `json:"profiles,omitempty"`
	Publish                     bool     `json:"publish,omitempty"`
	CreateBOM                   bool     `json:"createBOM,omitempty"`
	LogSuccessfulMavenTransfers bool     `json:"logSuccessfulMavenTransfers,omitempty"`
	GlobalSettingsFile          string   `json:"globalSettingsFile,omitempty"`
	DefaultNpmRegistry          string   `json:"defaultNpmRegistry,omitempty"`
	BuildSettingsInfo           string   `json:"buildSettingsInfo,omitempty"`
	DockerImage                 string   `json:"dockerImage,omitempty"`
}

func CreateBuildSettingsInfo(config *BuildOptions, buildTool string) (string, error) {
	currentBuildSettingsInfo := BuildOptions{
		CreateBOM:                   config.CreateBOM,
		GlobalSettingsFile:          config.GlobalSettingsFile,
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
		Profiles:                    config.Profiles,
		Publish:                     config.Publish,
		DefaultNpmRegistry:          config.DefaultNpmRegistry,
		DockerImage:                 config.DockerImage,
	}
	var jsonMap map[string][]interface{}
	var jsonResult []byte

	if config.BuildSettingsInfo != "" {

		err := json.Unmarshal([]byte(config.BuildSettingsInfo), &jsonMap)
		if err != nil {
			return "", errors.Wrapf(err, "failed to unmarshal existing build settings json '%v'", config.BuildSettingsInfo)
		}

		if build, exist := jsonMap[buildTool]; exist {
			if reflect.TypeOf(build).Kind() == reflect.Slice {
				jsonMap[buildTool] = append(build, currentBuildSettingsInfo)
			}
		} else {
			var settings []interface{}
			settings = append(settings, currentBuildSettingsInfo)
			jsonMap[buildTool] = settings
		}

		jsonResult, err = json.Marshal(&jsonMap)
		if err != nil {
			return "", errors.Wrapf(err, "Creating build settings failed with json marshalling")
		}
	} else {
		var settings []BuildOptions
		settings = append(settings, currentBuildSettingsInfo)
		var err error
		switch buildTool {
		case "golangBuild":
			jsonResult, err = json.Marshal(BuildSettings{
				GolangBuild: settings,
			})
		case "gradleBuild":
			jsonResult, err = json.Marshal(BuildSettings{
				GradleBuild: settings,
			})
		case "helmExecute":
			jsonResult, err = json.Marshal(BuildSettings{
				HelmExecute: settings,
			})
		case "kanikoExecute":
			jsonResult, err = json.Marshal(BuildSettings{
				KanikoExecute: settings,
			})
		case "mavenBuild":
			jsonResult, err = json.Marshal(BuildSettings{
				MavenBuild: settings,
			})
		case "mtaBuild":
			jsonResult, err = json.Marshal(BuildSettings{
				MtaBuild: settings,
			})
		case "pythonBuild":
			jsonResult, err = json.Marshal(BuildSettings{
				PythonBuild: settings,
			})
		case "npmExecuteScripts":
			jsonResult, err = json.Marshal(BuildSettings{
				NpmExecuteScripts: settings,
			})
		case "cnbBuild":
			jsonResult, err = json.Marshal(BuildSettings{
				CnbBuild: settings,
			})
		default:
			log.Entry().Warningf("buildTool '%s' not supported for creation of build settings", buildTool)
			return "", nil
		}
		if err != nil {
			return "", errors.Wrapf(err, "Creating build settings failed with json marshalling")
		}
	}

	log.Entry().Infof("build settings information successfully created with '%v", string(jsonResult))

	return string(jsonResult), nil

}
