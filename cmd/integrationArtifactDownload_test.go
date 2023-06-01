//go:build unit
// +build unit

package cmd

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunIntegrationArtifactDownload(t *testing.T) {
	t.Parallel()

	t.Run("Successfull Download of Integration flow Artifact", func(t *testing.T) {
		tempDir := t.TempDir()
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := integrationArtifactDownloadOptions{
			APIServiceKey:          apiServiceKey,
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			DownloadPath:           tempDir,
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactDownload", ResponseBody: ``, TestType: "PositiveAndGetetIntegrationArtifactDownloadResBody"}

		err := runIntegrationArtifactDownload(&config, nil, &httpClient)
		absolutePath := filepath.Join(tempDir, "flow1.zip")
		assert.DirExists(t, tempDir)
		if assert.NoError(t, err) {
			assert.Equal(t, fileExists(absolutePath), true)
			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/api/v1/IntegrationDesigntimeArtifacts(Id='flow1',Version='1.0.1')/$value", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})
		}
	})

	t.Run("Failed case of Integration Flow artifact Download", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := integrationArtifactDownloadOptions{
			APIServiceKey:          apiServiceKey,
			IntegrationFlowID:      "flow1",
			IntegrationFlowVersion: "1.0.1",
			DownloadPath:           "tmp",
		}

		httpClient := httpMockCpis{CPIFunction: "IntegrationArtifactDownload", ResponseBody: ``, TestType: "Negative"}

		err := runIntegrationArtifactDownload(&config, nil, &httpClient)

		assert.EqualError(t, err, "HTTP GET request to https://demo/api/v1/IntegrationDesigntimeArtifacts(Id='flow1',Version='1.0.1')/$value failed with error: Unable to download integration artifact, Response Status code:400")
	})
}
