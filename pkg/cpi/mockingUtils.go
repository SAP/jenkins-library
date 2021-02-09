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
	if testCaseType == "PositiveAndGetetIntegrationArtifactDownloadServiceResBody" {
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
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`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TUVUQS1JTkYvZGVwbG95bWVudC9xdWV1ZURlZmluaXRpb25zLmpzb26rViosTS1NdUlNy8zLLMnMzytWsoqOrQUAUEsHCA2c0qEZAAAAFwAAAFBLAQIUABQACAgIADQ2clCF1dW2cgUAAEUQAAAUAAQAAAAAAAAAAAAAAAAAAABNRVRBLUlORi9NQU5JRkVTVC5NRv7KAABQSwECFAAUAAgICAA0NnJQAAAAAAIAAAAAAAAACQAAAAAAAAAAAAAAAAC4BQAAT1NHSS1JTkYvUEsBAhQAFAAICAgANDZyUAAAAAACAAAAAAAAABMAAAAAAAAAAAAAAAAA8QUAAE9TR0ktSU5GL2JsdWVwcmludC9QSwECFAAUAAgICAA0NnJQANQsu8IHAABHHgAAHAAAAAAAAAAAAAAAAAA0BgAAT1NHSS1JTkYvYmx1ZXByaW50L2JlYW5zLnhtbFBLAQIUABQACAgIADQ2clAAAAAAAgAAAAAAAAAJAAAAAAAAAAAAAAAAAEAOAABNRVRBLUlORi9QSwECFAAUAAgICAA0NnJQAAAAAAIAAAAAAAAACgAAAAAAAAAAAAAAAAB5DgAAcmVzb3VyY2VzL1BLAQIUABQACAgIADQ2clDVF1a0QwAAAFIAAAAcAAAAAAAAAAAAAAAAALMOAAByZXNvdXJjZXMvcGFyYW1ldGVycy5wcm9wZGVmUEsBAhQAFAAICAgANDZyUKhEW7QgAAAAHgAAABkAAAAAAAAAAAAAAAAAQA8AAHJlc291cmNlcy9wYXJhbWV0ZXJzLnByb3BQSwECFAAUAAgICAA0NnJQAAAAAAIAAAAAAAAABAAAAAAAAAAAAAAAAACnDwAAc3JjL1BLAQIUABQACAgIADQ2clAAAAAAAgAAAAAAAAAJAAAAAAAAAAAAAAAAANsPAABzcmMvbWFpbi9QSwECFAAUAAgICAA0NnJQAAAAAAIAAAAAAAAAEwAAAAAAAAAAAAAAAAAUEAAAc3JjL21haW4vcmVzb3VyY2VzL1BLAQIUABQACAgIADQ2clAAAAAAAgAAAAAAAAAhAAAAAAAAAAAAAAAAAFcQAABzcmMvbWFpbi9yZXNvdXJjZXMvc2NlbmFyaW9mbG93cy9QSwECFAAUAAgICAA0NnJQAAAAAAIAAAAAAAAAMQAAAAAAAAAAAAAAAACoEAAAc3JjL21haW4vcmVzb3VyY2VzL3NjZW5hcmlvZmxvd3MvaW50ZWdyYXRpb25mbG93L1BLAQIUABQACAgIADQ2clBEkIxpigkAALZSAAA7AAAAAAAAAAAAAAAAAAkRAABzcmMvbWFpbi9yZXNvdXJjZXMvc2NlbmFyaW9mbG93cy9pbnRlZ3JhdGlvbmZsb3cvYXp1cmUuaWZsd1BLAQIUABQACAgIADQ2clDVF1a0QwAAAFIAAAAlAAAAAAAAAAAAAAAAAPwaAABzcmMvbWFpbi9yZXNvdXJjZXMvcGFyYW1ldGVycy5wcm9wZGVmUEsBAhQAFAAICAgANDZyUKhEW7QgAAAAHgAAACIAAAAAAAAAAAAAAAAAkhsAAHNyYy9tYWluL3Jlc291cmNlcy9wYXJhbWV0ZXJzLnByb3BQSwECFAAUAAgICAA0NnJQDZzSoRkAAAAXAAAAKQAAAAAAAAAAAAAAAAACHAAATUVUQS1JTkYvZGVwbG95bWVudC9xdWV1ZURlZmluaXRpb25zLmpzb25QSwUGAAAAABEAEQDDBAAAchwAAAAA
	`))),
	}
	return &resp, nil
}
