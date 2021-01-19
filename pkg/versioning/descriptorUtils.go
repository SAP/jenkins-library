package versioning

import (
	"github.com/Masterminds/sprig"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

const (
	// SchemeMajorVersion is the versioning scheme based on the major version only
	SchemeMajorVersion = `{{(split "." (split "-" .Version)._0)._0}}`
	// SchemeMajorMinorVersion is the versioning scheme based on the major version only
	SchemeMajorMinorVersion = `{{(split "." (split "-" .Version)._0)._0}}.{{(split "." (split "-" .Version)._0)._1}}`
	// SchemeSemanticVersion is the versioning scheme based on the major.minor.micro version
	SchemeSemanticVersion = `{{(split "." (split "-" .Version)._0)._0}}.{{(split "." (split "-" .Version)._0)._1}}.{{(split "." (split "-" .Version)._0)._2}}`
	// SchemeFullVersion is the versioning scheme based on the full version
	SchemeFullVersion = "{{.Version}}"
)

// DetermineProjectCoordinates resolve the coordinates of the project for use in 3rd party scan tools
func DetermineProjectCoordinates(nameTemplate, versionScheme string, gav Coordinates) (string, string) {
	projectName, err := piperutils.ExecuteTemplateFunctions(nameTemplate, sprig.HermeticTxtFuncMap(), gav)
	if err != nil {
		log.Entry().Warnf("Unable to resolve project name: %v", err)
	}

	var versionTemplate string
	if versionScheme == "full" {
		versionTemplate = SchemeFullVersion
	}
	if versionScheme == "major" {
		versionTemplate = SchemeMajorVersion
	}
	if versionScheme == "major-minor" {
		versionTemplate = SchemeMajorMinorVersion
	}
	if versionScheme == "semantic" {
		versionTemplate = SchemeSemanticVersion
	}

	projectVersion, err := piperutils.ExecuteTemplateFunctions(versionTemplate, sprig.HermeticTxtFuncMap(), gav)
	if err != nil {
		log.Entry().Warnf("Unable to resolve project version: %v", err)
	}
	return projectName, projectVersion
}
