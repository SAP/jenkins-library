//go:build unit
// +build unit

package cloudfoundry

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
)

func TestVarsFiles(t *testing.T) {

	defer func() {
		_fileUtils = piperutils.Files{}
	}()

	filesMock := mock.FilesMock{}
	filesMock.AddDir("/home/me")
	filesMock.Chdir("/home/me")
	filesMock.AddFile("varsA.yml", []byte("file content does not matter"))
	filesMock.AddFile("varsB.yml", []byte("file content does not matter"))
	_fileUtils = &filesMock

	t.Run("All vars files found", func(t *testing.T) {
		opts, err := GetVarsFileOptions([]string{"varsA.yml", "varsB.yml"})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"--vars-file", "varsA.yml", "--vars-file", "varsB.yml"}, opts)
		}
	})

	t.Run("Some vars files missing", func(t *testing.T) {
		opts, err := GetVarsFileOptions([]string{"varsA.yml", "varsC.yml", "varsD.yml"})
		if assert.EqualError(t, err, "Some vars files could not be found: [varsC.yml varsD.yml]") {
			assert.IsType(t, &VarsFilesNotFoundError{}, err)
			assert.Equal(t, []string{"--vars-file", "varsA.yml"}, opts)
		}
	})
}

func TestVars(t *testing.T) {

	t.Run("Empty vars", func(t *testing.T) {
		opts, err := GetVarsOptions([]string{})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{}, opts)
		}
	})

	t.Run("Some vars", func(t *testing.T) {
		opts, err := GetVarsOptions([]string{"a=b", "x=y"})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"--var", "a=b", "--var", "x=y"}, opts)
		}
	})
}
