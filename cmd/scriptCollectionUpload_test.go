package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type scriptCollectionUploadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newScriptCollectionUploadTestsUtils() scriptCollectionUploadMockUtils {
	utils := scriptCollectionUploadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunScriptCollectionUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := scriptCollectionUploadOptions{}

		utils := newScriptCollectionUploadTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runScriptCollectionUpload(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := scriptCollectionUploadOptions{}

		utils := newScriptCollectionUploadTestsUtils()

		// test
		err := runScriptCollectionUpload(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
