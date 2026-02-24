package versioning

import (
	"fmt"
	"os"
	"strings"
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
		v.readFile = os.ReadFile
	}

	if v.writeFile == nil {
		v.writeFile = os.WriteFile
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
		return "", fmt.Errorf("failed to read file '%v': %w", v.path, err)
	}

	return strings.TrimSpace(string(content)), nil
}

// SetVersion updates the version of the artifact
func (v *Versionfile) SetVersion(version string) error {
	v.init()

	err := v.writeFile(v.path, []byte(version), 0700)
	if err != nil {
		return fmt.Errorf("failed to write file '%v': %w", v.path, err)
	}

	return nil
}

// GetCoordinates returns the coordinates
func (v *Versionfile) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	return result, nil
}
