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
	content     []byte
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
	if len(c.content) > 0 {
		return nil
	}
	content, err := c.readFile(c.path)
	if err != nil {
		return fmt.Errorf("failed to read file '%v': %w", c.path, err)
	}
	var coords cargoCoordinates
	if _, err := toml.Decode(string(content), &coords); err != nil {
		return fmt.Errorf("failed to parse file '%v': %w", c.path, err)
	}
	c.content = content
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

// SetVersion updates the version in Cargo.toml.
// Replacement is scoped to the [package] section so that dependency entries
// that share the same version string are not inadvertently modified.
func (c *Cargo) SetVersion(newVersion string) error {
	current, err := c.GetVersion()
	if err != nil {
		return err
	}
	updated, err := replaceVersionInPackageSection(string(c.content), current, newVersion)
	if err != nil {
		return fmt.Errorf("failed to update version in file '%v': %w", c.path, err)
	}
	if err := c.writeFile(c.path, []byte(updated), 0600); err != nil {
		return fmt.Errorf("failed to write file '%v': %w", c.path, err)
	}
	c.content = []byte(updated)
	c.coordinates.Package.Version = newVersion
	return nil
}

// stripTOMLComment removes a trailing TOML comment (text from an unquoted '#' to end
// of line). A '#' inside a single- or double-quoted string is not a comment.
func stripTOMLComment(s string) string {
	inSingle, inDouble := false, false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return s[:i]
			}
		}
	}
	return s
}

// replaceVersionInPackageSection replaces the version line only within the [package]
// section of a Cargo.toml, leaving [dependencies] and other sections untouched.
// It processes the file line by line so that content inside other sections (or inside
// multi-line strings) that happens to start with "[" is never misidentified as a
// section boundary. Trailing TOML comments are stripped before matching, but the
// original line (including the comment) is preserved in the output.
func replaceVersionInPackageSection(content, current, newVersion string) (string, error) {
	lines := strings.Split(content, "\n")
	inPackage := false
	inMultilineStr := false
	for i, line := range lines {
		// Track triple-quoted multi-line strings so their content is never misread as
		// a section header. Each line with an odd number of `"""` toggles the state.
		if strings.Count(line, `"""`)%2 != 0 {
			inMultilineStr = !inMultilineStr
		}
		if inMultilineStr {
			continue
		}
		// Strip trailing comment before matching; the original line is used for replacement
		// so the comment and surrounding whitespace are preserved in the written file.
		trimmed := strings.TrimSpace(stripTOMLComment(line))
		if strings.HasPrefix(trimmed, "[") {
			inPackage = trimmed == "[package]"
			continue
		}
		if !inPackage {
			continue
		}
		key, val, found := strings.Cut(trimmed, "=")
		if !found || strings.TrimSpace(key) != "version" {
			continue
		}
		val = strings.TrimSpace(val)
		for _, quote := range []string{`"`, `'`} {
			oldVal := quote + current + quote
			if val == oldVal {
				lines[i] = strings.Replace(line, oldVal, quote+newVersion+quote, 1)
				return strings.Join(lines, "\n"), nil
			}
		}
	}
	return "", fmt.Errorf("version %q not found in [package] section", current)
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
	version, err := c.GetVersion()
	if err != nil {
		return result, err
	}
	result.ArtifactID = c.coordinates.Package.Name
	result.Version = version
	return result, nil
}
