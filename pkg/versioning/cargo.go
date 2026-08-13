package versioning

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

const CargoBuildDescriptor = "Cargo.toml"

// Cargo holds the content of a Cargo.toml build descriptor
type Cargo struct {
	path        string
	readFile    func(string) ([]byte, error)
	writeFile   func(string, []byte, os.FileMode) error
	coordinates cargoCoordinates
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
	content, err := c.readFile(c.path)
	if err != nil {
		return fmt.Errorf("failed to read file '%v': %w", c.path, err)
	}
	var coords cargoCoordinates
	if _, err := toml.Decode(string(content), &coords); err != nil {
		return fmt.Errorf("failed to parse file '%v': %w", c.path, err)
	}
	c.coordinates = coords
	return nil
}

// VersioningScheme returns the versioning scheme
func (c *Cargo) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the version from Cargo.toml
func (c *Cargo) GetVersion() (string, error) {
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
	content, err := c.readFile(c.path)
	if err != nil {
		return fmt.Errorf("failed to read file '%v': %w", c.path, err)
	}
	updated := strings.ReplaceAll(string(content),
		fmt.Sprintf("version = \"%v\"", current),
		fmt.Sprintf("version = \"%v\"", newVersion))
	updated = strings.ReplaceAll(updated,
		fmt.Sprintf("version = '%v'", current),
		fmt.Sprintf("version = '%v'", newVersion))
	if err := c.writeFile(c.path, []byte(updated), 0600); err != nil {
		return fmt.Errorf("failed to write file '%v': %w", c.path, err)
	}
	return nil
}

// GetCoordinates returns the artifact coordinates from Cargo.toml
func (c *Cargo) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	if err := c.init(); err != nil {
		return result, err
	}
	if len(c.coordinates.Package.Name) == 0 {
		return result, fmt.Errorf("no name information found in file '%v'", c.path)
	}
	result.ArtifactID = c.coordinates.Package.Name
	result.Version = c.coordinates.Package.Version
	return result, nil
}
