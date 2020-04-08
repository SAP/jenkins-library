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
	Path         string
	Content      map[string]interface{}
	VersionField string
	ReadFile     func(string) ([]byte, error)
	WriteFile    func(string, []byte, os.FileMode) error
}

func (j *JSONfile) init() {
	if len(j.VersionField) == 0 {
		j.VersionField = "version"
	}
	if j.ReadFile == nil {
		j.ReadFile = ioutil.ReadFile
	}

	if j.WriteFile == nil {
		j.WriteFile = ioutil.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (j *JSONfile) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact with a JSON-based build descriptor
func (j *JSONfile) GetVersion() (string, error) {
	j.init()

	content, err := j.ReadFile(j.Path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", j.Path)
	}

	err = json.Unmarshal(content, &j.Content)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read json content of file '%v'", j.Content)
	}

	return fmt.Sprint(j.Content[j.VersionField]), nil
}

// SetVersion updates the version of the artifact with a JSON-based build descriptor
func (j *JSONfile) SetVersion(version string) error {
	j.init()

	if j.Content == nil {
		_, err := j.GetVersion()
		if err != nil {
			return err
		}
	}
	j.Content[j.VersionField] = version

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
