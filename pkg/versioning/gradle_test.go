package versioning

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGradleGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		tmpFolder, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("failed to create temp dir")
		}
		defer os.RemoveAll(tmpFolder)

		gradlePropsFilePath := filepath.Join(tmpFolder, "gradle.properties")
		ioutil.WriteFile(gradlePropsFilePath, []byte("version = 1.2.3"), 0666)
		gradle := &Gradle{
			path: gradlePropsFilePath,
		}

		version, err := gradle.GetVersion()

		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})
}

func TestGradleSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		tmpFolder, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("failed to create temp dir")
		}
		defer os.RemoveAll(tmpFolder)

		gradlePropsFilePath := filepath.Join(tmpFolder, "gradle.properties")
		ioutil.WriteFile(gradlePropsFilePath, []byte("version = 0.0.1"), 0666)

		var content []byte
		gradle := &Gradle{
			path:      gradlePropsFilePath,
			writeFile: func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
		}
		err = gradle.SetVersion("1.2.3")
		assert.NoError(t, err)

		assert.Contains(t, string(content), "version = 1.2.3")
	})
}
