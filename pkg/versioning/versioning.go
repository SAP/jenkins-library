package versioning

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// Coordinates to address the artifact coordinates like groupId, artifactId, version and packaging
type Coordinates struct {
	GroupID    string
	ArtifactID string
	Version    string
	Packaging  string
}

// Artifact defines the versioning operations for various build tools
type Artifact interface {
	VersioningScheme() string
	GetVersion() (string, error)
	SetVersion(string) error
	GetCoordinates() (Coordinates, error)
}

// Options define build tool specific settings in order to properly retrieve e.g. the version / coordinates of an artifact
type Options struct {
	ProjectSettingsFile     string
	DockerImage             string
	GlobalSettingsFile      string
	M2Path                  string
	Defines                 []string
	VersionSource           string
	VersionSection          string
	VersionField            string
	VersioningScheme        string
	HelmUpdateAppVersion    bool
	CAPVersioningPreference string
}

// Utils defines the versioning operations for various build tools
type Utils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error

	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
	Glob(pattern string) (matches []string, err error)
	FileExists(filename string) (bool, error)
	Copy(src, dest string) (int64, error)
	MkdirAll(path string, perm os.FileMode) error
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRead(path string) ([]byte, error)
	FileRemove(path string) error
}

type mvnRunner struct{}

func (m *mvnRunner) Execute(options *maven.ExecuteOptions, utils maven.Utils) (string, error) {
	return maven.Execute(options, utils)
}
func (m *mvnRunner) Evaluate(options *maven.EvaluateOptions, expression string, utils maven.Utils) (string, error) {
	return maven.Evaluate(options, expression, utils)
}

var fileExists func(string) (bool, error)

// GetArtifact returns the build tool specific implementation for retrieving version, etc. of an artifact
func GetArtifact(buildTool, buildDescriptorFilePath string, opts *Options, utils Utils) (Artifact, error) {
	var artifact Artifact
	if fileExists == nil {
		fileExists = piperutils.FileExists
	}

	// CAPVersioningPreference can only be 'maven' or 'npm'. Verification done on artifactPrepareVersion.yaml level
	if buildTool == "CAP" {
		buildTool = opts.CAPVersioningPreference
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
			utils:            utils,
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
	case "gradle":
		if len(buildDescriptorFilePath) == 0 {
			buildDescriptorFilePath = "gradle.properties"
		}
		artifact = &Gradle{
			path:         buildDescriptorFilePath,
			versionField: opts.VersionField,
			utils:        utils,
		}
	case "golang":
		if len(buildDescriptorFilePath) == 0 {
			var err error
			buildDescriptorFilePath, err = searchDescriptor([]string{"go.mod", "VERSION", "version.txt"}, fileExists)
			if err != nil {
				return artifact, err
			}
		}

		switch buildDescriptorFilePath {
		case "go.mod":
			artifact = &GoMod{path: buildDescriptorFilePath, fileExists: fileExists}
			break
		default:
			artifact = &Versionfile{path: buildDescriptorFilePath}
		}
	case "helm":
		artifact = &HelmChart{
			path:             buildDescriptorFilePath,
			utils:            utils,
			updateAppVersion: opts.HelmUpdateAppVersion,
		}
	case "maven":
		if len(buildDescriptorFilePath) == 0 {
			buildDescriptorFilePath = "pom.xml"
		}
		artifact = &Maven{
			runner: &mvnRunner{},
			utils:  utils,
			options: maven.EvaluateOptions{
				PomPath:             buildDescriptorFilePath,
				ProjectSettingsFile: opts.ProjectSettingsFile,
				GlobalSettingsFile:  opts.GlobalSettingsFile,
				M2Path:              opts.M2Path,
				Defines:             opts.Defines,
			},
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
	case "npm", "yarn":
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
			buildDescriptorFilePath, err = searchDescriptor([]string{"setup.py", "version.txt", "VERSION"}, fileExists)
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
			var err error
			buildDescriptorFilePath, err = searchDescriptor([]string{"sbtDescriptor.json", "build.sbt"}, fileExists)
			if err != nil {
				return artifact, err
			}
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
