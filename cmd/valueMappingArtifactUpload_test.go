package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type valueMappingArtifactUploadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newValueMappingArtifactUploadTestsUtils() valueMappingArtifactUploadMockUtils {
	utils := valueMappingArtifactUploadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunValueMappingArtifactUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := valueMappingArtifactUploadOptions{}

		utils := newValueMappingArtifactUploadTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runValueMappingArtifactUpload(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := valueMappingArtifactUploadOptions{}

		utils := newValueMappingArtifactUploadTestsUtils()

		// test
		err := runValueMappingArtifactUpload(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
