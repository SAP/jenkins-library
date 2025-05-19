//go:build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type apiProviderDownloadTestUtilsBundle struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func apiProviderDownloadMockUtilsBundle() *apiProviderDownloadTestUtilsBundle {
	utilsBundle := apiProviderDownloadTestUtilsBundle{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return &utilsBundle
}

// Successful API Provider download cases
func TestApiProviderDownloadSuccess(t *testing.T) {
	t.Parallel()
	t.Run("Successful Download of API Provider", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`

		config := apiProviderDownloadOptions{
			APIServiceKey:   apiServiceKey,
			APIProviderName: "provider1",
			DownloadPath:    "APIProvider.json",
		}
		httpClient := httpMockCpis{CPIFunction: "APIProviderDownload", ResponseBody: ``, TestType: "Positive"}
		utilsMock := apiProviderDownloadMockUtilsBundle()
		err := runApiProviderDownload(&config, nil, &httpClient, utilsMock)

		if assert.NoError(t, err) {

			t.Run("Assert file download & content", func(t *testing.T) {
				fileExist := assert.True(t, utilsMock.HasWrittenFile(config.DownloadPath))
				if fileExist {
					providerbyteContent, _ := utilsMock.FileRead(config.DownloadPath)
					providerContent := string(providerbyteContent)
					assert.Equal(t, providerContent, "{\n\t\t\t\t\"d\": {\n\t\t\t\t\t\"results\": [\n\t\t\t\t\t\t{\n\t\t\t\t\t\t\t\"__metadata\": {\n\t\t\t\t\t\t\t\t\"id\": \"https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com:443/api/v1/MessageProcessingLogs('AGAS1GcWkfBv-ZtpS6j7TKjReO7t')\",\n\t\t\t\t\t\t\t\t\"uri\": \"https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com:443/api/v1/MessageProcessingLogs('AGAS1GcWkfBv-ZtpS6j7TKjReO7t')\",\n\t\t\t\t\t\t\t\t\"type\": \"com.sap.hci.api.MessageProcessingLog\"\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t\"MessageGuid\": \"AGAS1GcWkfBv-ZtpS6j7TKjReO7t\",\n\t\t\t\t\t\t\t\"CorrelationId\": \"AGAS1GevYrPodxieoYf4YSY4jd-8\",\n\t\t\t\t\t\t\t\"ApplicationMessageId\": null,\n\t\t\t\t\t\t\t\"ApplicationMessageType\": null,\n\t\t\t\t\t\t\t\"LogStart\": \"/Date(1611846759005)/\",\n\t\t\t\t\t\t\t\"LogEnd\": \"/Date(1611846759032)/\",\n\t\t\t\t\t\t\t\"Sender\": null,\n\t\t\t\t\t\t\t\"Receiver\": null,\n\t\t\t\t\t\t\t\"IntegrationFlowName\": \"flow1\",\n\t\t\t\t\t\t\t\"Status\": \"COMPLETED\",\n\t\t\t\t\t\t\t\"LogLevel\": \"INFO\",\n\t\t\t\t\t\t\t\"CustomStatus\": \"COMPLETED\",\n\t\t\t\t\t\t\t\"TransactionId\": \"aa220151116748eeae69db3e88f2bbc8\"\n\t\t\t\t\t\t}\n\t\t\t\t\t]\n\t\t\t\t}\n\t\t\t}")
				}
			})

			t.Run("Assert API Provider url", func(t *testing.T) {
				assert.Equal(t, "https://demo/apiportal/api/1.0/Management.svc/APIProviders('provider1')", httpClient.URL)
			})

			t.Run("Assert method as GET", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})
		}
	})
}

// API Provider download failure cases
func TestApiProviderDownloadFailure(t *testing.T) {
	t.Parallel()

	t.Run("Failed case of API Provider Download", func(t *testing.T) {
		apiServiceKey := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`
		config := apiProviderDownloadOptions{
			APIServiceKey:   apiServiceKey,
			APIProviderName: "provider1",
			DownloadPath:    "APIProvider.json",
		}
		httpClient := httpMockCpis{CPIFunction: "APIProviderDownloadFailure", ResponseBody: ``, TestType: "Negative"}
		utilsMock := apiProviderDownloadMockUtilsBundle()
		err := runApiProviderDownload(&config, nil, &httpClient, utilsMock)

		assert.False(t, utilsMock.HasWrittenFile(config.DownloadPath))

		assert.EqualError(t, err, "HTTP GET request to https://demo/apiportal/api/1.0/Management.svc/APIProviders('provider1') failed with error: Service not Found")
	})
}
