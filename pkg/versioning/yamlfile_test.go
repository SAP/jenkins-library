//go:build unit

package versioning

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYAMLfileGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		yamlfile := YAMLfile{
			path:     "my.yaml",
			readFile: func(filename string) ([]byte, error) { return []byte(`version: 1.2.3`), nil },
		}
		version, err := yamlfile.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case", func(t *testing.T) {
		yamlfile := YAMLfile{
			path:         "my.yaml",
			versionField: "theversion",
			readFile:     func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
		}
		version, err := yamlfile.GetVersion()
		assert.EqualError(t, err, "failed to get key theversion: failed to read file 'my.yaml': read error")
		assert.Equal(t, "", version)
	})
}

func TestYAMLfileGetArtifactID(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		yamlfile := YAMLfile{
			path:     "my.yaml",
			readFile: func(filename string) ([]byte, error) { return []byte(`ID: artifact-id`), nil },
		}
		artifactID, err := yamlfile.GetArtifactID()
		assert.NoError(t, err)
		assert.Equal(t, "artifact-id", artifactID)
	})

	t.Run("error case", func(t *testing.T) {
		yamlfile := YAMLfile{
			path:            "my.yaml",
			artifactIDField: "theArtifact",
			readFile:        func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
		}
		artifactID, err := yamlfile.GetArtifactID()
		assert.EqualError(t, err, "failed to get key theArtifact: failed to read file 'my.yaml': read error")
		assert.Equal(t, "", artifactID)
	})
}

func TestYAMLfileSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		var content []byte
		yamlfile := YAMLfile{
			path:         "my.yaml",
			versionField: "theversion",
			readFile:     func(filename string) ([]byte, error) { return []byte(`theversion: 1.2.3`), nil },
			writeFile:    func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err := yamlfile.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Contains(t, string(content), "theversion: 1.2.4")
	})

	t.Run("error case", func(t *testing.T) {
		yamlfile := YAMLfile{
			path:         "my.yaml",
			versionField: "theversion",
			readFile:     func(filename string) ([]byte, error) { return []byte(`theversion: 1.2.3`), nil },
			writeFile:    func(filename string, filecontent []byte, mode os.FileMode) error { return fmt.Errorf("write error") },
		}
		err := yamlfile.SetVersion("1.2.4")
		assert.EqualError(t, err, "failed to write file 'my.yaml': write error")
	})
}
