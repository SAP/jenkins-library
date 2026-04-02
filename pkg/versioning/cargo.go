package versioning

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	CargoBuildDescriptor = "Cargo.toml"
)

// Cargo utility to interact with Rust/Cargo specific versioning
type Cargo struct {
	path                   string
	readFile               func(string) ([]byte, error)
	writeFile              func(string, []byte, os.FileMode) error
	fileExists             func(string) (bool, error)
	buildDescriptorContent string
	coordinates            cargoCoordinates
}

type cargoCoordinates struct {
	Package struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
	} `toml:"package"`
}

func (c *Cargo) init() error {
	if c.readFile == nil {
		c.readFile = os.ReadFile
	}
	if c.writeFile == nil {
		c.writeFile = os.WriteFile
	}
	if len(c.buildDescriptorContent) == 0 {
		content, err := c.readFile(c.path)
		if err != nil {
			return fmt.Errorf("failed to read file '%v': %w", c.path, err)
		}
		c.buildDescriptorContent = string(content)
	}

	var coordinates cargoCoordinates
	if _, err := toml.Decode(c.buildDescriptorContent, &coordinates); err != nil {
		return fmt.Errorf("failed to parse '%v': %w", c.path, err)
	}
	c.coordinates = coordinates
	return nil
}

// GetVersion returns the version from VERSION/version.txt if present, else from Cargo.toml [package].version
func (c *Cargo) GetVersion() (string, error) {
	// Check for VERSION/version.txt overrides first
	if c.fileExists != nil {
		for _, vf := range []string{"VERSION", "version.txt"} {
			exists, _ := c.fileExists(vf)
			if exists {
				versionfile := &Versionfile{
					path:             vf,
					versioningScheme: c.VersioningScheme(),
					readFile:         c.readFile,
				}
				return versionfile.GetVersion()
			}
		}
	}

	if err := c.init(); err != nil {
		return "", err
	}
	if len(c.coordinates.Package.Version) == 0 {
		return "", fmt.Errorf("no version information found in file '%v'", c.path)
	}
	return c.coordinates.Package.Version, nil
}

// SetVersion updates the version in Cargo.toml
func (c *Cargo) SetVersion(newVersion string) error {
	current, err := c.GetVersion()
	if err != nil {
		return err
	}
	// Re-init to ensure buildDescriptorContent is loaded from Cargo.toml
	if err := c.init(); err != nil {
		return err
	}
	// Replace with double quotes
	c.buildDescriptorContent = strings.ReplaceAll(
		c.buildDescriptorContent,
		fmt.Sprintf("version = \"%v\"", current),
		fmt.Sprintf("version = \"%v\"", newVersion))
	// Replace with single quotes
	c.buildDescriptorContent = strings.ReplaceAll(
		c.buildDescriptorContent,
		fmt.Sprintf("version = '%v'", current),
		fmt.Sprintf("version = '%v'", newVersion))

	if err := c.writeFile(c.path, []byte(c.buildDescriptorContent), 0600); err != nil {
		return fmt.Errorf("failed to write file '%v': %w", c.path, err)
	}
	return nil
}

// VersioningScheme returns the relevant versioning scheme
func (c *Cargo) VersioningScheme() string {
	return "semver2"
}

// GetCoordinates returns the Cargo.toml build descriptor coordinates
func (c *Cargo) GetCoordinates() (Coordinates, error) {
	if err := c.init(); err != nil {
		return Coordinates{}, err
	}
	version, err := c.GetVersion()
	if err != nil {
		return Coordinates{}, fmt.Errorf("failed to retrieve coordinates: %w", err)
	}
	return Coordinates{
		ArtifactID: c.coordinates.Package.Name,
		GroupID:    "",
		Version:    version,
	}, nil
}
