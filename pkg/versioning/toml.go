package versioning

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	TomlBuildDescriptor = "pyproject.toml"
)

// Pip utility to interact with Python specific versioning
type Toml struct {
	Pip
	coordinates tomlCoordinates
}

type tomlCoordinates struct {
	Project struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
	} `toml:"project"`
}

func (p *Toml) init() error {
	var coordinates tomlCoordinates

	if !strings.Contains(p.Pip.path, TomlBuildDescriptor) {
		return fmt.Errorf("file '%v' is not a %s", p.Pip.path, TomlBuildDescriptor)
	}

	if err := p.Pip.init(); err != nil {
		return err
	}

	if _, err := toml.Decode(p.Pip.buildDescriptorContent, &coordinates); err != nil {
		return err
	}
	p.coordinates = coordinates
	return nil
}

// GetName returns the name from the build descriptor
func (p *Toml) GetName() (string, error) {
	if err := p.init(); err != nil {
		return "", fmt.Errorf("failed to read file '%v': %w", p.Pip.path, err)
	}
	if len(p.coordinates.Project.Name) == 0 {
		return "", fmt.Errorf("no name information found in file '%v'", p.Pip.path)
	}
	return p.coordinates.Project.Name, nil
}

// // GetVersion returns the current version from the build descriptor
func (p *Toml) GetVersion() (string, error) {
	if err := p.init(); err != nil {
		return "", fmt.Errorf("failed to read file '%v': %w", p.Pip.path, err)
	}
	if len(p.coordinates.Project.Version) == 0 {
		return "", fmt.Errorf("no version information found in file '%v'", p.Pip.path)
	}
	return p.coordinates.Project.Version, nil
}

// SetVersion updates the version in the build descriptor
func (p *Toml) SetVersion(new string) error {
	if current, err := p.GetVersion(); err != nil {
		return err
	} else {
		// replace with single quotes
		p.Pip.buildDescriptorContent = strings.ReplaceAll(
			p.Pip.buildDescriptorContent,
			fmt.Sprintf("version = '%v'", current),
			fmt.Sprintf("version = '%v'", new))
		// replace with double quotes as well
		p.Pip.buildDescriptorContent = strings.ReplaceAll(
			p.Pip.buildDescriptorContent,
			fmt.Sprintf("version = \"%v\"", current),
			fmt.Sprintf("version = \"%v\"", new))
		err = p.Pip.writeFile(p.Pip.path, []byte(p.Pip.buildDescriptorContent), 0600)
		if err != nil {
			return fmt.Errorf("failed to write file '%v': %w", p.Pip.path, err)
		}
		return nil
	}
}

// GetCoordinates returns the build descriptor coordinates
func (p *Toml) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	// get name
	if name, err := p.GetName(); err != nil {
		return result, fmt.Errorf("failed to retrieve coordinates: %w", err)
	} else {
		result.ArtifactID = name
	}
	// get version
	if version, err := p.GetVersion(); err != nil {
		return result, fmt.Errorf("failed to retrieve coordinates: %w", err)
	} else {
		result.Version = version
	}

	return result, nil
}
