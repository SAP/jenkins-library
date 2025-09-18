package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunApiKeyValueMapDownload(t *testing.T) {
	t.Parallel()

	t.Run("Successfull Download of API Key Value Map", func(t *testing.T) {
		file, err := os.CreateTemp("", "CustKVM.json")
		if err != nil {
			t.FailNow()
		}
		defer os.RemoveAll(file.Name()) // clean up
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := apiKeyValueMapDownloadOptions{
			APIServiceKey:   apiServiceKey,
			KeyValueMapName: "CustKVM",
			DownloadPath:    file.Name(),
		}
		httpClient := httpMockCpis{CPIFunction: "APIKeyValueMapDownload", ResponseBody: ``, TestType: "Positive"}
		errResp := runApiKeyValueMapDownload(&config, nil, &httpClient)
		if assert.NoError(t, errResp) {
			t.Run("check file", func(t *testing.T) {
				assert.Equal(t, fileExists(file.Name()), true)
			})

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/apiportal/api/1.0/Management.svc/KeyMapEntries('CustKVM')", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})
		}
	})

	t.Run("Failed case of API Key Value Map Download", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := apiKeyValueMapDownloadOptions{
			APIServiceKey:   apiServiceKey,
			KeyValueMapName: "CustKVM",
			DownloadPath:    "",
		}
		httpClient := httpMockCpis{CPIFunction: "APIKeyValueMapDownloadFailure", ResponseBody: ``, TestType: "Negative"}
		err := runApiKeyValueMapDownload(&config, nil, &httpClient)
		assert.EqualError(t, err, "HTTP GET request to https://demo/apiportal/api/1.0/Management.svc/KeyMapEntries('CustKVM') failed with error: Service not Found")
	})
}
