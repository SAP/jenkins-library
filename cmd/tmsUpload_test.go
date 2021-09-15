package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type tmsUploadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTmsUploadTestsUtils() tmsUploadMockUtils {
	utils := tmsUploadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunTmsUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := tmsUploadOptions{}

		utils := newTmsUploadTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runTmsUpload(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := tmsUploadOptions{}

		utils := newTmsUploadTestsUtils()

		// test
		err := runTmsUpload(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
