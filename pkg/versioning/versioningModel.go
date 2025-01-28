package versioning

import (
	"github.com/Masterminds/sprig/v3"
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

const (
	VersioningModelFull       string = "full"
	VersioningModelSemantic   string = "semantic"
	VersioningModelMajorMinor string = "major-minor"
	VersioningModelMajor      string = "major"
)

func ApplyVersioningModel(model, projectVersion string) string {
	var versioningScheme string

	switch model {
	case VersioningModelFull:
		versioningScheme = SchemeFullVersion
	case VersioningModelSemantic:
		versioningScheme = SchemeSemanticVersion
	case VersioningModelMajorMinor:
		versioningScheme = SchemeMajorMinorVersion
	case VersioningModelMajor:
		versioningScheme = SchemeMajorVersion
	default:
		log.Entry().Warnf("versioning model not supported: %s", model)
	}

	version, err := piperutils.ExecuteTemplateFunctions(versioningScheme, sprig.HermeticTxtFuncMap(), Coordinates{Version: projectVersion})
	if err != nil {
		log.Entry().Warnf("unable to resolve project version: %v", err)
	}
	return version
}
