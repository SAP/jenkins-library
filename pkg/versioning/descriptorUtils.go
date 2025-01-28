package versioning

import (
	"github.com/Masterminds/sprig/v3"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// DetermineProjectCoordinatesWithCustomVersion resolves the coordinates of the project for use in 3rd party scan tools
// It considers a custom version if provided instead of using the GAV version adapted according to the versionScheme
func DetermineProjectCoordinatesWithCustomVersion(nameTemplate, versionScheme, customVersion string, gav Coordinates) (string, string) {
	name, version := DetermineProjectCoordinates(nameTemplate, versionScheme, gav)
	if len(customVersion) > 0 {
		log.Entry().Infof("Using custom version: %v", customVersion)
		return name, customVersion
	}
	return name, version
}

// DetermineProjectCoordinates resolves the coordinates of the project for use in 3rd party scan tools
func DetermineProjectCoordinates(nameTemplate, versionScheme string, gav Coordinates) (string, string) {
	projectName, err := piperutils.ExecuteTemplateFunctions(nameTemplate, sprig.HermeticTxtFuncMap(), gav)
	if err != nil {
		log.Entry().Warnf("Unable to resolve project name: %v", err)
	}

	projectVersion := ApplyVersioningModel(versionScheme, gav.Version)
	return projectName, projectVersion
}
