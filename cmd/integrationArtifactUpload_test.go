//go:build unit
// +build unit

package cmd

import (
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunIntegrationArtifactUpload(t *testing.T) {
	t.Parallel()

	t.Run("Successfull Integration Flow Create Test", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		path := filepath.Join("tempDir", "iflow4.zip")
		filesMock.AddFile(path, []byte("dummy content"))
		exists, err := filesMock.FileExists(path)
		assert.NoError(t, err)
		assert.True(t, exists)

		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := integrationArtifactUploadOptions{
			APIServiceKey:       apiServiceKey,
			IntegrationFlowName: "flow4",
			IntegrationFlowID:   "flow4",
			PackageID:           "CICD",
			FilePath:            path,
		}

		httpClient := httpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "PositiveAndCreateIntegrationDesigntimeArtifactResBody"}

		err = runIntegrationArtifactUpload(&config, nil, &filesMock, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/IntegrationDesigntimeArtifacts(Id='flow4',Version='Active')", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "PUT", httpClient.Method)
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
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactUploadOptions{
			APIServiceKey:       apiServiceKey,
			IntegrationFlowName: "flow4",
			IntegrationFlowID:   "flow4",
			PackageID:           "CICD",
			FilePath:            path,
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

		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactUploadOptions{
			APIServiceKey:       apiServiceKey,
			IntegrationFlowName: "flow4",
			IntegrationFlowID:   "flow4",
			PackageID:           "CICD",
			FilePath:            "path",
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

		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactUploadOptions{
			APIServiceKey:       apiServiceKey,
			IntegrationFlowName: "flow4",
			IntegrationFlowID:   "flow4",
			PackageID:           "CICD",
			FilePath:            path,
		}

		httpClient := httpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "NegativeAndCreateIntegrationDesigntimeArtifactResBody"}

		err = runIntegrationArtifactUpload(&config, nil, &files, &httpClient)
		assert.EqualError(t, err, "HTTP PUT request to https://demo/api/v1/IntegrationDesigntimeArtifacts(Id='flow4',Version='Active') failed with error: : 401 Unauthorized")
	})

	t.Run("Failed case of Integration Flow Create Test", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		path := filepath.Join("tempDir", "iflow4.zip")
		filesMock.AddFile(path, []byte("dummy content"))
		exists, err := filesMock.FileExists(path)
		assert.NoError(t, err)
		assert.True(t, exists)

		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := integrationArtifactUploadOptions{
			APIServiceKey:       apiServiceKey,
			IntegrationFlowName: "flow4",
			IntegrationFlowID:   "flow4",
			PackageID:           "CICD",
			FilePath:            path,
		}

		httpClient := httpMockCpis{CPIFunction: "", ResponseBody: ``, TestType: "NegativeAndUpdateIntegrationDesigntimeArtifactResBody"}

		err = runIntegrationArtifactUpload(&config, nil, &filesMock, &httpClient)
		assert.EqualError(t, err, "HTTP POST request to https://demo/api/v1/IntegrationDesigntimeArtifacts failed with error: : 401 Unauthorized")
	})
}
