package cmd

import (
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type integrationArtifactUploadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newIntegrationArtifactUploadTestsUtils() integrationArtifactUploadMockUtils {
	utils := integrationArtifactUploadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunIntegrationArtifactUpload(t *testing.T) {
	t.Parallel()

	t.Run("Successfull Integration Flow Create Test", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		path := filepath.Join("tempDir", "iflow4.zip")
		filesMock.AddFile(path, []byte("dummy content"))
		exists, err := filesMock.FileExists(path)
		assert.NoError(t, err)
		assert.True(t, exists)

		config := integrationArtifactUploadOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowName:    "flow4",
			IntegrationFlowID:      "flow4",
			IntegrationFlowVersion: "1.0.4",
			PackageID:              "CICD",
			FilePath:               path,
		}

		httpClient := httpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "PositiveAndCreateIntegrationDesigntimeArtifactResBody"}

		err = runIntegrationArtifactUpload(&config, nil, &filesMock, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/IntegrationDesigntimeArtifactSaveAsVersion?Id='flow4'&SaveAsVersion='1.0.4'", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})
		}
	})

	t.Run("Successfull Integration Flow Update Test", func(t *testing.T) {

		files := mock.FilesMock{}
		path := filepath.Join("tempDir", "iflow4.zip")
		files.AddFile(path, []byte("dummy content"))
		exists, err := files.FileExists(path)
		assert.NoError(t, err)
		assert.True(t, exists)
		config := integrationArtifactUploadOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowName:    "flow4",
			IntegrationFlowID:      "flow4",
			IntegrationFlowVersion: "1.0.4",
			PackageID:              "CICD",
			FilePath:               path,
		}

		httpClient := httpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "PositiveAndUpdateIntegrationDesigntimeArtifactResBody"}

		err = runIntegrationArtifactUpload(&config, nil, &files, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/IntegrationDesigntimeArtifacts", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})
		}
	})

	t.Run("Failed case of Integration Flow Get Test", func(t *testing.T) {

		config := integrationArtifactUploadOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowName:    "flow4",
			IntegrationFlowID:      "flow4",
			IntegrationFlowVersion: "1.0.4",
			PackageID:              "CICD",
			FilePath:               "path",
		}

		httpClient := httpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "NegativeAndGetIntegrationDesigntimeArtifactResBody"}

		err := runIntegrationArtifactUpload(&config, nil, nil, &httpClient)
		assert.Error(t, err)
	})

	t.Run("Failed case of Integration Flow Update Test", func(t *testing.T) {
		files := mock.FilesMock{}
		path := filepath.Join("tempDir", "iflow4.zip")
		files.AddFile(path, []byte("dummy content"))
		exists, err := files.FileExists(path)
		assert.NoError(t, err)
		assert.True(t, exists)

		config := integrationArtifactUploadOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowName:    "flow4",
			IntegrationFlowID:      "flow4",
			IntegrationFlowVersion: "1.0.4",
			PackageID:              "CICD",
			FilePath:               path,
		}

		httpClient := httpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "NegativeAndCreateIntegrationDesigntimeArtifactResBody"}

		err = runIntegrationArtifactUpload(&config, nil, &files, &httpClient)
		assert.EqualError(t, err, "HTTP POST request to https://demo/api/v1/IntegrationDesigntimeArtifactSaveAsVersion?Id='flow4'&SaveAsVersion='1.0.4' failed with error: : Internal error")
	})

	t.Run("Failed case of Integration Flow Create Test", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		path := filepath.Join("tempDir", "iflow4.zip")
		filesMock.AddFile(path, []byte("dummy content"))
		exists, err := filesMock.FileExists(path)
		assert.NoError(t, err)
		assert.True(t, exists)

		config := integrationArtifactUploadOptions{
			Host:                   "https://demo",
			OAuthTokenProviderURL:  "https://demo/oauth/token",
			Username:               "demouser",
			Password:               "******",
			IntegrationFlowName:    "flow4",
			IntegrationFlowID:      "flow4",
			IntegrationFlowVersion: "1.0.4",
			PackageID:              "CICD",
			FilePath:               path,
		}

		httpClient := httpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "NegativeAndUpdateIntegrationDesigntimeArtifactResBody"}

		err = runIntegrationArtifactUpload(&config, nil, &filesMock, &httpClient)
		assert.EqualError(t, err, "HTTP POST request to https://demo/api/v1/IntegrationDesigntimeArtifacts failed with error: : Internal error")
	})
}
