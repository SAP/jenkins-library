//go:build unit

package versioning

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestPipGetVersion(t *testing.T) {
	t.Parallel()

	t.Run("success case - setup.py", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile("setup.py", []byte(`setup(name="simple-python",version="1.2.3"`))

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		version, err := pip.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("success case - setup.py & version.txt", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile("setup.py", []byte(`setup(name="simple-python",`))
		fileUtils.AddFile("version.txt", []byte(`1.2.4`))

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		version, err := pip.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.4", version)
	})

	t.Run("success case - setup.py & VERSION", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile("setup.py", []byte(`setup(name="simple-python",`))
		fileUtils.AddFile("VERSION", []byte(`1.2.5`))

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		version, err := pip.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.5", version)
	})

	t.Run("error to read file", func(t *testing.T) {
		fileUtils := mock.FilesMock{}

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		_, err := pip.GetVersion()
		assert.Contains(t, fmt.Sprint(err), "failed to read file 'setup.py'")
	})

	t.Run("error to retrieve version", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile("setup.py", []byte(`setup(name="simple-python",`))

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		_, err := pip.GetVersion()
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve version")
	})
}

func TestPipSetVersion(t *testing.T) {
	t.Run("success case - setup.py", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile("setup.py", []byte(`setup(name="simple-python",version="1.2.3"`))

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		err := pip.SetVersion("2.0.0")
		assert.NoError(t, err)
		content, _ := fileUtils.FileRead("setup.py")
		assert.Contains(t, string(content), `version="2.0.0"`)
	})

	t.Run("success case - setup.py & version.txt", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile("setup.py", []byte(`setup(name="simple-python",`))
		fileUtils.AddFile("version.txt", []byte(`1.2.3`))

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		err := pip.SetVersion("2.0.0")
		assert.NoError(t, err)
		content, _ := fileUtils.FileRead("version.txt")
		assert.Equal(t, "2.0.0", string(content))
	})

	t.Run("success case - setup.py & VERSION", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile("setup.py", []byte(`setup(name="simple-python",`))
		fileUtils.AddFile("VERSION", []byte(`1.2.3`))

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		err := pip.SetVersion("2.0.0")
		assert.NoError(t, err)
		content, _ := fileUtils.FileRead("VERSION")
		assert.Equal(t, "2.0.0", string(content))
	})

	t.Run("error to read file", func(t *testing.T) {
		fileUtils := mock.FilesMock{}

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		err := pip.SetVersion("2.0.0")
		assert.Contains(t, fmt.Sprint(err), "failed to read file 'setup.py'")
	})

	t.Run("error to retrieve version", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile("setup.py", []byte(`setup(name="simple-python",`))

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		err := pip.SetVersion("2.0.0")
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve version")
	})
}

func TestPipGetCoordinates(t *testing.T) {
	t.Run("success case - setup.py", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile("setup.py", []byte(`setup(name="simple-python",version="1.2.3"`))

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		coordinates, err := pip.GetCoordinates()
		assert.NoError(t, err)
		assert.Equal(t, "simple-python", coordinates.ArtifactID)
		assert.Equal(t, "1.2.3", coordinates.Version)

	})

	t.Run("success case - only version", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile("setup.py", []byte(`setup(version="1.2.3"`))

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		coordinates, err := pip.GetCoordinates()
		assert.NoError(t, err)
		assert.Equal(t, "", coordinates.ArtifactID)
		assert.Equal(t, "1.2.3", coordinates.Version)

	})

	t.Run("error to retrieve setup.py", func(t *testing.T) {
		fileUtils := mock.FilesMock{}

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		_, err := pip.GetCoordinates()
		assert.Contains(t, fmt.Sprint(err), "failed to read file 'setup.py'")
	})

	t.Run("error to retrieve version", func(t *testing.T) {
		fileUtils := mock.FilesMock{}
		fileUtils.AddFile("setup.py", []byte(`setup(name="simple-python"`))

		pip := Pip{
			path:       "setup.py",
			fileExists: fileUtils.FileExists,
			readFile:   fileUtils.FileRead,
			writeFile:  fileUtils.FileWrite,
		}

		_, err := pip.GetCoordinates()
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve version")
	})
}
