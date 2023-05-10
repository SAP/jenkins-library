//go:build unit
// +build unit

package versioning

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionfileInit(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		versionfile := Versionfile{}
		versionfile.init()
		assert.Equal(t, "VERSION", versionfile.path)
	})

	t.Run("no default", func(t *testing.T) {
		versionfile := Versionfile{path: "my/VERSION"}
		versionfile.init()
		assert.Equal(t, "my/VERSION", versionfile.path)
	})
}

func TestVersionfileVersioningScheme(t *testing.T) {
	versionfile := Versionfile{}
	assert.Equal(t, "semver2", versionfile.VersioningScheme())
}

func TestVersionfileGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		versionfile := Versionfile{
			path:     "my/VERSION",
			readFile: func(filename string) ([]byte, error) { return []byte("1.2.3"), nil },
		}
		version, err := versionfile.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("success case - trimming", func(t *testing.T) {
		versionfile := Versionfile{
			path:     "my/VERSION",
			readFile: func(filename string) ([]byte, error) { return []byte("1.2.3 \n"), nil },
		}
		version, err := versionfile.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case", func(t *testing.T) {
		versionfile := Versionfile{
			path:     "my/VERSION",
			readFile: func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
		}
		_, err := versionfile.GetVersion()
		assert.EqualError(t, err, "failed to read file 'my/VERSION': read error")
	})
}

func TestVersionfileSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		var content []byte
		versionfile := Versionfile{
			path:      "my/VERSION",
			readFile:  func(filename string) ([]byte, error) { return []byte("1.2.3"), nil },
			writeFile: func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err := versionfile.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Contains(t, string(content), "1.2.4")
	})

	t.Run("error case", func(t *testing.T) {
		versionfile := Versionfile{
			path:      "my/VERSION",
			readFile:  func(filename string) ([]byte, error) { return []byte("1.2.3"), nil },
			writeFile: func(filename string, filecontent []byte, mode os.FileMode) error { return fmt.Errorf("write error") },
		}
		err := versionfile.SetVersion("1.2.4")
		assert.EqualError(t, err, "failed to write file 'my/VERSION': write error")
	})
}
