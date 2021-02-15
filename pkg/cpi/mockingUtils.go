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
		return GetParameterKeyMissingResponseBody()
	case "IntegrationArtifactGetMplStatus":
		return GetIntegrationArtifactGetMplStatusCommandMockResponse(testType)
	case "IntegrationArtifactGetServiceEndpoint":
		return GetIntegrationArtifactGetServiceEndpointCommandMockResponse(testType)
	case "IntegrationArtifactDownload":
		return IntegrationArtifactDownloadCommandMockResponse(testType)
	case "GetIntegrationDesigntimeArtifact":
		return GetIntegrationDesigntimeArtifactMockResponse(testType)
	case "UploadIntegrationDesigntimeArtifact":
		return GetIntegrationDesigntimeArtifactMockResponse(testType)
	case "UploadIntegrationDesigntimeArtifactNegative":
		return GetRespBodyHTTPStatusServiceErrorResponse()
	case "UpdateIntegrationDesigntimeArtifactNegative":
		return GetRespBodyHTTPStatusServiceErrorResponse()
	case "UpdateIntegrationDesigntimeArtifact":
		return UpdateIntegrationDesigntimeArtifactMockResponse(testType)
	case "IntegrationDesigntimeArtifactUpdate":
		return IntegrationDesigntimeArtifactUpdateMockResponse(testType)
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

//GetParameterKeyMissingResponseBody -Parameter key missing http respose body
func GetParameterKeyMissingResponseBody() (*http.Response, error) {
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

//GetRespBodyHTTPStatusOK -Provide http respose body for Http StatusOK
func GetRespBodyHTTPStatusOK() (*http.Response, error) {

	resp := http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
	}
	return &resp, nil
}

//GetRespBodyHTTPStatusCreated -Provide http respose body for Http StatusOK
func GetRespBodyHTTPStatusCreated() (*http.Response, error) {

	resp := http.Response{
		StatusCode: 201,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
	}
	return &resp, nil
}

//GetRespBodyHTTPStatusServiceNotFound -Provide http respose body for Http URL not Found
func GetRespBodyHTTPStatusServiceNotFound() (*http.Response, error) {

	resp := http.Response{
		StatusCode: 404,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
	}
	return &resp, errors.New("Integration Package not found")
}

//GetRespBodyHTTPStatusServiceErrorResponse -Provide http respose body for server error
func GetRespBodyHTTPStatusServiceErrorResponse() (*http.Response, error) {

	resp := http.Response{
		StatusCode: 500,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
	}
	return &resp, errors.New("Internal error")
}

//IntegrationArtifactDownloadCommandMockResponse -Provide http respose body
func IntegrationArtifactDownloadCommandMockResponse(testType string) (*http.Response, error) {

	response, error := GetPositiveCaseResponseByTestType(testType)

	if response == nil && error == nil {

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
	return response, error
}

//GetIntegrationDesigntimeArtifactMockResponse -Provide http respose body
func GetIntegrationDesigntimeArtifactMockResponse(testType string) (*http.Response, error) {

	response, error := GetPositiveCaseResponseByTestType(testType)

	if response == nil && error == nil {

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
		return &res, errors.New("Unable to get status of integration artifact, Response Status code:400")
	}
	return response, error
}

//IntegrationDesigntimeArtifactUpdateMockResponse -Provide http respose body
func IntegrationDesigntimeArtifactUpdateMockResponse(testType string) (*http.Response, error) {

	response, error := GetPositiveCaseResponseByTestType(testType)

	if response == nil && error == nil {

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
		return &res, errors.New("Unable to get status of integration artifact, Response Status code:400")
	}
	return response, error
}

//UpdateIntegrationDesigntimeArtifactMockResponse -Provide http respose body
func UpdateIntegrationDesigntimeArtifactMockResponse(testType string) (*http.Response, error) {

	response, error := GetRespBodyHTTPStatusCreated()

	if response == nil && error == nil {

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
		return &res, errors.New("Unable to get status of integration artifact, Response Status code:400")
	}
	return response, error
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

//GetPositiveCaseResponseByTestType - get postive response by test case type
func GetPositiveCaseResponseByTestType(testType string) (*http.Response, error) {
	switch testType {
	case "PositiveAndGetetIntegrationArtifactDownloadResBody":
		return IntegrationArtifactDownloadCommandMockResponsePositiveCaseRespBody()
	case "PositiveAndCreateIntegrationDesigntimeArtifactResBody":
		return GetRespBodyHTTPStatusOK()
	case "NegativeAndCreateIntegrationDesigntimeArtifactResBody":
		return GetRespBodyHTTPStatusOK()
	case "PositiveAndUpdateIntegrationDesigntimeArtifactResBody":
		return GetRespBodyHTTPStatusServiceNotFound()
	case "NegativeAndUpdateIntegrationDesigntimeArtifactResBody":
		return GetRespBodyHTTPStatusServiceNotFound()
	default:
		return nil, nil
	}
}

//GetCPIFunctionNameByURLCheck - get postive response by test case type
func GetCPIFunctionNameByURLCheck(url, method, testType string) string {
	switch url {
	case "https://demo/api/v1/IntegrationDesigntimeArtifacts(Id='flow4',Version='1.0.4')":
		return GetFunctionNameByTestTypeAndMethod(method, testType)

	case "https://demo/api/v1/IntegrationDesigntimeArtifactSaveAsVersion?Id='flow4'&SaveAsVersion='1.0.4'":
		return GetFunctionNameByTestTypeAndMethod(method, testType)

	case "https://demo/api/v1/IntegrationDesigntimeArtifacts":
		return GetFunctionNameByTestTypeAndMethod(method, testType)

	default:
		return ""
	}
}

//GetFunctionNameByTestTypeAndMethod -get function name by test tyep
func GetFunctionNameByTestTypeAndMethod(method, testType string) string {

	switch testType {

	case "PositiveAndCreateIntegrationDesigntimeArtifactResBody":
		if method == "GET" {
			return "GetIntegrationDesigntimeArtifact"
		}
		if method == "POST" {
			return "UploadIntegrationDesigntimeArtifact"
		}

	case "PositiveAndUpdateIntegrationDesigntimeArtifactResBody":
		if method == "GET" {
			return "IntegrationDesigntimeArtifactUpdate"
		}
		if method == "POST" {
			return "UpdateIntegrationDesigntimeArtifact"
		}

	case "NegativeAndGetIntegrationDesigntimeArtifactResBody":
		if method == "GET" {
			return "GetIntegrationDesigntimeArtifact"
		}

	case "NegativeAndCreateIntegrationDesigntimeArtifactResBody":
		if method == "GET" {
			return "GetIntegrationDesigntimeArtifact"
		}
		if method == "POST" {
			return "UploadIntegrationDesigntimeArtifactNegative"
		}

	case "NegativeAndUpdateIntegrationDesigntimeArtifactResBody":
		if method == "GET" {
			return "GetIntegrationDesigntimeArtifact"
		}
		if method == "POST" {
			return "UpdateIntegrationDesigntimeArtifactNegative"
		}
	default:
		return ""

	}
	return ""
}
