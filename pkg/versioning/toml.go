package versioning

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
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

func (p *Toml) GetName() (string, error) {
	if !strings.Contains(p.Pip.path, TomlBuildDescriptor) {
		return "", fmt.Errorf("file '%v' is not a %s", p.Pip.path, TomlBuildDescriptor)
	}

	if err := p.init(); err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", p.Pip.path)
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

func (p *Toml) GetVersion() (string, error) {
	if !strings.Contains(p.Pip.path, TomlBuildDescriptor) {
		return "", fmt.Errorf("file '%v' is not a %s", p.Pip.path, TomlBuildDescriptor)
	}

	if err := p.init(); err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", p.Pip.path)
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

func (p *Toml) SetVersion(new string) error {
	current, err := p.GetVersion()
	if err != nil {
		return err
	}

	p.Pip.buildDescriptorContent = strings.ReplaceAll(p.Pip.buildDescriptorContent, fmt.Sprintf("version = '%v'", current), fmt.Sprintf("version = '%v'", new))
	p.Pip.buildDescriptorContent = strings.ReplaceAll(p.Pip.buildDescriptorContent, fmt.Sprintf("version = \"%v\"", current), fmt.Sprintf("version = \"%v\"", new))
	p.Pip.writeFile(p.Pip.path, []byte(p.Pip.buildDescriptorContent), 0600)
	return nil
}

// GetCoordinates returns the pip build descriptor coordinates
func (p *Toml) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}

	if name, err := p.GetName(); err != nil {
		result.ArtifactID = ""
	} else {
		result.ArtifactID = name
	}

	if version, err := p.GetVersion(); err != nil {
		return result, errors.Wrap(err, "failed to retrieve coordinates")
	} else {
		result.Version = version
	}

	return result, nil
}

func hasMatch(value, regex string) bool {
	return evaluateResult(value, regex)
}
