package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/apim"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunApiProviderUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("test.json", []byte("Test content"))
		config := getDefaultOptionsForApiProvider()
		httpClientMock := httpMock{StatusCode: 201, ResponseBody: ``}
		apim := apim.APIMBundle{APIServiceKey: config.APIServiceKey, Client: &httpClientMock}
		// test
		err := createApiProvider(&config, apim, filesMock.FileRead)
		// assert
		if assert.NoError(t, err) {
			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "/apiportal/api/1.0/Management.svc/APIProviders", httpClientMock.URL)
			})
			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClientMock.Method)
			})
		}
	})

	t.Run("Failure Path", func(t *testing.T) {
		filesMock := mock.FilesMock{}
		filesMock.AddFile("test.json", []byte("Test content"))
		config := getDefaultOptionsForApiProvider()
		httpClientMock := httpMockGcts{StatusCode: 400, ResponseBody: ``}
		apim := apim.APIMBundle{APIServiceKey: config.APIServiceKey, Client: &httpClientMock}
		// test
		err := createApiProvider(&config, apim, filesMock.FileRead)
		// assert
		assert.EqualError(t, err, "HTTP POST request to /apiportal/api/1.0/Management.svc/APIProviders failed with error: : a http error occurred")
	})

}

func getDefaultOptionsForApiProvider() apiProviderUploadOptions {
	return apiProviderUploadOptions{
		APIServiceKey: `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
			}
		}`,
		FilePath: "test.json",
	}
}
