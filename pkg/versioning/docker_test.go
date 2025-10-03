package versioning

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerGetVersion(t *testing.T) {
	t.Run("success case - FROM", func(t *testing.T) {
		docker := Docker{
			readFile:      func(filename string) ([]byte, error) { return []byte("FROM test:1.2.3"), nil },
			versionSource: "FROM",
		}
		version, err := docker.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case - FROM failed", func(t *testing.T) {
		docker := Docker{
			readFile:      func(filename string) ([]byte, error) { return []byte("FROM test"), nil },
			versionSource: "FROM",
		}
		_, err := docker.GetVersion()
		assert.EqualError(t, err, "no version information available in FROM statement")
	})

	t.Run("error case - FROM read error", func(t *testing.T) {
		docker := Docker{
			readFile:      func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
			versionSource: "FROM",
		}
		_, err := docker.GetVersion()
		assert.EqualError(t, err, "failed to read file 'Dockerfile': read error")
	})

	t.Run("success case - buildTool", func(t *testing.T) {

		dir := t.TempDir()
		filePath := filepath.Join(dir, "package.json")
		err := os.WriteFile(filePath, []byte(`{"version": "1.2.3"}`), 0700)
		if err != nil {
			t.Fatal("Failed to create test file")
		}
		docker := Docker{
			path:          filePath,
			versionSource: "npm",
		}
		version, err := docker.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("success case - ENV", func(t *testing.T) {
		docker := Docker{
			readFile:      func(filename string) ([]byte, error) { return []byte("FROM test:latest\n\nENV VERSION_ENV 1.2.3"), nil },
			versionSource: "VERSION_ENV",
		}
		version, err := docker.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case - ENV failed", func(t *testing.T) {
		docker := Docker{
			readFile:      func(filename string) ([]byte, error) { return []byte("FROM test:latest\n\nENV VERSION_ENV 1.2.3"), nil },
			versionSource: "NOT_FOUND",
		}
		_, err := docker.GetVersion()
		assert.EqualError(t, err, "no version information available in ENV 'NOT_FOUND'")
	})

	t.Run("error case - ENV read error", func(t *testing.T) {
		docker := Docker{
			readFile:      func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
			versionSource: "VERSION_ENV",
		}
		_, err := docker.GetVersion()
		assert.EqualError(t, err, "failed to read file 'Dockerfile': read error")
	})

	t.Run("error case - fallback", func(t *testing.T) {
		docker := Docker{
			readFile:      func(filename string) ([]byte, error) { return []byte{}, fmt.Errorf("read error") },
			versionSource: "",
		}
		_, err := docker.GetVersion()
		assert.Contains(t, fmt.Sprint(err), "failed to read file 'VERSION': open VERSION")
	})
}

func TestDockerSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		var content []byte
		docker := Docker{
			readFile:      func(filename string) ([]byte, error) { return []byte("FROM test:1.2.3"), nil },
			writeFile:     func(filename string, filecontent []byte, mode os.FileMode) error { content = filecontent; return nil },
			versionSource: "FROM",
		}
		err := docker.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Contains(t, string(content), "1.2.4")
	})

	t.Run("error case", func(t *testing.T) {
		docker := Docker{
			readFile:      func(filename string) ([]byte, error) { return []byte("FROM test:1.2.3"), nil },
			writeFile:     func(filename string, filecontent []byte, mode os.FileMode) error { return fmt.Errorf("write error") },
			versionSource: "FROM",
		}
		err := docker.SetVersion("1.2.4")
		assert.EqualError(t, err, "failed to write file 'VERSION': write error")
	})

	t.Run("success case - buildTool", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "package.json")
		err := os.WriteFile(filePath, []byte(`{"version": "1.2.3"}`), 0700)
		if err != nil {
			t.Fatal("Failed to create test file")
		}
		docker := Docker{
			path:          filePath,
			versionSource: "npm",
		}
		_, err = docker.GetVersion()
		assert.NoError(t, err)
		err = docker.SetVersion("1.2.4")
		assert.NoError(t, err)
		packageJSON, err := os.ReadFile(filePath)
		assert.Contains(t, string(packageJSON), `"version": "1.2.4"`)
		versionContent, err := os.ReadFile(filepath.Join(dir, "VERSION"))
		assert.Equal(t, "1.2.4", string(versionContent))
	})
}

func TestVersionFromBaseImageTag(t *testing.T) {
	tt := []struct {
		docker   *Docker
		expected string
	}{
		{docker: &Docker{content: []byte("")}, expected: ""},
		{docker: &Docker{content: []byte("FROM test")}, expected: ""},
		//{docker: &Docker{content: []byte("FROM test:latest")}, expected: ""},
		{docker: &Docker{content: []byte("FROM test:1.2.3")}, expected: "1.2.3"},
		{docker: &Docker{content: []byte("#COMMENT\nFROM test:1.2.3")}, expected: "1.2.3"},
		//{docker: &Docker{content: []byte("FROM my.registry:55555/test")}, expected: ""},
		//{docker: &Docker{content: []byte("FROM my.registry:55555/test:latest")}, expected: ""},
		{docker: &Docker{content: []byte("FROM my.registry:55555/test:1.2.3")}, expected: "1.2.3"},
	}
	for _, test := range tt {
		assert.Equal(t, test.expected, test.docker.versionFromBaseImageTag())
	}
}

func TestGetCoordinates(t *testing.T) {
	docker := Docker{
		readFile:      func(filename string) ([]byte, error) { return []byte("FROM test:1.2.3"), nil },
		versionSource: "FROM",
		options:       &Options{DockerImage: "my/test/image:tag"},
	}

	coordinates, err := docker.GetCoordinates()
	assert.NoError(t, err)
	assert.Equal(t, Coordinates{GroupID: "", ArtifactID: "my_test_image_tag", Version: ""}, coordinates)
}
