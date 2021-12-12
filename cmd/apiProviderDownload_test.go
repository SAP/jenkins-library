package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type apiProviderDownloadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func TestRunApiProviderDownload(t *testing.T) {
	t.Parallel()

	t.Run("Successful Download of API Provider", func(t *testing.T) {
		file, err := ioutil.TempFile("", "provider1.json")
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

		config := apiProviderDownloadOptions{
			APIServiceKey:   apiServiceKey,
			APIProviderName: "provider1",
			DownloadPath:    file.Name(),
		}
		// test
		httpClient := httpMockCpis{CPIFunction: "APIProviderDownload", ResponseBody: ``, TestType: "Positive"}
		errResp := runApiProviderDownload(&config, nil, &httpClient)

		if assert.NoError(t, errResp) {
			t.Run("check file", func(t *testing.T) {
				assert.Equal(t, fileExists(file.Name()), true)
			})
			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "https://demo/apiportal/api/1.0/Management.svc/APIProviders('provider1')", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})
		}
	})

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
			DownloadPath:    "tmp",
		}
		httpClient := httpMockCpis{CPIFunction: "APIProviderDownloadFailure", ResponseBody: ``, TestType: "Negative"}
		errResp := runApiProviderDownload(&config, nil, &httpClient)
		assert.EqualError(t, errResp, "HTTP GET request to https://demo/apiportal/api/1.0/Management.svc/APIProviders('provider1') failed with error: Service not Found")
	})
}
