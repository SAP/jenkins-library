package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type valueMappingArtifactDownloadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newValueMappingArtifactDownloadTestsUtils() valueMappingArtifactDownloadMockUtils {
	utils := valueMappingArtifactDownloadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunValueMappingArtifactDownload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := valueMappingArtifactDownloadOptions{}

		utils := newValueMappingArtifactDownloadTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runValueMappingArtifactDownload(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := valueMappingArtifactDownloadOptions{}

		utils := newValueMappingArtifactDownloadTestsUtils()

		// test
		err := runValueMappingArtifactDownload(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
