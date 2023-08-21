package versioning

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

// INIfile defines an artifact using a ini file for versioning
type INIfile struct {
	path             string
	content          *ini.File
	versioningScheme string
	versionSection   string
	versionField     string
	readFile         func(string) ([]byte, error)
	writeFile        func(string, []byte, os.FileMode) error
}

func (i *INIfile) init() error {
	if len(i.versionField) == 0 {
		i.versionField = "version"
	}
	if i.readFile == nil {
		i.readFile = os.ReadFile
	}
	if i.writeFile == nil {
		i.writeFile = os.WriteFile
	}
	if i.content == nil {
		conf, err := i.readFile(i.path)
		if err != nil {
			return errors.Wrapf(err, "failed to read file '%v'", i.path)
		}
		i.content, err = ini.Load(conf)
		if err != nil {
			return errors.Wrapf(err, "failed to load content from file '%v'", i.path)
		}
	}
	return nil
}

// VersioningScheme returns the relevant versioning scheme
func (i *INIfile) VersioningScheme() string {
	if len(i.versioningScheme) == 0 {
		return "semver2"
	}
	return i.versioningScheme
}

// GetVersion returns the current version of the artifact with a ini-file-based build descriptor
func (i *INIfile) GetVersion() (string, error) {
	if i.content == nil {
		err := i.init()
		if err != nil {
			return "", err
		}
	}
	section := i.content.Section(i.versionSection)
	if section.HasKey(i.versionField) {
		return section.Key(i.versionField).String(), nil
	}
	return "", fmt.Errorf("field '%v' not found in section '%v'", i.versionField, i.versionSection)
}

// SetVersion updates the version of the artifact with a ini-file-based build descriptor
func (i *INIfile) SetVersion(version string) error {
	if i.content == nil {
		err := i.init()
		if err != nil {
			return err
		}
	}
	section := i.content.Section(i.versionSection)
	section.Key(i.versionField).SetValue(version)
	var buf bytes.Buffer
	i.content.WriteTo(&buf)
	err := i.writeFile(i.path, buf.Bytes(), 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%v'", i.path)
	}
	return nil
}

// GetCoordinates returns the coordinates
func (i *INIfile) GetCoordinates() (Coordinates, error) {
	return Coordinates{}, nil
}
