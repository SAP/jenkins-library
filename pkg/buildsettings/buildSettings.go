package buildsettings

import (
	"encoding/json"
	"reflect"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

type BuildSettingsInfo struct {
	Profiles                    []string `json:"profiles,omitempty"`
	Publish                     bool     `json:"publish,omitempty"`
	CreateBOM                   bool     `json:"createBOM,omitempty"`
	LogSuccessfulMavenTransfers bool     `json:"logSuccessfulMavenTransfers,omitempty"`
	GlobalSettingsFile          string   `json:"globalSettingsFile,omitempty"`
	DefaultNpmRegistry          string   `json:"defaultNpmRegistry,omitempty"`
}

type BuildSettings struct {
	MavenBuild  []BuildSettingsInfo `json:"mavenBuild,omitempty"`
	NpmBuild    []BuildSettingsInfo `json:"npmBuild,omitempty"`
	DockerBuild []BuildSettingsInfo `json:"dockerBuild,omitempty"`
	MtaBuild    []BuildSettingsInfo `json:"mtaBuild,omitempty"`
}

type BuildOptions struct {
	Profiles                    []string `json:"profiles,omitempty"`
	GlobalSettingsFile          string   `json:"globalSettingsFile,omitempty"`
	LogSuccessfulMavenTransfers bool     `json:"logSuccessfulMavenTransfers,omitempty"`
	CreateBOM                   bool     `json:"createBOM,omitempty"`
	Publish                     bool     `json:"publish,omitempty"`
	BuildSettingsInfo           string   `json:"buildSettingsInfo,omitempty"`
	DefaultNpmRegistry          string   `json:"defaultNpmRegistry,omitempty"`
}

func CreateBuildSettingsInfo(config *BuildOptions, buildTool string) (string, error) {
	currentBuildSettingsInfo := BuildSettingsInfo{
		CreateBOM:                   config.CreateBOM,
		GlobalSettingsFile:          config.GlobalSettingsFile,
		LogSuccessfulMavenTransfers: config.LogSuccessfulMavenTransfers,
		Profiles:                    config.Profiles,
		Publish:                     config.Publish,
		DefaultNpmRegistry:          config.DefaultNpmRegistry,
	}
	var jsonMap map[string][]interface{}
	var jsonResult []byte

	if len(config.BuildSettingsInfo) > 0 {

		err := json.Unmarshal([]byte(config.BuildSettingsInfo), &jsonMap)
		if err != nil {
			return "", errors.Wrapf(err, "failed to unmarshal existing build settings json '%v'", config.BuildSettingsInfo)
		}

		if mavenBuild, exist := jsonMap[buildTool]; exist {
			if reflect.TypeOf(mavenBuild).Kind() == reflect.Slice {
				jsonMap[buildTool] = append(mavenBuild, currentBuildSettingsInfo)
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
		var settings []BuildSettingsInfo
		settings = append(settings, currentBuildSettingsInfo)
		var err error
		jsonResult, err = json.Marshal(BuildSettings{
			MavenBuild: settings,
		})
		if err != nil {
			return "", errors.Wrapf(err, "Creating build settings failed with json marshalling")
		}
	}

	log.Entry().Infof("build settings infomration successfully created with '%v", string(jsonResult))

	return string(jsonResult), nil

}
