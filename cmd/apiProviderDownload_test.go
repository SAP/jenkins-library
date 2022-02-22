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

//Successful API Provider download cases
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
		// test
		httpClient := httpMockCpis{CPIFunction: "APIProviderDownload", ResponseBody: ``, TestType: "Positive"}
		utilsMock := apiProviderDownloadMockUtilsBundle()
		err := runApiProviderDownload(&config, nil, &httpClient, utilsMock)

		if assert.NoError(t, err) {
			t.Run("Assert file download", func(t *testing.T) {
				assert.True(t, utilsMock.HasWrittenFile("APIProvider.json"))
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

//API Provider download failure cases
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

		assert.False(t, utilsMock.HasWrittenFile("APIProvider.json"))

		assert.EqualError(t, err, "HTTP GET request to https://demo/apiportal/api/1.0/Management.svc/APIProviders('provider1') failed with error: Service not Found")
	})
}
