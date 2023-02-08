package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type integrationPackageUploadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newIntegrationPackageUploadTestsUtils() integrationPackageUploadMockUtils {
	utils := integrationPackageUploadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunIntegrationPackageUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := integrationPackageUploadOptions{}

		utils := newIntegrationPackageUploadTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runIntegrationPackageUpload(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := integrationPackageUploadOptions{}

		utils := newIntegrationPackageUploadTestsUtils()

		// test
		err := runIntegrationPackageUpload(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
