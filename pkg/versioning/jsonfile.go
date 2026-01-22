package versioning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/iancoleman/orderedmap"
	"github.com/pkg/errors"
)

// JSONfile defines an artifact using a json file for versioning
type JSONfile struct {
	path         string
	content      *orderedmap.OrderedMap
	versionField string
	readFile     func(string) ([]byte, error)
	writeFile    func(string, []byte, os.FileMode) error
}

func (j *JSONfile) init() {
	if len(j.versionField) == 0 {
		j.versionField = "version"
	}
	if j.readFile == nil {
		j.readFile = os.ReadFile
	}

	if j.writeFile == nil {
		j.writeFile = os.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (j *JSONfile) VersioningScheme() string {
	return "npm11-cloud"
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

	version, _ := j.content.Get(j.versionField)

	return fmt.Sprint(version), nil
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
	j.content.Set(j.versionField, version)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(j.content); err != nil {
		return errors.Wrapf(err, "failed to create json content for '%v'", j.path)
	}
	err := j.writeFile(j.path, buf.Bytes(), 0700)
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
	projectName, _ := j.content.Get("name")

	result.ArtifactID = fmt.Sprint(projectName)
	result.Version = projectVersion

	return result, nil
}
