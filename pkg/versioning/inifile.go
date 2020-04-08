package versioning

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

// INIfile defines an artifact using a json file for versioning
type INIfile struct {
	Path           string
	Content        *ini.File
	VersionSection string
	VersionField   string
	ReadFile       func(string) ([]byte, error)
	WriteFile      func(string, []byte, os.FileMode) error
}

func (i *INIfile) init() error {
	if len(i.VersionField) == 0 {
		i.VersionField = "version"
	}
	if i.ReadFile == nil {
		i.ReadFile = ioutil.ReadFile
	}
	if i.WriteFile == nil {
		i.WriteFile = ioutil.WriteFile
	}
	if i.Content == nil {
		conf, err := i.ReadFile(i.Path)
		if err != nil {
			return errors.Wrapf(err, "failed to read file '%v'", i.Path)
		}
		i.Content, err = ini.Load(conf)
		if err != nil {
			return errors.Wrapf(err, "failed to load content from file '%v'", i.Path)
		}
	}
	return nil
}

// VersioningScheme returns the relevant versioning scheme
func (i *INIfile) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact with a ini-file-based build descriptor
func (i *INIfile) GetVersion() (string, error) {
	if i.Content == nil {
		err := i.init()
		if err != nil {
			return "", err
		}
	}
	section := i.Content.Section(i.VersionSection)
	if section.HasKey(i.VersionField) {
		return section.Key(i.VersionField).String(), nil
	}
	return "", fmt.Errorf("field '%v' not found in section '%v'", i.VersionField, i.VersionSection)
}

// SetVersion updates the version of the artifact with a ini-file-based build descriptor
func (i *INIfile) SetVersion(version string) error {
	if i.Content == nil {
		err := i.init()
		if err != nil {
			return err
		}
	}
	section := i.Content.Section(i.VersionSection)
	section.Key(i.VersionField).SetValue(version)
	var buf bytes.Buffer
	i.Content.WriteTo(&buf)
	err := i.WriteFile(i.Path, buf.Bytes(), 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%v'", i.Path)
	}
	return nil
}
