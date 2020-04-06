package versioning

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNpmInit(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		npm := Npm{}
		npm.init()
		assert.Equal(t, "package.json", npm.PackageJSONPath)
	})

	t.Run("no default", func(t *testing.T) {
		npm := Npm{PackageJSONPath: "my/package.json"}
		npm.init()
		assert.Equal(t, "my/package.json", npm.PackageJSONPath)
	})
}

func TestNpmVersioningScheme(t *testing.T) {
	npm := Npm{}
	assert.Equal(t, "semver2", npm.VersioningScheme())
}

func TestNpmGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		npm := Npm{
			PackageJSONPath: "my/package.json",
			ReadFile:        func(filename string) ([]byte, error) { return []byte(`{"name": "test","version": "1.2.3"}`), nil },
		}
		version, err := npm.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case", func(t *testing.T) {
		npm := Npm{
			PackageJSONPath: "my/package.json",
			ReadFile:        func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
		}
		_, err := npm.GetVersion()
		assert.EqualError(t, err, "failed to read file 'my/package.json': read error")
	})
}

func TestNpmSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		var content []byte
		npm := Npm{
			PackageJSONPath: "my/package.json",
			ReadFile:        func(filename string) ([]byte, error) { return []byte(`{"name": "test","version": "1.2.3"}`), nil },
			WriteFile:       func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err := npm.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Contains(t, string(content), "1.2.4")
	})

	t.Run("error case", func(t *testing.T) {
		npm := Npm{
			PackageJSONPath: "my/package.json",
			ReadFile:        func(filename string) ([]byte, error) { return []byte(`{"name": "test","version": "1.2.3"}`), nil },
			WriteFile:       func(filename string, filecontent []byte, mode os.FileMode) error { return fmt.Errorf("write error") },
		}
		err := npm.SetVersion("1.2.4")
		assert.EqualError(t, err, "failed to write file 'my/package.json': write error")
	})
}
