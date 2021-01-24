// +build !release

package cpi

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

//GetCPIFunctionMockResponse -Generate mock response payload for different CPI functions
func GetCPIFunctionMockResponse(functionName, testType string) (*http.Response, error) {
	switch functionName {
	case "IntegrationArtifactDeploy":
		if testType == "Positive" {
			return GetEmptyHTTPResponseBody()
		}
		res := http.Response{
			StatusCode: 500,
			Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
						"code": "Internal Server Error",
						"message": {
						"@lang": "en",
						"#text": "Cannot deploy artifact with Id 'flow1'!"
						}
					}`))),
		}
		return &res, errors.New("Internal Server Error")
	default:
		res := http.Response{
			StatusCode: 404,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}
		return &res, errors.New("Service not Found")
	}
}

//GetEmptyHTTPResponseBody -Empty http respose body
func GetEmptyHTTPResponseBody() (*http.Response, error) {
	res := http.Response{
		StatusCode: 202,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
	}
	return &res, nil
}
