package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type integrationPackageDownloadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newIntegrationPackageDownloadTestsUtils() integrationPackageDownloadMockUtils {
	utils := integrationPackageDownloadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunIntegrationPackageDownload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := integrationPackageDownloadOptions{}

		utils := newIntegrationPackageDownloadTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runIntegrationPackageDownload(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := integrationPackageDownloadOptions{}

		utils := newIntegrationPackageDownloadTestsUtils()

		// test
		err := runIntegrationPackageDownload(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
