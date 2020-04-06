package versioning

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipVersioningScheme(t *testing.T) {
	pip := Pip{}
	assert.Equal(t, "pep440", pip.VersioningScheme())
}

func TestPipGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		pip := Pip{
			VersionPath: "my/version.txt",
			ReadFile:    func(filename string) ([]byte, error) { return []byte("1.2.3"), nil },
		}
		version, err := pip.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case", func(t *testing.T) {
		pip := Pip{
			VersionPath: "my/version.txt",
			ReadFile:    func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
		}
		_, err := pip.GetVersion()
		assert.EqualError(t, err, "failed to read file 'my/version.txt': read error")
	})
}

func TestPipSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		var content []byte
		pip := Pip{
			VersionPath: "my/version.txt",
			ReadFile:    func(filename string) ([]byte, error) { return []byte("1.2.3"), nil },
			WriteFile:   func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err := pip.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Contains(t, string(content), "1.2.4")
	})

	t.Run("error case", func(t *testing.T) {
		pip := Pip{
			VersionPath: "my/version.txt",
			ReadFile:    func(filename string) ([]byte, error) { return []byte("1.2.3"), nil },
			WriteFile:   func(filename string, filecontent []byte, mode os.FileMode) error { return fmt.Errorf("write error") },
		}
		err := pip.SetVersion("1.2.4")
		assert.EqualError(t, err, "failed to write file 'my/version.txt': write error")
	})
}
