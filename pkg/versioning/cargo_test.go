//go:build unit
// +build unit

package versioning

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCargoToml = `[package]
name = "my-rust-app"
version = "1.2.3"
edition = "2021"
`

const testCargoTomlSingleQuote = `[package]
name = "my-rust-app"
version = '1.2.3'
edition = "2021"
`

func TestCargoGetVersion(t *testing.T) {
	t.Parallel()

	t.Run("version from Cargo.toml", func(t *testing.T) {
		t.Parallel()
		c := &Cargo{
			path:     "Cargo.toml",
			readFile: func(string) ([]byte, error) { return []byte(testCargoToml), nil },
		}
		version, err := c.GetVersion()
		require.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("version from VERSION file override", func(t *testing.T) {
		t.Parallel()
		c := &Cargo{
			path:     "Cargo.toml",
			readFile: func(string) ([]byte, error) { return []byte(testCargoToml), nil },
			fileExists: func(path string) (bool, error) {
				return path == "VERSION", nil
			},
		}
		// Override readFile to also serve VERSION
		c.readFile = func(path string) ([]byte, error) {
			if path == "VERSION" {
				return []byte("2.0.0\n"), nil
			}
			return []byte(testCargoToml), nil
		}
		version, err := c.GetVersion()
		require.NoError(t, err)
		assert.Equal(t, "2.0.0", version)
	})

	t.Run("version from version.txt override", func(t *testing.T) {
		t.Parallel()
		c := &Cargo{
			path: "Cargo.toml",
			readFile: func(path string) ([]byte, error) {
				if path == "version.txt" {
					return []byte("3.0.0"), nil
				}
				return []byte(testCargoToml), nil
			},
			fileExists: func(path string) (bool, error) {
				if path == "VERSION" {
					return false, nil
				}
				return path == "version.txt", nil
			},
		}
		version, err := c.GetVersion()
		require.NoError(t, err)
		assert.Equal(t, "3.0.0", version)
	})

	t.Run("error when Cargo.toml missing", func(t *testing.T) {
		t.Parallel()
		c := &Cargo{
			path:     "Cargo.toml",
			readFile: func(string) ([]byte, error) { return nil, fmt.Errorf("file not found") },
		}
		_, err := c.GetVersion()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file 'Cargo.toml'")
	})

	t.Run("error when no version in Cargo.toml", func(t *testing.T) {
		t.Parallel()
		c := &Cargo{
			path:     "Cargo.toml",
			readFile: func(string) ([]byte, error) { return []byte("[package]\nname = \"foo\"\n"), nil },
		}
		_, err := c.GetVersion()
		assert.EqualError(t, err, "no version information found in file 'Cargo.toml'")
	})
}

func TestCargoSetVersion(t *testing.T) {
	t.Parallel()

	t.Run("set version with double quotes", func(t *testing.T) {
		t.Parallel()
		var writtenContent []byte
		c := &Cargo{
			path:     "Cargo.toml",
			readFile: func(string) ([]byte, error) { return []byte(testCargoToml), nil },
			writeFile: func(_ string, content []byte, _ os.FileMode) error {
				writtenContent = content
				return nil
			},
		}
		err := c.SetVersion("1.3.0")
		require.NoError(t, err)
		assert.Contains(t, string(writtenContent), `version = "1.3.0"`)
		assert.NotContains(t, string(writtenContent), `version = "1.2.3"`)
	})

	t.Run("set version with single quotes", func(t *testing.T) {
		t.Parallel()
		var writtenContent []byte
		c := &Cargo{
			path:     "Cargo.toml",
			readFile: func(string) ([]byte, error) { return []byte(testCargoTomlSingleQuote), nil },
			writeFile: func(_ string, content []byte, _ os.FileMode) error {
				writtenContent = content
				return nil
			},
		}
		err := c.SetVersion("1.3.0")
		require.NoError(t, err)
		assert.Contains(t, string(writtenContent), `version = '1.3.0'`)
	})

	t.Run("error when Cargo.toml missing", func(t *testing.T) {
		t.Parallel()
		c := &Cargo{
			path:     "Cargo.toml",
			readFile: func(string) ([]byte, error) { return nil, fmt.Errorf("file not found") },
		}
		err := c.SetVersion("1.3.0")
		assert.Error(t, err)
	})

	t.Run("write error propagated", func(t *testing.T) {
		t.Parallel()
		c := &Cargo{
			path:     "Cargo.toml",
			readFile: func(string) ([]byte, error) { return []byte(testCargoToml), nil },
			writeFile: func(_ string, _ []byte, _ os.FileMode) error {
				return fmt.Errorf("disk full")
			},
		}
		err := c.SetVersion("1.3.0")
		assert.EqualError(t, err, "failed to write file 'Cargo.toml': disk full")
	})
}

func TestCargoGetCoordinates(t *testing.T) {
	t.Parallel()

	t.Run("coordinates from Cargo.toml", func(t *testing.T) {
		t.Parallel()
		c := &Cargo{
			path:     "Cargo.toml",
			readFile: func(string) ([]byte, error) { return []byte(testCargoToml), nil },
		}
		coords, err := c.GetCoordinates()
		require.NoError(t, err)
		assert.Equal(t, "my-rust-app", coords.ArtifactID)
		assert.Equal(t, "", coords.GroupID)
		assert.Equal(t, "1.2.3", coords.Version)
	})

	t.Run("error when Cargo.toml missing", func(t *testing.T) {
		t.Parallel()
		c := &Cargo{
			path:     "Cargo.toml",
			readFile: func(string) ([]byte, error) { return nil, fmt.Errorf("file not found") },
		}
		_, err := c.GetCoordinates()
		assert.Error(t, err)
	})
}

func TestCargoVersioningScheme(t *testing.T) {
	t.Parallel()
	c := &Cargo{}
	assert.Equal(t, "semver2", c.VersioningScheme())
}
