package versioning

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Versionfile defines an artifact containing the version in a file, e.g. VERSION
type Versionfile struct {
	path             string
	readFile         func(string) ([]byte, error)
	writeFile        func(string, []byte, os.FileMode) error
	versioningScheme string
}

func (v *Versionfile) init() {
	if len(v.path) == 0 {
		v.path = "VERSION"
	}
	if v.readFile == nil {
		v.readFile = ioutil.ReadFile
	}

	if v.writeFile == nil {
		v.writeFile = ioutil.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (v *Versionfile) VersioningScheme() string {
	if len(v.versioningScheme) == 0 {
		return "semver2"
	}
	return v.versioningScheme
}

// GetVersion returns the current version of the artifact
func (v *Versionfile) GetVersion() (string, error) {
	v.init()

	content, err := v.readFile(v.path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", v.path)
	}

	return strings.TrimSpace(string(content)), nil
}

// SetVersion updates the version of the artifact
func (v *Versionfile) SetVersion(version string) error {
	v.init()

	err := v.writeFile(v.path, []byte(version), 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%v'", v.path)
	}

	return nil
}
