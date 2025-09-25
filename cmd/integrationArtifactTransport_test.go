package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/apim"
	apimhttp "github.com/SAP/jenkins-library/pkg/apim"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type integrationArtifactTransportMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newIntegrationArtifactTransportTestsUtils() integrationArtifactTransportMockUtils {
	utils := integrationArtifactTransportMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunIntegrationArtifactTransport(t *testing.T) {
	t.Parallel()

	t.Run("Create Transport Request Successful test", func(t *testing.T) {
		config := getDefaultOptionsForIntegrationArtifactTransport()
		httpClientMock := &apimhttp.HttpMockAPIM{StatusCode: 202, ResponseBody: `{"processId": "100", "state": "FINISHED"}`}
		apim := apim.Bundle{APIServiceKey: config.CasServiceKey, Client: httpClientMock}
		// test
		err := CreateIntegrationArtifactTransportRequest(&config, apim)
		// assert
		if assert.NoError(t, err) {
			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "/v1/operations/100", httpClientMock.URL)
			})
			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClientMock.Method)
			})
		}
	})

	t.Run("getIntegrationTransportProcessingStatus successful test", func(t *testing.T) {
		config := getDefaultOptionsForIntegrationArtifactTransport()
		httpClientMock := &apimhttp.HttpMockAPIM{StatusCode: 200, ResponseBody: `{"state": "FINISHED"}`}
		// test
		resp, err := getIntegrationTransportProcessingStatus(&config, httpClientMock, "demo", "100")

		// assert
		assert.Equal(t, "FINISHED", resp)
		assert.NoError(t, err)
	})

	t.Run("getIntegrationTransportError successful test", func(t *testing.T) {
		config := getDefaultOptionsForIntegrationArtifactTransport()
		httpClientMock := &apimhttp.HttpMockAPIM{StatusCode: 200, ResponseBody: `{ "logs": [] }`}
		// test
		resp, err := getIntegrationTransportError(&config, httpClientMock, "demo", "100")

		// assert
		assert.Equal(t, "{ \"logs\": [] }", resp)
		// assert
		if assert.NoError(t, err) {
			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "demo/v1/operations/100/logs", httpClientMock.URL)
			})
			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClientMock.Method)
			})
		}
	})

	t.Run("GetCPITransportReqPayload successful test", func(t *testing.T) {
		config := getDefaultOptionsForIntegrationArtifactTransport()
		// test
		resp, err := GetCPITransportReqPayload(&config)
		fmt.Println(resp.String())
		// assert
		expJson := `{"contentType":"d9c3fe08ceeb47a2991e53049f2ed766","id":"TestTransport","name":"TestTransport","resourceID":"d9c3fe08ceeb47a2991e53049f2ed766","subType":"package","type":"CloudIntegration","version":"1.0"}`
		actJson := resp.String()
		assert.Contains(t, actJson, expJson)
		assert.NoError(t, err)
	})

	t.Run("Create Transport Request negative test1", func(t *testing.T) {
		config := getDefaultOptionsForIntegrationArtifactTransport()
		httpClientMock := &apimhttp.HttpMockAPIM{StatusCode: 202, ResponseBody: `{"processId": ""}`}
		apim := apim.Bundle{APIServiceKey: config.CasServiceKey, Client: httpClientMock}
		// test
		err := CreateIntegrationArtifactTransportRequest(&config, apim)
		assert.Equal(t, "/v1/contentResources/export", httpClientMock.URL)
		assert.Equal(t, "POST", httpClientMock.Method)
		assert.Error(t, err)
	})

	t.Run("Create Transport Request negative test2", func(t *testing.T) {
		config := getDefaultOptionsForIntegrationArtifactTransport()
		httpClientMock := &apimhttp.HttpMockAPIM{StatusCode: 400, ResponseBody: ``}
		apim := apim.Bundle{APIServiceKey: config.CasServiceKey, Client: httpClientMock}
		// test
		err := CreateIntegrationArtifactTransportRequest(&config, apim)
		assert.EqualError(t, err, "HTTP POST request to /v1/contentResources/export failed with error: Bad Request")
	})

}

func getDefaultOptionsForIntegrationArtifactTransport() integrationArtifactTransportOptions {

	apiServiceKey := `{
		"oauth": {
			"url": "https://demo",
			"clientid": "sb-2d0622c9",
			"clientsecret": "edb5c506=",
			"tokenurl": "https://demo/oauth/token"
		}
	}`

	return integrationArtifactTransportOptions{
		CasServiceKey:        apiServiceKey,
		IntegrationPackageID: "TestTransport",
		ResourceID:           "d9c3fe08ceeb47a2991e53049f2ed766",
		Name:                 "TestTransport",
		Version:              "1.0",
	}
}
