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
	path         string
	content      map[string]interface{}
	versionField string
	readFile     func(string) ([]byte, error)
	writeFile    func(string, []byte, os.FileMode) error
}

func (y *YAMLfile) init() {
	if len(y.versionField) == 0 {
		y.versionField = "version"
	}
	if y.readFile == nil {
		y.readFile = ioutil.ReadFile
	}

	if y.writeFile == nil {
		y.writeFile = ioutil.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (y *YAMLfile) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact with a YAML-based build descriptor
func (y *YAMLfile) GetVersion() (string, error) {
	y.init()

	content, err := y.readFile(y.path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", y.path)
	}

	err = yaml.Unmarshal(content, &y.content)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read yaml content of file '%v'", y.content)
	}

	return strings.TrimSpace(fmt.Sprint(y.content[y.versionField])), nil
}

// SetVersion updates the version of the artifact with a YAML-based build descriptor
func (y *YAMLfile) SetVersion(version string) error {
	y.init()

	if y.content == nil {
		_, err := y.GetVersion()
		if err != nil {
			return err
		}
	}
	y.content[y.versionField] = version

	content, err := yaml.Marshal(y.content)
	if err != nil {
		return errors.Wrapf(err, "failed to create yaml content for '%v'", y.path)
	}
	err = y.writeFile(y.path, content, 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%v'", y.path)
	}

	return nil
}
