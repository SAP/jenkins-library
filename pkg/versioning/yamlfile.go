package versioning

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// YAMLfile defines an artifact using a yaml file for versioning
type YAMLfile struct {
	Path         string
	Content      map[string]interface{}
	VersionField string
	ReadFile     func(string) ([]byte, error)
	WriteFile    func(string, []byte, os.FileMode) error
}

func (y *YAMLfile) init() {
	if len(y.VersionField) == 0 {
		y.VersionField = "version"
	}
	if y.ReadFile == nil {
		y.ReadFile = ioutil.ReadFile
	}

	if y.WriteFile == nil {
		y.WriteFile = ioutil.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (y *YAMLfile) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact with a YAML-based build descriptor
func (y *YAMLfile) GetVersion() (string, error) {
	y.init()

	content, err := y.ReadFile(y.Path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", y.Path)
	}

	err = yaml.Unmarshal(content, &y.Content)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read yaml content of file '%v'", y.Content)
	}

	return strings.TrimSpace(fmt.Sprint(y.Content[y.VersionField])), nil
}

// SetVersion updates the version of the artifact with a YAML-based build descriptor
func (y *YAMLfile) SetVersion(version string) error {
	y.init()

	if y.Content == nil {
		_, err := y.GetVersion()
		if err != nil {
			return err
		}
	}
	y.Content[y.VersionField] = version

	content, err := yaml.Marshal(y.Content)
	if err != nil {
		return errors.Wrapf(err, "failed to create yaml content for '%v'", y.Path)
	}
	err = y.WriteFile(y.Path, content, 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%v'", y.Path)
	}

	return nil
}
