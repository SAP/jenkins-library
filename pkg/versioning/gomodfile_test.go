package versioning

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestGolangGetVersion(t *testing.T) {
	t.Run("success case - go.mod", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("go.mod", []byte("module github.com/SAP/jenkins-library"))
		gomod := GoMod{
			path:       "go.mod",
			readFile:   filesMock.FileRead,
			fileExists: filesMock.FileExists,
		}
		version, err := gomod.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "", version)
	})

	t.Run("success case - go.mod & VERSION", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("go.mod", []byte("module github.com/SAP/jenkins-library"))
		filesMock.AddFile("VERSION", []byte("1.2.3"))
		gomod := GoMod{
			path:       "go.mod",
			readFile:   filesMock.FileRead,
			fileExists: filesMock.FileExists,
		}
		version, err := gomod.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("success case - VERSION", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("VERSION", []byte("1.2.3"))
		gomod := GoMod{
			readFile:   filesMock.FileRead,
			fileExists: filesMock.FileExists,
		}
		version, err := gomod.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case - read file", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("go.mod", []byte("module github.com/SAP/jenkins-library"))
		filesMock.FileReadErrors = map[string]error{"go.mod": fmt.Errorf("cannot open")}
		gomod := GoMod{
			path:       "go.mod",
			readFile:   filesMock.FileRead,
			fileExists: filesMock.FileExists,
		}
		_, err := gomod.GetVersion()
		assert.EqualError(t, err, "failed to read file 'go.mod': cannot open")
	})
}

func TestGolangSetVersion(t *testing.T) {
	t.Run("success case - go.mod", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("go.mod", []byte("module github.com/SAP/jenkins-library"))
		filesMock.AddFile("VERSION", []byte("1.2.3"))
		gomod := GoMod{
			path:       "go.mod",
			readFile:   filesMock.FileRead,
			fileExists: filesMock.FileExists,
			writeFile:  filesMock.FileWrite,
		}
		err := gomod.SetVersion("1.2.4")
		assert.NoError(t, err)
		content, err := filesMock.FileRead("VERSION")
		assert.NoError(t, err)
		assert.Equal(t, "1.2.4", string(content))
	})

	t.Run("success case - no go.mod", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("VERSION", []byte("1.2.3"))
		gomod := GoMod{
			path:       "VERSION",
			readFile:   filesMock.FileRead,
			fileExists: filesMock.FileExists,
			writeFile:  filesMock.FileWrite,
		}
		err := gomod.SetVersion("1.2.4")
		assert.NoError(t, err)
		content, err := filesMock.FileRead("VERSION")
		assert.NoError(t, err)
		assert.Equal(t, "1.2.4", string(content))
	})

	t.Run("error case - no version file", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("go.mod", []byte("module github.com/SAP/jenkins-library"))
		gomod := GoMod{
			path:       "go.mod",
			readFile:   filesMock.FileRead,
			fileExists: filesMock.FileExists,
			writeFile:  filesMock.FileWrite,
		}
		err := gomod.SetVersion("1.2.4")
		assert.EqualError(t, err, "no version.txt/VERSION file available but required: no build descriptor available, supported: [VERSION version.txt]")
	})
}

func TestGolangGetCoordinates(t *testing.T) {
	t.Run("success case - go.mod & VERSION", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("go.mod", []byte("module github.com/SAP/jenkins-library"))
		filesMock.AddFile("VERSION", []byte("1.2.3"))
		gomod := GoMod{
			path:       "go.mod",
			readFile:   filesMock.FileRead,
			fileExists: filesMock.FileExists,
		}
		coordinates, err := gomod.GetCoordinates()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", coordinates.Version)
		assert.Equal(t, "jenkins-library", coordinates.ArtifactID)
		assert.Equal(t, "github.com/SAP", coordinates.GroupID)
	})

	t.Run("success case - unspecified", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("go.mod", []byte("module github.com/SAP/jenkins-library"))
		gomod := GoMod{
			path:       "go.mod",
			readFile:   filesMock.FileRead,
			fileExists: filesMock.FileExists,
		}
		coordinates, err := gomod.GetCoordinates()
		assert.NoError(t, err)
		assert.Equal(t, "unspecified", coordinates.Version)
		assert.Equal(t, "jenkins-library", coordinates.ArtifactID)
		assert.Equal(t, "github.com/SAP", coordinates.GroupID)
	})

	t.Run("success case - VERSION only", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("VERSION", []byte("1.2.3"))
		gomod := GoMod{
			path:       "VERSION",
			readFile:   filesMock.FileRead,
			fileExists: filesMock.FileExists,
		}
		coordinates, err := gomod.GetCoordinates()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", coordinates.Version)
		assert.Equal(t, "", coordinates.ArtifactID)
		assert.Equal(t, "", coordinates.GroupID)
	})

	t.Run("error case - invalid go.mod", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("go.mod", []byte("molule"))
		gomod := GoMod{
			path:       "go.mod",
			readFile:   filesMock.FileRead,
			fileExists: filesMock.FileExists,
		}
		_, err := gomod.GetCoordinates()
		assert.Contains(t, fmt.Sprint(err), "failed to parse go.mod file")
	})
}
