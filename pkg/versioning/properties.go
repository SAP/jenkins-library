package versioning

import (
	"bytes"
	"fmt"
	"os"

	"github.com/magiconair/properties"
)

// PropertiesFile defines an artifact using a properties file for versioning
type PropertiesFile struct {
	path             string
	content          *properties.Properties
	versioningScheme string
	versionField     string
	writeFile        func(string, []byte, os.FileMode) error
}

func (p *PropertiesFile) init() error {
	if len(p.versionField) == 0 {
		p.versionField = "version"
	}
	if p.writeFile == nil {
		p.writeFile = os.WriteFile
	}
	if p.content == nil {
		var err error
		p.content, err = properties.LoadFile(p.path, properties.UTF8)
		if err != nil {
			return fmt.Errorf("failed to load file %v: %w", p.path, err)
		}
	}
	return nil
}

// VersioningScheme returns the relevant versioning scheme
func (p *PropertiesFile) VersioningScheme() string {
	if len(p.versioningScheme) == 0 {
		return "semver2"
	}
	return p.versioningScheme
}

// GetVersion returns the current version of the artifact with a ini-file-based build descriptor
func (p *PropertiesFile) GetVersion() (string, error) {
	err := p.init()
	if err != nil {
		return "", err
	}
	version := p.content.GetString(p.versionField, "")
	if len(version) == 0 {
		return "", fmt.Errorf("no version found in field %v", p.versionField)
	}
	return version, nil
}

// SetVersion updates the version of the artifact with a ini-file-based build descriptor
func (p *PropertiesFile) SetVersion(version string) error {
	err := p.init()
	if err != nil {
		return err
	}
	err = p.content.SetValue(p.versionField, version)
	if err != nil {
		return fmt.Errorf("failed to set version: %w", err)
	}

	var propsContent bytes.Buffer
	_, err = p.content.Write(&propsContent, properties.UTF8)
	if err != nil {
		return fmt.Errorf("failed to write version: %w", err)
	}
	err = p.writeFile(p.path, propsContent.Bytes(), 0666)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// GetCoordinates returns the coordinates
func (p *PropertiesFile) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	return result, nil
}
