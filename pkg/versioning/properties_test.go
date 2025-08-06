package versioning

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/magiconair/properties"
	"github.com/stretchr/testify/assert"
)

func TestPropertiesFileGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		tmpFolder := t.TempDir()

		propsFilePath := filepath.Join(tmpFolder, "my.props")
		os.WriteFile(propsFilePath, []byte("version = 1.2.3"), 0666)

		propsfile := PropertiesFile{
			path: propsFilePath,
		}
		version, err := propsfile.GetVersion()

		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("success case - custom version field", func(t *testing.T) {
		tmpFolder := t.TempDir()

		propsFilePath := filepath.Join(tmpFolder, "my.props")
		os.WriteFile(propsFilePath, []byte("customversion = 1.2.3"), 0666)

		propsfile := PropertiesFile{
			path:         propsFilePath,
			versionField: "customversion",
		}
		version, err := propsfile.GetVersion()

		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case - file not found", func(t *testing.T) {
		tmpFolder := t.TempDir()

		propsFilePath := filepath.Join(tmpFolder, "my.props")

		propsfile := PropertiesFile{
			path: propsFilePath,
		}
		_, err := propsfile.GetVersion()

		assert.Contains(t, fmt.Sprint(err), "failed to load")
	})

	t.Run("error case - no version found", func(t *testing.T) {
		tmpFolder := t.TempDir()

		propsFilePath := filepath.Join(tmpFolder, "my.props")
		os.WriteFile(propsFilePath, []byte("versionx = 1.2.3"), 0666)

		propsfile := PropertiesFile{
			path: propsFilePath,
		}
		_, err := propsfile.GetVersion()

		assert.EqualError(t, err, "no version found in field version")
	})
}

func TestPropertiesFileSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		tmpFolder := t.TempDir()

		propsFilePath := filepath.Join(tmpFolder, "my.props")
		os.WriteFile(propsFilePath, []byte("version = 0.0.1"), 0666)

		var content []byte
		propsfile := PropertiesFile{
			path:      propsFilePath,
			writeFile: func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err := propsfile.SetVersion("1.2.3")
		assert.NoError(t, err)

		assert.Contains(t, string(content), "version = 1.2.3")
	})

	t.Run("error case - write failed", func(t *testing.T) {
		props := properties.LoadMap(map[string]string{"version": "0.0.1"})
		propsfile := PropertiesFile{
			content:      props,
			path:         "gradle.properties",
			versionField: "version",
			writeFile:    func(filename string, filecontent []byte, mode os.FileMode) error { return fmt.Errorf("write error") },
		}
		err := propsfile.SetVersion("1.2.3")
		assert.EqualError(t, err, "failed to write file: write error")
	})
}
