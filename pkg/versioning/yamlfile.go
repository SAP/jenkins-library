package versioning

import (
	"fmt"
	"io/ioutil"
	"os"

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

func (j *YAMLfile) init() {
	if j.ReadFile == nil {
		j.ReadFile = ioutil.ReadFile
	}

	if j.WriteFile == nil {
		j.WriteFile = ioutil.WriteFile
	}
}

// GetVersion returns the current version of the artifact with a JSON build descriptor
func (j *YAMLfile) GetVersion(versionField string) (string, error) {
	j.init()

	content, err := j.ReadFile(j.Path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", j.Path)
	}

	err = yaml.Unmarshal(content, &j.Content)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read yaml content of file '%v'", j.Content)
	}

	return fmt.Sprint(j.Content[versionField]), nil
}

// SetVersion updates the version of the artifact with a JSON build descriptor
func (j *YAMLfile) SetVersion(versionField, version string) error {
	j.init()

	if j.Content == nil {
		_, err := j.GetVersion(versionField)
		if err != nil {
			return err
		}
	}
	j.Content[versionField] = version

	content, err := yaml.Marshal(j.Content)
	if err != nil {
		return errors.Wrapf(err, "failed to create yaml content for '%v'", j.Path)
	}
	err = j.WriteFile(j.Path, content, 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%v'", j.Path)
	}

	return nil
}
