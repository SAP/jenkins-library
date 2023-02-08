package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type scriptCollectionDownloadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newScriptCollectionDownloadTestsUtils() scriptCollectionDownloadMockUtils {
	utils := scriptCollectionDownloadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunScriptCollectionDownload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := scriptCollectionDownloadOptions{}

		utils := newScriptCollectionDownloadTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runScriptCollectionDownload(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := scriptCollectionDownloadOptions{}

		utils := newScriptCollectionDownloadTestsUtils()

		// test
		err := runScriptCollectionDownload(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
