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
	path         string
	content      map[string]interface{}
	versionField string
	readFile     func(string) ([]byte, error)
	writeFile    func(string, []byte, os.FileMode) error
}

func (j *JSONfile) init() {
	if len(j.versionField) == 0 {
		j.versionField = "version"
	}
	if j.readFile == nil {
		j.readFile = ioutil.ReadFile
	}

	if j.writeFile == nil {
		j.writeFile = ioutil.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (j *JSONfile) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact with a JSON-based build descriptor
func (j *JSONfile) GetVersion() (string, error) {
	j.init()

	content, err := j.readFile(j.path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", j.path)
	}

	err = json.Unmarshal(content, &j.content)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read json content of file '%v'", j.content)
	}

	return fmt.Sprint(j.content[j.versionField]), nil
}

// SetVersion updates the version of the artifact with a JSON-based build descriptor
func (j *JSONfile) SetVersion(version string) error {
	j.init()

	if j.content == nil {
		_, err := j.GetVersion()
		if err != nil {
			return err
		}
	}
	j.content[j.versionField] = version

	content, err := json.MarshalIndent(j.content, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "failed to create json content for '%v'", j.path)
	}
	err = j.writeFile(j.path, content, 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%v'", j.path)
	}

	return nil
}

// GetCoordinates returns the coordinates
func (j *JSONfile) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	projectVersion, err := j.GetVersion()
	if err != nil {
		return result, err
	}
	projectName := j.content["name"].(string)

	result.ArtifactID = projectName
	result.Version = projectVersion

	return result, nil
}
