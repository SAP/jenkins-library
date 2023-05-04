//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/apim"
	apimhttp "github.com/SAP/jenkins-library/pkg/apim"
	"github.com/stretchr/testify/assert"
)

func TestRunApiProviderList(t *testing.T) {
	t.Parallel()

	t.Run("Get API providers successfull test", func(t *testing.T) {
		config := getDefaultOptionsForApiProviderList()
		httpClientMock := &apimhttp.HttpMockAPIM{StatusCode: 200, ResponseBody: `{"some": "test"}`}
		seOut := apiProviderListCommonPipelineEnvironment{}
		apim := apim.Bundle{APIServiceKey: config.APIServiceKey, Client: httpClientMock}
		// test
		err := getApiProviderList(&config, apim, &seOut)

		// assert
		if assert.NoError(t, err) {
			assert.EqualValues(t, seOut.custom.APIProviderList, "{\"some\": \"test\"}")
			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "/apiportal/api/1.0/Management.svc/APIProviders?orderby=value&$select=name&$top=2", httpClientMock.URL)
			})
			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClientMock.Method)
			})
		}
	})

	t.Run("Get API provider failed test", func(t *testing.T) {
		config := getDefaultOptionsForApiProviderList()
		httpClientMock := &apimhttp.HttpMockAPIM{StatusCode: 400}
		seOut := apiProviderListCommonPipelineEnvironment{}
		apim := apim.Bundle{APIServiceKey: config.APIServiceKey, Client: httpClientMock}
		// test
		err := getApiProviderList(&config, apim, &seOut)
		// assert
		assert.EqualError(t, err, "HTTP GET request to /apiportal/api/1.0/Management.svc/APIProviders?orderby=value&$select=name&$top=2 failed with error: : Bad Request")
	})
}

func getDefaultOptionsForApiProviderList() apiProviderListOptions {
	return apiProviderListOptions{
		APIServiceKey: apimhttp.GetServiceKey(),
		Top:           2,
		Select:        "name",
		Orderby:       "value",
	}
}
