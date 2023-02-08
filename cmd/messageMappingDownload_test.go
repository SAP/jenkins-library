package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type messageMappingDownloadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newMessageMappingDownloadTestsUtils() messageMappingDownloadMockUtils {
	utils := messageMappingDownloadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunMessageMappingDownload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := messageMappingDownloadOptions{}

		utils := newMessageMappingDownloadTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runMessageMappingDownload(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := messageMappingDownloadOptions{}

		utils := newMessageMappingDownloadTestsUtils()

		// test
		err := runMessageMappingDownload(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
