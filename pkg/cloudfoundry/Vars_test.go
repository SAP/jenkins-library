package cloudfoundry

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
	"testing"
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
		opts, err := GetVarFileOptions([]string{"varsA.yml", "varsB.yml"})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"--vars-file", "varsA.yml", "--vars-file", "varsB.yml"}, opts)
		}
	})

	t.Run("Some vars files missing", func(t *testing.T) {
		opts, err := GetVarFileOptions([]string{"varsA.yml", "varsC.yml", "varsD.yml"})
		if assert.EqualError(t, err, "Some vars files could not be found: [varsC.yml varsD.yml]") {
			assert.IsType(t, &VarsFilesNotFoundError{}, err)
			assert.Equal(t, []string{"--vars-file", "varsA.yml"}, opts)
		}
	})
}
