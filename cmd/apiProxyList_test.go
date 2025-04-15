//go:build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/apim"
	apimhttp "github.com/SAP/jenkins-library/pkg/apim"
	"github.com/stretchr/testify/assert"
)

func TestRunApiProxyList(t *testing.T) {
	t.Parallel()

	t.Run("Get API Proxy List successfull test", func(t *testing.T) {
		config := getDefaultOptionsForApiProxyList()
		httpClientMock := &apimhttp.HttpMockAPIM{StatusCode: 200, ResponseBody: `{"some": "test"}`}
		seOut := apiProxyListCommonPipelineEnvironment{}
		apim := apim.Bundle{APIServiceKey: config.APIServiceKey, Client: httpClientMock}
		// test
		err := getApiProxyList(&config, apim, &seOut)

		// assert
		if assert.NoError(t, err) {
			assert.EqualValues(t, seOut.custom.APIProxyList, "{\"some\": \"test\"}")
			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "/apiportal/api/1.0/Management.svc/APIProxies?filter=isCopy+eq+false&$orderby=name&$skip=1&$top=4", httpClientMock.URL)
			})
			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClientMock.Method)
			})
		}
	})

	t.Run("Get API Proxy List failed test", func(t *testing.T) {
		config := getDefaultOptionsForApiProxyList()
		httpClientMock := &apimhttp.HttpMockAPIM{StatusCode: 400}
		seOut := apiProxyListCommonPipelineEnvironment{}
		apim := apim.Bundle{APIServiceKey: config.APIServiceKey, Client: httpClientMock}
		// test
		err := getApiProxyList(&config, apim, &seOut)
		// assert
		assert.EqualError(t, err, "HTTP GET request to /apiportal/api/1.0/Management.svc/APIProxies?filter=isCopy+eq+false&$orderby=name&$skip=1&$top=4 failed with error: : Bad Request")
	})
}

func getDefaultOptionsForApiProxyList() apiProxyListOptions {
	return apiProxyListOptions{
		APIServiceKey: apimhttp.GetServiceKey(),
		Top:           4,
		Skip:          1,
		Filter:        "isCopy eq false",
		Orderby:       "name",
	}
}
