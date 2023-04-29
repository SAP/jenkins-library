//go:build unit
// +build unit

package cmd

import (
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunIntegrationArtifactResource(t *testing.T) {
	t.Parallel()

	t.Run("Create Resource Test", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		path := filepath.Join("tempDir", "demo.xsl")
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
		config := integrationArtifactResourceOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
			Operation:         "create",
			ResourcePath:      path,
		}
		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactResourceCreate", ResponseBody: ``, TestType: "Positive"}

		// test
		err = runIntegrationArtifactResource(&config, nil, &filesMock, &httpClient)

		// assert
		assert.NoError(t, err)
	})

	t.Run("Update Resource Test", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		path := filepath.Join("tempDir", "demo.xsl")
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
		config := integrationArtifactResourceOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
			Operation:         "update",
			ResourcePath:      path,
		}
		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactResourceUpdate", ResponseBody: ``, TestType: "Positive"}

		// test
		err = runIntegrationArtifactResource(&config, nil, &filesMock, &httpClient)

		// assert
		assert.NoError(t, err)
	})

	t.Run("Delete Resource Test", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		path := filepath.Join("tempDir", "demo.xsl")
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
		config := integrationArtifactResourceOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
			Operation:         "delete",
			ResourcePath:      path,
		}
		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactResourceDelete", ResponseBody: ``, TestType: "Positive"}

		// test
		err = runIntegrationArtifactResource(&config, nil, &filesMock, &httpClient)

		// assert
		assert.NoError(t, err)
	})

	t.Run("Create Resource Negative Test", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		path := filepath.Join("tempDir", "demo.xsl")
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
		config := integrationArtifactResourceOptions{
			APIServiceKey:     apiServiceKey,
			IntegrationFlowID: "flow1",
			Operation:         "create",
			ResourcePath:      path,
		}
		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactResourceCreate", ResponseBody: ``, TestType: "Negative"}

		// test
		err = runIntegrationArtifactResource(&config, nil, &filesMock, &httpClient)

		// assert
		assert.Error(t, err)
	})
}
