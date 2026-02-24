package versioning

import (
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
)

// YAMLDescriptor holds the unique identifier combination for an artifact
type YAMLDescriptor struct {
	GroupID    string
	ArtifactID string
	Version    string
}

// YAMLfile defines an artifact using a yaml file for versioning
type YAMLfile struct {
	path            string
	content         map[string]interface{}
	versionField    string
	artifactIDField string
	readFile        func(string) ([]byte, error)
	writeFile       func(string, []byte, os.FileMode) error
}

func (y *YAMLfile) init() {
	if len(y.versionField) == 0 {
		y.versionField = "version"
	}
	if len(y.artifactIDField) == 0 {
		y.artifactIDField = "ID"
	}
	if y.readFile == nil {
		y.readFile = os.ReadFile
	}
	if y.writeFile == nil {
		y.writeFile = os.WriteFile
	}
}

func (y *YAMLfile) readContent() error {
	y.init()
	if y.content != nil {
		return nil
	}
	content, err := y.readFile(y.path)
	if err != nil {
		return fmt.Errorf("failed to read file '%v': %w", y.path, err)
	}
	err = yaml.Unmarshal(content, &y.content)
	if err != nil {
		return fmt.Errorf("failed to read yaml content of file '%v': %w", y.content, err)
	}
	return nil
}

func (y *YAMLfile) readField(key string) (string, error) {
	err := y.readContent()
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return strings.TrimSpace(fmt.Sprint(y.content[key])), nil
}

// VersioningScheme returns the relevant versioning scheme
func (y *YAMLfile) VersioningScheme() string {
	return "semver2"
}

// GetArtifactID returns the current ID of the artifact
func (y *YAMLfile) GetArtifactID() (string, error) {
	y.init()
	return y.readField(y.artifactIDField)
}

// GetVersion returns the current version of the artifact with a YAML-based build descriptor
func (y *YAMLfile) GetVersion() (string, error) {
	y.init()
	return y.readField(y.versionField)
}

// SetVersion updates the version of the artifact with a YAML-based build descriptor
func (y *YAMLfile) SetVersion(version string) error {
	err := y.readContent()
	if err != nil {
		return fmt.Errorf("failed to set version: %w", err)
	}

	y.content[y.versionField] = version

	content, err := yaml.Marshal(y.content)
	if err != nil {
		return fmt.Errorf("failed to create yaml content for '%v': %w", y.path, err)
	}
	err = y.writeFile(y.path, content, 0700)
	if err != nil {
		return fmt.Errorf("failed to write file '%v': %w", y.path, err)
	}

	return nil
}

// GetCoordinates returns the coordinates
func (y *YAMLfile) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	var err error
	result.ArtifactID, err = y.GetArtifactID()
	if err != nil {
		return result, err
	}
	result.Version, err = y.GetVersion()
	if err != nil {
		return result, err
	}
	return result, nil
}
