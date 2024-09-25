package versioning

import (
	"os"
	"strings"

	"golang.org/x/mod/modfile"

	"github.com/pkg/errors"
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
			return errors.Wrapf(err, "failed to read file '%v'", m.path)
		}
		m.buildDescriptorContent = string(content)
	}
	return nil
}

// GetVersion returns the go.mod descriptor version property
func (m *GoMod) GetVersion() (string, error) {
	buildDescriptorFilePath := m.path
	var err error
	if strings.Contains(m.path, "go.mod") {
		buildDescriptorFilePath, err = searchDescriptor([]string{"version.txt", "VERSION"}, m.fileExists)
		if err != nil {
			err = m.init()
			if err != nil {
				return "", errors.Wrapf(err, "failed to read file '%v'", m.path)
			}

			parsed, err := modfile.Parse(m.path, []byte(m.buildDescriptorContent), nil)
			if err != nil {
				return "", errors.Wrap(err, "failed to parse go.mod file")
			}
			if parsed.Module.Mod.Version != "" {
				return parsed.Module.Mod.Version, nil
			}

			return "", errors.Wrap(err, "failed to retrieve version")
		}
	}
	artifact := &Versionfile{
		path:             buildDescriptorFilePath,
		versioningScheme: m.VersioningScheme(),
	}
	return artifact.GetVersion()
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
		return result, errors.Wrap(err, "failed to parse go.mod file")
	}

	if parsed.Module == nil {
		return result, errors.Wrap(err, "failed to parse go.mod file")
	}
	if parsed.Module.Mod.Path != "" {
		artifactSplit := strings.Split(parsed.Module.Mod.Path, "/")
		artifactID := artifactSplit[len(artifactSplit)-1]
		result.ArtifactID = artifactID

		goModPath := parsed.Module.Mod.Path
		lastIndex := strings.LastIndex(goModPath, "/")
		groupID := goModPath[0:lastIndex]
		result.GroupID = groupID
	}
	result.Version = parsed.Module.Mod.Version
	if result.Version == "" {
		result.Version = "unspecified"
	}
	return result, nil
}
