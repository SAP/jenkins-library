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
		return GetEmptyHTTPResponseBodyAndErrorNil()
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
	case "GetIntegrationArtifactDeployStatus":
		return GetIntegrationArtifactDeployStatusMockResponse(testType)
	case "GetIntegrationArtifactDeployErrorDetails":
		return GetIntegrationArtifactDeployErrorDetailsMockResponse(testType)
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

	return GetMockResponseByTestTypeAndMockFunctionName("IntegrationArtifactDownloadCommandMockResponse", testType)
}

//GetIntegrationArtifactDeployErrorDetailsMockResponse -Provide http respose body
func GetIntegrationArtifactDeployErrorDetailsMockResponse(testType string) (*http.Response, error) {

	return GetMockResponseByTestTypeAndMockFunctionName("GetIntegrationArtifactDeployErrorDetailsMockResponse", "PostiveAndGetDeployedIntegrationDesigntimeArtifactErrorResBody")
}

//GetIntegrationDesigntimeArtifactMockResponse -Provide http respose body
func GetIntegrationDesigntimeArtifactMockResponse(testType string) (*http.Response, error) {

	return GetMockResponseByTestTypeAndMockFunctionName("GetIntegrationDesigntimeArtifactMockResponse", testType)
}

//IntegrationDesigntimeArtifactUpdateMockResponse -Provide http respose body
func IntegrationDesigntimeArtifactUpdateMockResponse(testType string) (*http.Response, error) {

	return GetMockResponseByTestTypeAndMockFunctionName("IntegrationDesigntimeArtifactUpdateMockResponse", testType)
}

//GetIntegrationArtifactDeployStatusMockResponse -Provide http respose body
func GetIntegrationArtifactDeployStatusMockResponse(testType string) (*http.Response, error) {

	return GetMockResponseByTestTypeAndMockFunctionName("GetIntegrationArtifactDeployStatusMockResponse", testType)
}

//GetIntegrationArtifactDeployStatusMockResponseBody -Provide http respose body
func GetIntegrationArtifactDeployStatusMockResponseBody() (*http.Response, error) {

	resp := http.Response{
		StatusCode: 200,
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
			"d": {
				"__metadata": {
					"id": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com/api/v1/IntegrationRuntimeArtifacts('flow1')",
					"uri": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com/api/v1/IntegrationRuntimeArtifacts('flow1')",
					"type": "com.sap.hci.api.IntegrationRuntimeArtifact",
					"content_type": "application/octet-stream",
					"media_src": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com/api/v1/IntegrationRuntimeArtifacts('flow1')/$value",
					"edit_media": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com/api/v1/IntegrationRuntimeArtifacts('flow1')/$value"
				},
				"Id": "flow1",
				"Version": "1.0.24",
				"Name": "flow1",
				"Type": "INTEGRATION_FLOW",
				"DeployedBy": "sb-8ff0b149-c3e6-417e-ad27-21fa5a3349dd!b15187|it!b11463",
				"DeployedOn": "2021-02-18T11:57:02.885",
				"Status": "STARTED",
				"ErrorInformation": {
					"__deferred": {
						"uri": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com/api/v1/IntegrationRuntimeArtifacts('flow1')/ErrorInformation"
					}
				}
			}
		}`))),
	}
	return &resp, nil
}

//GetIntegrationArtifactDeployStatusErrorMockResponseBody -Provide http respose body
func GetIntegrationArtifactDeployStatusErrorMockResponseBody() (*http.Response, error) {

	resp := http.Response{
		StatusCode: 200,
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
			"d": {
				"__metadata": {
					"id": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com/api/v1/IntegrationRuntimeArtifacts('smtp')",
					"uri": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com/api/v1/IntegrationRuntimeArtifacts('smtp')",
					"type": "com.sap.hci.api.IntegrationRuntimeArtifact",
					"content_type": "application/octet-stream",
					"media_src": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com/api/v1/IntegrationRuntimeArtifacts('smtp')/$value",
					"edit_media": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com/api/v1/IntegrationRuntimeArtifacts('smtp')/$value"
				},
				"Id": "smtp",
				"Version": "2.0",
				"Name": "smtp",
				"Type": "INTEGRATION_FLOW",
				"DeployedBy": "sb-8ff0b149-c3e6-417e-ad27-21fa5a3349dd!b15187|it!b11463",
				"DeployedOn": "2021-02-18T10:52:57.501",
				"Status": "ERROR",
				"ErrorInformation": {
					"__deferred": {
						"uri": "https://roverpoc.it-accd002.cfapps.sap.hana.ondemand.com/api/v1/IntegrationRuntimeArtifacts('smtp')/ErrorInformation"
					}
				}
			}
		}`))),
	}
	return &resp, nil
}

//GetIntegrationArtifactDeployErrorStatusMockResponseBody -Provide http respose body
func GetIntegrationArtifactDeployErrorStatusMockResponseBody() (*http.Response, error) {

	resp := http.Response{
		StatusCode: 200,
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`{	"message": {
			"subsystemName": "CONTENT",
			"subsytemPartName": "CONTENT_DEPLOY",
			"messageId": "InstanceError",
			"messageText": ""
		},
		"parameter": [
			"{\"message\":\"ERROR\",\"childMessageInstances\":[{\"message\":\"EXCEPTION\",\"parameters\":[\"org.osgi.service.blueprint.container.ComponentDefinitionException: Error when instantiating bean MessageFlow_28_configurator of class null\"],\"childMessageInstances\":[{\"message\":\"CAUSE\",\"parameters\":[\"java.lang.IllegalStateException: No credentials for 'smtp' found\"]}]}]}"
		]
	}`))),
	}
	return &resp, nil
}

//GetMockResponseByTestTypeAndMockFunctionName - Get mock response by testtype and mock function name
func GetMockResponseByTestTypeAndMockFunctionName(mockFuntionName, testType string) (*http.Response, error) {

	response, error := GetPositiveCaseResponseByTestType(testType)

	switch mockFuntionName {

	case "IntegrationDesigntimeArtifactUpdateMockResponse":
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
	case "GetIntegrationDesigntimeArtifactMockResponse":
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

	case "IntegrationArtifactDownloadCommandMockResponse":
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

	case "GetIntegrationArtifactDeployStatusMockResponse":
		if response == nil && error == nil {

			res := http.Response{
				StatusCode: 400,
				Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
							"code": "Bad Request",
							"message": {
							"@lang": "en",
							"#text": "Bad request"
							}
						}`))),
			}
			return &res, errors.New("Unable to get integration artifact deploy status, Response Status code:400")
		}

	case "GetIntegrationArtifactDeployErrorDetailsMockResponse":
		if response == nil && error == nil {

			res := http.Response{
				StatusCode: 500,
				Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
							"code": "Internal Server Error",
							"message": {
							"@lang": "en",
							"#text": "Internal Processing Error"
							}
						}`))),
			}
			return &res, errors.New("Unable to get integration artifact deploy error status, Response Status code:400")
		}

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
	case "PostiveAndDeployIntegrationDesigntimeArtifactResBody":
		return GetIntegrationArtifactDeployStatusMockResponseBody()
	case "PostiveAndGetDeployedIntegrationDesigntimeArtifactErrorResBody":
		return GetIntegrationArtifactDeployErrorStatusMockResponseBody()
	case "NegativeAndDeployIntegrationDesigntimeArtifactResBody":
		return GetIntegrationArtifactDeployStatusErrorMockResponseBody()
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
	case "https://demo/api/v1/DeployIntegrationDesigntimeArtifact?Id='flow1'&Version='1.0.1'":
		return GetFunctionNameByTestTypeAndMethod(method, testType)
	case "https://demo/api/v1/IntegrationRuntimeArtifacts('flow1')":
		return "GetIntegrationArtifactDeployStatus"
	case "https://demo/api/v1/IntegrationRuntimeArtifacts('flow1')/ErrorInformation/$value":
		return "GetIntegrationArtifactDeployErrorDetails"
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
	case "PostiveAndDeployIntegrationDesigntimeArtifactResBody":
		if method == "POST" {
			return "IntegrationArtifactDeploy"
		}

	case "NegativeAndDeployIntegrationDesigntimeArtifactResBody":
		if method == "POST" {
			return "IntegrationArtifactDeploy"
		}

	default:
		return ""

	}
	return ""
}
