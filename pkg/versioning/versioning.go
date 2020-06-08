package versioning

import (
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/maven"
)

// Coordinates to address the artifact
type Coordinates interface{}

// Artifact defines the versioning operations for various build tools
type Artifact interface {
	VersioningScheme() string
	GetVersion() (string, error)
	SetVersion(string) error
	GetCoordinates() (Coordinates, error)
}

// Options define build tool specific settings in order to properly retrieve e.g. the version of an artifact
type Options struct {
	ProjectSettingsFile string
	GlobalSettingsFile  string
	M2Path              string
	VersionSource       string
	VersionSection      string
	VersionField        string
	VersioningScheme    string
}

type mvnRunner struct{}

func (m *mvnRunner) Execute(options *maven.ExecuteOptions, execRunner mavenExecRunner) (string, error) {
	return maven.Execute(options, execRunner)
}
func (m *mvnRunner) Evaluate(pomFile, expression string, execRunner mavenExecRunner) (string, error) {
	return maven.Evaluate(pomFile, expression, execRunner)
}

var fileExists func(string) (bool, error)

// GetArtifact returns the build tool specific implementation for retrieving version, etc. of an artifact
func GetArtifact(buildTool, buildDescriptorFilePath string, opts *Options, execRunner mavenExecRunner) (Artifact, error) {
	var artifact Artifact
	if fileExists == nil {
		fileExists = piperutils.FileExists
	}
	switch buildTool {
	case "custom":
		var err error
		artifact, err = customArtifact(buildDescriptorFilePath, opts.VersionField, opts.VersionSection, opts.VersioningScheme)
		if err != nil {
			return artifact, err
		}
	case "docker":
		artifact = &Docker{
			execRunner:       execRunner,
			options:          opts,
			path:             buildDescriptorFilePath,
			versionSource:    opts.VersionSource,
			versioningScheme: opts.VersioningScheme,
		}
	case "dub":
		if len(buildDescriptorFilePath) == 0 {
			buildDescriptorFilePath = "dub.json"
		}
		artifact = &JSONfile{
			path:         buildDescriptorFilePath,
			versionField: "version",
		}
	case "golang":
		if len(buildDescriptorFilePath) == 0 {
			var err error
			buildDescriptorFilePath, err = searchDescriptor([]string{"VERSION", "version.txt"}, fileExists)
			if err != nil {
				return artifact, err
			}
		}
		artifact = &Versionfile{
			path: buildDescriptorFilePath,
		}
	case "maven":
		if len(buildDescriptorFilePath) == 0 {
			buildDescriptorFilePath = "pom.xml"
		}
		artifact = &Maven{
			runner:              &mvnRunner{},
			execRunner:          execRunner,
			pomPath:             buildDescriptorFilePath,
			projectSettingsFile: opts.ProjectSettingsFile,
			globalSettingsFile:  opts.GlobalSettingsFile,
			m2Path:              opts.M2Path,
		}
	case "mta":
		if len(buildDescriptorFilePath) == 0 {
			buildDescriptorFilePath = "mta.yaml"
		}
		artifact = &YAMLfile{
			path:            buildDescriptorFilePath,
			versionField:    "version",
			artifactIDField: "ID",
		}
	case "npm":
		if len(buildDescriptorFilePath) == 0 {
			buildDescriptorFilePath = "package.json"
		}
		artifact = &JSONfile{
			path:         buildDescriptorFilePath,
			versionField: "version",
		}
	case "pip":
		if len(buildDescriptorFilePath) == 0 {
			var err error
			buildDescriptorFilePath, err = searchDescriptor([]string{"version.txt", "VERSION", "setup.py"}, fileExists)
			if err != nil {
				return artifact, err
			}
		}
		artifact = &Pip{
			path:       buildDescriptorFilePath,
			fileExists: fileExists,
		}
	case "sbt":
		if len(buildDescriptorFilePath) == 0 {
			buildDescriptorFilePath = "sbtDescriptor.json"
		}
		artifact = &JSONfile{
			path:         buildDescriptorFilePath,
			versionField: "version",
		}
	default:
		return artifact, fmt.Errorf("build tool '%v' not supported", buildTool)
	}

	return artifact, nil
}

func searchDescriptor(supported []string, existsFunc func(string) (bool, error)) (string, error) {
	var descriptor string
	for _, f := range supported {
		exists, _ := existsFunc(f)
		if exists {
			descriptor = f
			break
		}
	}
	if len(descriptor) == 0 {
		return "", fmt.Errorf("no build descriptor available, supported: %v", supported)
	}
	return descriptor, nil
}

func customArtifact(buildDescriptorFilePath, field, section, scheme string) (Artifact, error) {
	switch filepath.Ext(buildDescriptorFilePath) {
	case ".cfg", ".ini":
		return &INIfile{
			path:             buildDescriptorFilePath,
			versionField:     field,
			versionSection:   section,
			versioningScheme: scheme,
		}, nil
	case ".json":
		return &JSONfile{
			path:         buildDescriptorFilePath,
			versionField: field,
		}, nil
	case ".yaml", ".yml":
		return &YAMLfile{
			path:         buildDescriptorFilePath,
			versionField: field,
		}, nil
	case ".txt", "":
		return &Versionfile{
			path:             buildDescriptorFilePath,
			versioningScheme: scheme,
		}, nil
	default:
		return nil, fmt.Errorf("file type not supported: '%v'", buildDescriptorFilePath)
	}
}
