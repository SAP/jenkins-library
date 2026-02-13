package versioning

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	// NameRegex is used to match the pip descriptor artifact name
	NameRegex = "(?s)(.*)name=['\"](.*?)['\"](.*)"
	// VersionRegex is used to match the pip descriptor artifact version
	VersionRegex = "(?s)(.*)version=['\"](.*?)['\"](.*)"
)

// Pip utility to interact with Python specific versioning
type Pip struct {
	path                   string
	readFile               func(string) ([]byte, error)
	writeFile              func(string, []byte, os.FileMode) error
	fileExists             func(string) (bool, error)
	buildDescriptorContent string
}

func (p *Pip) init() error {
	if p.readFile == nil {
		p.readFile = os.ReadFile
	}

	if p.writeFile == nil {
		p.writeFile = os.WriteFile
	}

	if len(p.buildDescriptorContent) == 0 {
		content, err := p.readFile(p.path)
		if err != nil {
			return fmt.Errorf("failed to read file '%v': %w", p.path, err)
		}
		p.buildDescriptorContent = string(content)
	}
	return nil
}

// GetVersion returns the Pip descriptor version property
func (p *Pip) GetVersion() (string, error) {
	buildDescriptorFilePath := p.path
	var err error
	if strings.Contains(p.path, "setup.py") {
		buildDescriptorFilePath, err = searchDescriptor([]string{"version.txt", "VERSION"}, p.fileExists)
		if err != nil {
			initErr := p.init()
			if initErr != nil {
				return "", fmt.Errorf("failed to read file '%v': %w", p.path, initErr)
			}
			if evaluateResult(p.buildDescriptorContent, VersionRegex) {
				compile := regexp.MustCompile(VersionRegex)
				values := compile.FindStringSubmatch(p.buildDescriptorContent)
				return values[2], nil
			}
			return "", fmt.Errorf("failed to retrieve version: %w", err)
		}
	}
	artifact := &Versionfile{
		path:             buildDescriptorFilePath,
		versioningScheme: p.VersioningScheme(),
		readFile:         p.readFile,
	}
	return artifact.GetVersion()
}

// SetVersion sets the Pip descriptor version property
func (p *Pip) SetVersion(v string) error {
	buildDescriptorFilePath := p.path
	var err error
	if strings.Contains(p.path, "setup.py") {
		buildDescriptorFilePath, err = searchDescriptor([]string{"version.txt", "VERSION"}, p.fileExists)
		if err != nil {
			initErr := p.init()
			if initErr != nil {
				return fmt.Errorf("failed to read file '%v': %w", p.path, initErr)
			}
			if evaluateResult(p.buildDescriptorContent, VersionRegex) {
				compile := regexp.MustCompile(VersionRegex)
				values := compile.FindStringSubmatch(p.buildDescriptorContent)
				p.buildDescriptorContent = strings.ReplaceAll(p.buildDescriptorContent, fmt.Sprintf("version='%v'", values[2]), fmt.Sprintf("version='%v'", v))
				p.buildDescriptorContent = strings.ReplaceAll(p.buildDescriptorContent, fmt.Sprintf("version=\"%v\"", values[2]), fmt.Sprintf("version=\"%v\"", v))
				p.writeFile(p.path, []byte(p.buildDescriptorContent), 0600)
			} else {
				return fmt.Errorf("failed to retrieve version: %w", err)
			}
		}
	}
	artifact := &Versionfile{
		path:             buildDescriptorFilePath,
		versioningScheme: p.VersioningScheme(),
		writeFile:        p.writeFile,
	}
	return artifact.SetVersion(v)
}

// VersioningScheme returns the relevant versioning scheme
func (p *Pip) VersioningScheme() string {
	return "pep440"
}

// GetCoordinates returns the pip build descriptor coordinates
func (p *Pip) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	err := p.init()
	if err != nil {
		return result, err
	}

	if evaluateResult(p.buildDescriptorContent, NameRegex) {
		compile := regexp.MustCompile(NameRegex)
		values := compile.FindStringSubmatch(p.buildDescriptorContent)
		result.ArtifactID = values[2]
	} else {
		result.ArtifactID = ""
	}

	result.Version, err = p.GetVersion()
	if err != nil {
		return result, fmt.Errorf("failed to retrieve coordinates: %w", err)
	}

	return result, nil
}

func evaluateResult(value, regex string) bool {
	if len(value) > 0 {
		match, _ := regexp.MatchString(regex, value)
		return match
	}
	return true
}
