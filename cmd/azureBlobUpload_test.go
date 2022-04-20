package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type azureBlobUploadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newAzureBlobUploadTestsUtils() azureBlobUploadMockUtils {
	utils := azureBlobUploadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunAzureBlobUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := azureBlobUploadOptions{}

		utils := newAzureBlobUploadTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runAzureBlobUpload(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := azureBlobUploadOptions{}

		utils := newAzureBlobUploadTestsUtils()

		// test
		err := runAzureBlobUpload(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
