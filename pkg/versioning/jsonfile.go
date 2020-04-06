package versioning

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// JSONfile defines an artifact using a json file for versioning
type JSONfile struct {
	Path      string
	Content   map[string]interface{}
	ReadFile  func(string) ([]byte, error)
	WriteFile func(string, []byte, os.FileMode) error
}

func (j *JSONfile) init() {
	if j.ReadFile == nil {
		j.ReadFile = ioutil.ReadFile
	}

	if j.WriteFile == nil {
		j.WriteFile = ioutil.WriteFile
	}
}

// GetVersion returns the current version of the artifact with a JSON build descriptor
func (j *JSONfile) GetVersion(versionField string) (string, error) {
	j.init()

	content, err := j.ReadFile(j.Path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", j.Path)
	}

	err = json.Unmarshal(content, &j.Content)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read json content of file '%v'", j.Content)
	}

	return fmt.Sprint(j.Content[versionField]), nil
}

// SetVersion updates the version of the artifact with a JSON build descriptor
func (j *JSONfile) SetVersion(versionField, version string) error {
	j.init()

	if j.Content == nil {
		_, err := j.GetVersion(versionField)
		if err != nil {
			return err
		}
	}
	j.Content[versionField] = version

	content, err := json.MarshalIndent(j.Content, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "failed to create json content for '%v'", j.Path)
	}
	err = j.WriteFile(j.Path, content, 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%v'", j.Path)
	}

	return nil
}
