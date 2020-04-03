package versioning

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/maven"
)

// Artifact ...
type Artifact interface {
	VersioningScheme() string
	GetVersion() (string, error)
	SetVersion(string) error
}

// Options ...
type Options struct {
	ProjectSettingsFile string
	GlobalSettingsFile  string
	M2Path              string
}

type mvnRunner struct{}

func (m *mvnRunner) Execute(options *maven.ExecuteOptions, execRunner mavenExecRunner) (string, error) {
	return maven.Execute(options, execRunner)
}
func (m *mvnRunner) Evaluate(pomFile, expression string, execRunner mavenExecRunner) (string, error) {
	return maven.Evaluate(pomFile, expression, execRunner)
}

// GetArtifact ...
func GetArtifact(buildTool, buildDescriptorFilePath string, opts *Options, execRunner mavenExecRunner) (Artifact, error) {
	var artifact Artifact
	switch buildTool {
	case "maven":
		artifact = &Maven{
			Runner:              &mvnRunner{},
			ExecRunner:          execRunner,
			PomPath:             buildDescriptorFilePath,
			ProjectSettingsFile: opts.ProjectSettingsFile,
			GlobalSettingsFile:  opts.GlobalSettingsFile,
			M2Path:              opts.M2Path,
		}
	case "npm":
		artifact = &Npm{
			PackageJSONPath: buildDescriptorFilePath,
		}
	default:
		return artifact, fmt.Errorf("build tool '%v' not supported", buildTool)
	}

	return artifact, nil
}
