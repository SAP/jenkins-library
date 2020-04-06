package versioning

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Versionfile defines an artifact containing the version in a file, e.g. VERSION
type Versionfile struct {
	Path      string
	ReadFile  func(string) ([]byte, error)
	WriteFile func(string, []byte, os.FileMode) error
}

func (v *Versionfile) init() {
	if len(v.Path) == 0 {
		v.Path = "VERSION"
	}
	if v.ReadFile == nil {
		v.ReadFile = ioutil.ReadFile
	}

	if v.WriteFile == nil {
		v.WriteFile = ioutil.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (v *Versionfile) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact
func (v *Versionfile) GetVersion() (string, error) {
	v.init()

	content, err := v.ReadFile(v.Path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", v.Path)
	}

	return strings.TrimSpace(string(content)), nil
}

// SetVersion updates the version of the artifact
func (v *Versionfile) SetVersion(version string) error {
	v.init()

	err := v.WriteFile(v.Path, []byte(version), 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%v'", v.Path)
	}

	return nil
}
