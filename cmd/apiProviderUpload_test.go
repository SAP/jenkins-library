//go:build unit
// +build unit

package cmd

import (
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/apim"
	apimhttp "github.com/SAP/jenkins-library/pkg/apim"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunApiProviderUpload(t *testing.T) {
	t.Parallel()

	t.Run("API Provider upload succesfull test", func(t *testing.T) {
		file, tmpErr := os.CreateTemp("", "test.json")
		if tmpErr != nil {
			t.FailNow()
		}
		defer os.RemoveAll(file.Name()) // clean up
		filesMock := mock.FilesMock{}
		filesMock.AddFile(file.Name(), []byte(apimhttp.GetServiceKey()))
		config := getDefaultOptionsForApiProvider()
		config.FilePath = file.Name()
		httpClientMock := &apimhttp.HttpMockAPIM{StatusCode: 201, ResponseBody: ``}
		apim := apim.Bundle{APIServiceKey: config.APIServiceKey, Client: httpClientMock}
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

	t.Run("API Provider upload failed test", func(t *testing.T) {
		file, tmpErr := os.CreateTemp("", "test.json")
		if tmpErr != nil {
			t.FailNow()
		}
		defer os.RemoveAll(file.Name()) // clean up
		filesMock := mock.FilesMock{}
		filesMock.AddFile(file.Name(), []byte(apimhttp.GetServiceKey()))
		config := getDefaultOptionsForApiProvider()
		config.FilePath = file.Name()
		httpClientMock := &apimhttp.HttpMockAPIM{StatusCode: 400}
		apim := apim.Bundle{APIServiceKey: config.APIServiceKey, Client: httpClientMock}
		// test
		err := createApiProvider(&config, apim, filesMock.FileRead)
		// assert
		assert.EqualError(t, err, "HTTP POST request to /apiportal/api/1.0/Management.svc/APIProviders failed with error: : Bad Request")
	})

	t.Run("valid api provider payload test", func(t *testing.T) {
		apiProviderPayload := `{
			"oauth": {
				"url": "https://demo",
				"clientid": "demouser",
				"clientsecret": "******",
				"tokenurl": "https://demo/oauth/token"
				}
			}`
		apimData := apim.Bundle{Payload: apiProviderPayload}
		assert.Equal(t, apimData.IsPayloadJSON(), true)
	})

	t.Run("invalid api provider payload test", func(t *testing.T) {
		apiProviderPayload := `this is not json`
		apimData := apim.Bundle{Payload: apiProviderPayload}
		assert.Equal(t, apimData.IsPayloadJSON(), false)
	})

}

func getDefaultOptionsForApiProvider() apiProviderUploadOptions {
	return apiProviderUploadOptions{
		APIServiceKey: apimhttp.GetServiceKey(),
		FilePath:      "test.json",
	}
}
