package versioning

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

// GoMod utility to interact with Go Modules specific versioning
type GoMod struct {
	path                   string
	readFile               func(string) ([]byte, error)
	writeFile              func(string, []byte, os.FileMode) error
	fileExists             func(string) (bool, error)
	buildDescriptorContent string
}

func (m *GoMod) init() error {
	if m.readFile == nil {
		m.readFile = os.ReadFile
	}
	if m.writeFile == nil {
		m.writeFile = os.WriteFile
	}
	if len(m.buildDescriptorContent) == 0 {
		content, err := m.readFile(m.path)
		if err != nil {
			return fmt.Errorf("failed to read file '%v': %w", m.path, err)
		}
		m.buildDescriptorContent = string(content)
	}
	return nil
}

// GetVersion returns the version from a VERSION file, or falls back to go.mod
func (m *GoMod) GetVersion() (string, error) {
	// If path is not go.mod, just read it as a version file
	if !strings.Contains(m.path, "go.mod") {
		return m.readVersionFile(m.path)
	}

	// For go.mod projects, first try to find a dedicated version file
	versionFile, versionFileErr := m.findVersionFile()
	if versionFileErr == nil {
		return m.readVersionFile(versionFile)
	}

	// Fall back to extracting version from go.mod itself
	version, gomodErr := m.extractVersionFromGoMod()
	if gomodErr != nil {
		return "", fmt.Errorf("no version file found (%v) and %w", versionFileErr, gomodErr)
	}
	if version == "" {
		return "", fmt.Errorf("no version file found (%v) and go.mod has no version", versionFileErr)
	}
	return version, nil
}

func (m *GoMod) findVersionFile() (string, error) {
	return searchDescriptor([]string{"version.txt", "VERSION"}, m.fileExists)
}

func (m *GoMod) readVersionFile(path string) (string, error) {
	artifact := &Versionfile{
		path:             path,
		versioningScheme: m.VersioningScheme(),
	}
	return artifact.GetVersion()
}

func (m *GoMod) extractVersionFromGoMod() (string, error) {
	if err := m.init(); err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	parsed, err := modfile.Parse(m.path, []byte(m.buildDescriptorContent), nil)
	if err != nil {
		return "", fmt.Errorf("failed to parse go.mod: %w", err)
	}

	return parsed.Module.Mod.Version, nil
}

// SetVersion sets the go.mod descriptor version property
func (m *GoMod) SetVersion(v string) error {
	return nil
}

// VersioningScheme returns the relevant versioning scheme
func (m *GoMod) VersioningScheme() string {
	return "semver2"
}

// GetCoordinates returns the go.mod build descriptor coordinates
func (m *GoMod) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	err := m.init()
	if err != nil {
		return result, err
	}

	parsed, err := modfile.Parse(m.path, []byte(m.buildDescriptorContent), nil)
	if err != nil {
		return result, fmt.Errorf("failed to parse go.mod file: %w", err)
	}

	if parsed.Module == nil {
		return result, errors.New("failed to parse go.mod file: no module found")
	}

	// validate module path as defined by golang
	if err = module.CheckPath(parsed.Module.Mod.Path); err != nil {
		return result, fmt.Errorf("failed to parse go.mod file: %w", err)
	}

	if parsed.Module.Mod.Path != "" {
		modulePath := parsed.Module.Mod.Path
		separatorIndex := strings.LastIndex(modulePath, "/")
		result.ArtifactID = modulePath[separatorIndex+1:]

		// extract groupID from module path
		if separatorIndex >= 0 {
			result.GroupID = modulePath[0:separatorIndex]
		}
	}

	result.Version = parsed.Module.Mod.Version
	if result.Version == "" {
		result.Version = "unspecified"
	}
	return result, nil
}
