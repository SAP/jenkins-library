package versioning

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// YAMLfile defines an npm artifact used for versioning
type YAMLfile struct {
	Path      string
	Content   map[string]interface{}
	ReadFile  func(string) ([]byte, error)
	WriteFile func(string, []byte, os.FileMode) error
}

func (y *YAMLfile) init() {
	if y.ReadFile == nil {
		y.ReadFile = ioutil.ReadFile
	}

	if y.WriteFile == nil {
		y.WriteFile = ioutil.WriteFile
	}
}

// GetVersion returns the current version of the artifact with a JSON build descriptor
func (y *YAMLfile) GetVersion(versionField string) (string, error) {
	y.init()

	content, err := y.ReadFile(y.Path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", y.Path)
	}

	err = yaml.Unmarshal(content, &y.Content)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read yaml content of file '%v'", y.Content)
	}

	return strings.TrimSpace(fmt.Sprint(y.Content[versionField])), nil
}

// SetVersion updates the version of the artifact with a JSON build descriptor
func (y *YAMLfile) SetVersion(versionField, version string) error {
	y.init()

	if y.Content == nil {
		_, err := y.GetVersion(versionField)
		if err != nil {
			return err
		}
	}
	y.Content[versionField] = version

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
