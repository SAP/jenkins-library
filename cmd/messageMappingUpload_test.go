package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type messageMappingUploadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newMessageMappingUploadTestsUtils() messageMappingUploadMockUtils {
	utils := messageMappingUploadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunMessageMappingUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := messageMappingUploadOptions{}

		utils := newMessageMappingUploadTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runMessageMappingUpload(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := messageMappingUploadOptions{}

		utils := newMessageMappingUploadTestsUtils()

		// test
		err := runMessageMappingUpload(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
