package versioning

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	TomlBuildDescriptor = "pyproject.toml"
	// TomlNameRegex is used to match the pip descriptor artifact name
	TomlNameRegex = "(?s).*?name = ['\"](.*?)['\"].*"
	// TomlVersionRegex is used to match the pip descriptor artifact version
	TomlVersionRegex = "(?s).*?version = ['\"](.*?)['\"].*"
)

// Pip utility to interact with Python specific versioning
type Toml struct {
	Pip
}

func (p *Toml) init() error {
	return p.Pip.init()
}

// GetName returns the name from the build descriptor
func (p *Toml) GetName() (string, error) {
	if !strings.Contains(p.Pip.path, TomlBuildDescriptor) {
		return "", fmt.Errorf("file '%v' is not a %s", p.Pip.path, TomlBuildDescriptor)
	}

	if err := p.init(); err != nil {
		return "", fmt.Errorf("failed to read file '%v': %w", p.Pip.path, err)
	}
	if !hasMatch(p.Pip.buildDescriptorContent, TomlNameRegex) {
		return "", fmt.Errorf("no name information found in file '%v'", p.Pip.path)
	}
	values := regexp.MustCompile(TomlNameRegex).FindStringSubmatch(p.Pip.buildDescriptorContent)
	if len(values) < 2 {
		return "", fmt.Errorf("no name information found in file '%v'", p.Pip.path)
	}
	return values[1], nil
}

// GetVersion returns the current version from the build descriptor
func (p *Toml) GetVersion() (string, error) {
	if !strings.Contains(p.Pip.path, TomlBuildDescriptor) {
		return "", fmt.Errorf("file '%v' is not a %s", p.Pip.path, TomlBuildDescriptor)
	}

	if err := p.init(); err != nil {
		return "", fmt.Errorf("failed to read file '%v': %w", p.Pip.path, err)
	}
	if !hasMatch(p.Pip.buildDescriptorContent, TomlVersionRegex) {
		return "", fmt.Errorf("no version information found in file '%v'", p.Pip.path)
	}
	values := regexp.MustCompile(TomlVersionRegex).FindStringSubmatch(p.Pip.buildDescriptorContent)
	if len(values) < 2 {
		return "", fmt.Errorf("no version information found in file '%v'", p.Pip.path)
	}
	return values[1], nil
}

// SetVersion updates the version in the build descriptor
func (p *Toml) SetVersion(new string) error {
	if current, err := p.GetVersion(); err != nil {
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
		p.Pip.writeFile(p.Pip.path, []byte(p.Pip.buildDescriptorContent), 0600)
		return nil
	} else {
		return err
	}
}

// GetCoordinates returns the build descriptor coordinates
func (p *Toml) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	// get name
	if name, err := p.GetName(); err != nil {
		result.ArtifactID = ""
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

func hasMatch(value, regex string) bool {
	return evaluateResult(value, regex)
}
