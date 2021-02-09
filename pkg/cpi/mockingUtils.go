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
			return GetEmptyHTTPResponseBodyAndErrorNil()
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
	case "IntegrationArtifactUpdateConfiguration":
		if testType == "Positive" {
			return GetEmptyHTTPResponseBodyAndErrorNil()
		}
		if testType == "Negative_With_ResponseBody" {
			return GetNegativeCaseHTTPResponseBodyAndErrorNil()
		}
		res := http.Response{
			StatusCode: 404,
			Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
						"code": "Not Found",
						"message": {
						"@lang": "en",
						"#text": "Parameter key 'Parameter1' not found."
						}
					}`))),
		}
		return &res, errors.New("Not found - either wrong version for the given Id or wrong parameter key")
	case "IntegrationArtifactGetMplStatus":
		return GetIntegrationArtifactGetMplStatusCommandMockResponse(testType)
	case "IntegrationArtifactGetServiceEndpoint":
		return GetIntegrationArtifactGetServiceEndpointCommandMockResponse(testType)
	case "IntegrationArtifactDownload":
		return IntegrationArtifactDownloadCommandMockResponse(testType)
	default:
		res := http.Response{
			StatusCode: 404,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}
		return &res, errors.New("Service not Found")
	}
}

//GetEmptyHTTPResponseBodyAndErrorNil -Empty http respose body
func GetEmptyHTTPResponseBodyAndErrorNil() (*http.Response, error) {
	res := http.Response{
		StatusCode: 202,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
	}
	return &res, nil
}

//GetNegativeCaseHTTPResponseBodyAndErrorNil -Negative case http respose body
func GetNegativeCaseHTTPResponseBodyAndErrorNil() (*http.Response, error) {
	res := http.Response{
		StatusCode: 400,
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
					"code": "Bad Request",
					"message": {
					"@lang": "en",
					"#text": "Wrong body format for the expected parameter value"
					}
				}`))),
	}
	return &res, nil
}

//GetIntegrationArtifactGetMplStatusCommandMockResponse -Provide http respose body
func GetIntegrationArtifactGetMplStatusCommandMockResponse(testType string) (*http.Response, error) {
	if testType == "Positive" {
		res := http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
				"d": {
					"results": [
						{
							"__metadata": {
								"id": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com:443/api/v1/MessageProcessingLogs('AGAS1GcWkfBv-ZtpS6j7TKjReO7t')",
								"uri": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com:443/api/v1/MessageProcessingLogs('AGAS1GcWkfBv-ZtpS6j7TKjReO7t')",
								"type": "com.sap.hci.api.MessageProcessingLog"
							},
							"MessageGuid": "AGAS1GcWkfBv-ZtpS6j7TKjReO7t",
							"CorrelationId": "AGAS1GevYrPodxieoYf4YSY4jd-8",
							"ApplicationMessageId": null,
							"ApplicationMessageType": null,
							"LogStart": "/Date(1611846759005)/",
							"LogEnd": "/Date(1611846759032)/",
							"Sender": null,
							"Receiver": null,
							"IntegrationFlowName": "flow1",
							"Status": "COMPLETED",
							"LogLevel": "INFO",
							"CustomStatus": "COMPLETED",
							"TransactionId": "aa220151116748eeae69db3e88f2bbc8"
						}
					]
				}
			}`))),
		}
		return &res, nil
	}
	res := http.Response{
		StatusCode: 400,
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
					"code": "Bad Request",
					"message": {
					"@lang": "en",
					"#text": "Invalid order by expression"
					}
				}`))),
	}
	return &res, errors.New("Unable to get integration flow MPL status, Response Status code:400")
}

//GetIntegrationArtifactGetServiceEndpointCommandMockResponse -Provide http respose body
func GetIntegrationArtifactGetServiceEndpointCommandMockResponse(testCaseType string) (*http.Response, error) {
	if testCaseType == "PositiveAndGetetIntegrationArtifactGetServiceResBody" {
		return GetIntegrationArtifactGetServiceEndpointPositiveCaseRespBody()
	}
	res := http.Response{
		StatusCode: 400,
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
					"code": "Bad Request",
					"message": {
					"@lang": "en",
					"#text": "invalid service endpoint query"
					}
				}`))),
	}
	return &res, errors.New("Unable to get integration flow service endpoint, Response Status code:400")
}

//GetIntegrationArtifactGetServiceEndpointPositiveCaseRespBody -Provide http respose body for positive case
func GetIntegrationArtifactGetServiceEndpointPositiveCaseRespBody() (*http.Response, error) {

	resp := http.Response{
		StatusCode: 200,
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
			"d": {
				"results": [
					{
						"__metadata": {
							"id": "https://demo.cfapps.sap.hana.ondemand.com:443/api/v1/ServiceEndpoints('CPI_IFlow_Call_using_Cert%24endpointAddress%3Dtestwithcert')",
							"uri": "https://demo.cfapps.sap.hana.ondemand.com:443/api/v1/ServiceEndpoints('CPI_IFlow_Call_using_Cert%24endpointAddress%3Dtestwithcert')",
							"type": "com.sap.hci.api.ServiceEndpoint"
						},
						"Name": "CPI_IFlow_Call_using_Cert",
						"Id": "CPI_IFlow_Call_using_Cert$endpointAddress=testwithcert",
						"EntryPoints": {
							"results": [
								{
									"__metadata": {
										"id": "https://demo.cfapps.sap.hana.ondemand.com:443/api/v1/EntryPoints('https%3A%2F%2Froverpoc.it-accd002-rt.cfapps.sap.hana.ondemand.com%2Fhttp%2Ftestwithcert')",
										"uri": "https://demo.cfapps.sap.hana.ondemand.com:443/api/v1/EntryPoints('https%3A%2F%2Froverpoc.it-accd002-rt.cfapps.sap.hana.ondemand.com%2Fhttp%2Ftestwithcert')",
										"type": "com.sap.hci.api.EntryPoint"
									},
									"Name": "CPI_IFlow_Call_using_Cert",
									"Url": "https://demo.cfapps.sap.hana.ondemand.com/http/testwithcert",
									"Type": "PROD",
									"AdditionalInformation": ""
								}
							]
						}
					}
				]
			}
		}`))),
	}
	return &resp, nil
}

//IntegrationArtifactDownloadCommandMockResponse -Provide http respose body
func IntegrationArtifactDownloadCommandMockResponse(testCaseType string) (*http.Response, error) {
	if testCaseType != "Negative" {
		return IntegrationArtifactDownloadCommandMockResponsePositiveCaseRespBody()
	}
	res := http.Response{
		StatusCode: 400,
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
					"code": "Bad Request",
					"message": {
					"@lang": "en",
					"#text": "invalid request"
					}
				}`))),
	}
	return &res, errors.New("Unable to download integration artifact, Response Status code:400")
}

//IntegrationArtifactDownloadCommandMockResponsePositiveCaseRespBody -Provide http respose body for positive case
func IntegrationArtifactDownloadCommandMockResponsePositiveCaseRespBody() (*http.Response, error) {
	header := make(http.Header)
	headerValue := "attachment; filename=flow1.zip"
	header.Add("Content-Disposition", headerValue)
	resp := http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`UEsDBBQACAgIADQ2clAAAAAAAAAAAAAAAAAUAAQATU`))),
	}
	return &resp, nil
}
