package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockedClientReq struct {
	url    *url.URL
	method string
	header http.Header
	body   io.ReadCloser
}

type mockedServerResp struct {
	statusCode int
	headers    map[string]string
}

var lPAPIServiceKey = `{
	"url": "https://some-url.com",
	"uaa": {
	    "clientid": "some-client-id",
	    "url": "https://some-uaa-url.com",
	    "certificate": "-----BEGIN CERTIFICATE-----\nMIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw\nDgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow\nEjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d\n7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B\n5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr\nBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1\nNDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l\nWf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc\n6MF9+Yw1Yy0t\n-----END CERTIFICATE-----",
	    "certurl": "https://some-cert-url.com",
	    "key": "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49\nAwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q\nEKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==\n-----END EC PRIVATE KEY-----"
	},
	"vendor": "SAP"
    }`

var updateAddOnConfig = abapLandscapePortalUpdateAddOnProductOptions{
	LandscapePortalAPIServiceKey: lPAPIServiceKey,
	AbapSystemNumber:             "abap-system-number",
	AddonDescriptorFileName:      "./testdata/TestAbapLandscapePortalUpdateAddOnProduct/addon.yml",
	AddonDescriptor: `{
		"addonProduct": "some-addon-product",
		"addonVersion": "some-addon-version"
		}`,
}

var httpClient = http.Client{}
var httpClientAT = http.Client{}

var servKey = serviceKey{}

var resp_200 = mockedServerResp{
	statusCode: 200,
}
var resp_204 = mockedServerResp{
	statusCode: 204,
}
var resp_400 = mockedServerResp{
	statusCode: 400,
}

var wantedAccessToken = accessTokenResp{
	AccessToken: "some-access-token",
	TokenType:   "Bearer",
}

// this function is used to parse a raw url into *url.URL format
func parseRawURL(rawURL string) *url.URL {
	parsedURL, _ := url.Parse(rawURL)
	return parsedURL
}

// // this function is used to encode request body from struct to io.ReadCloser
func encodeReqBody[T any](reqBody T, url string) *http.Request {
	var reqBuff bytes.Buffer

	json.NewEncoder(&reqBuff).Encode(reqBody)

	httpReq, _ := http.NewRequest(http.MethodPost, url, &reqBuff)

	return httpReq
}

// this function is used to provide a mocked http server
func mockServer(req *mockedClientReq, resp mockedServerResp, wantedResult any) *httptest.Server {
	mockedServer := httptest.NewServer(http.HandlerFunc(func(handlerWriter http.ResponseWriter, handlerReq *http.Request) {
		// set up client request to the mocked server
		handlerReq.URL = req.url
		handlerReq.Method = req.method
		handlerReq.Header = req.header
		handlerReq.Body = req.body

		// write response header
		for key, value := range resp.headers {
			handlerWriter.Header().Add(key, value)
		}

		// write response status code
		handlerWriter.WriteHeader(resp.statusCode)

		// write response body
		if wantedResult != nil {
			jsonResp, _ := json.Marshal(wantedResult)
			handlerWriter.Write(jsonResp)
		}
	}))

	return mockedServer
}

// this function is used to generate a mocked request for getting LP API access token
func mockGetLPAPIAccessTokenReq(servKey serviceKey) mockedClientReq {
	rawURL := servKey.Uaa.CertUrl + "/oauth/token"

	reqBody := url.Values{}
	reqBody.Set("grant_type", "client_credentials")
	reqBody.Set("client_id", servKey.Uaa.ClientId)

	encodedReqBody := reqBody.Encode()

	httpReq, _ := http.NewRequest(http.MethodPost, rawURL, strings.NewReader(encodedReqBody))

	req := mockedClientReq{
		url:    parseRawURL(rawURL),
		method: http.MethodPost,
		header: httpReq.Header,
		body:   httpReq.Body,
	}

	return req
}

// this function is used to mock getLPAPIAccessToken
func mockGetLPAPIAccessToken(client http.Client, servKey serviceKey) (string, serviceKey) {
	servKey_temp := servKey

	req := mockGetLPAPIAccessTokenReq(servKey)
	mockedServer := mockServer(&req, resp_200, wantedAccessToken)

	servKey_temp.Uaa.CertUrl = mockedServer.URL

	accessToken, _ := getLPAPIAccessToken(httpClientAT, servKey_temp)

	return accessToken, servKey_temp
}

// this function is used to convert mockedClientReq to http.Request
func convertMockedReqToHttpReq(mockedReq mockedClientReq) http.Request {
	httpReq := http.Request{
		URL:    mockedReq.url,
		Method: mockedReq.method,
		Header: mockedReq.header,
		Body:   mockedReq.body,
	}

	return httpReq
}

// this function is used to mock the last part of runAbapLandcapePortalUpdateAddOnProduct "keep pulling status of update AddOn request until it reaches a final status (C/E/X)"
func keepPullingUntilFinalStatusReached(fromStatus *string, toStatus string, mockedServer *httptest.Server, req mockedClientReq) error {
	httpReq := convertMockedReqToHttpReq(req)

	// mock the process of pulling status of update AddOn request
	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		err := pullStatusOfUpdateAddOn(httpClient, &httpReq, "some-request-guid", fromStatus)

		if err != nil {
			return err
		}
	}

	// mock the server which returns the update AddOn request with a final status, and update the url of http request
	wantedRequest := reqEntity{
		RequestId: "some-req-id",
		ZoneId:    "some-zone-id",
		Status:    toStatus,
		SystemId:  "some-system-id",
	}

	mockedServer = mockServer(&req, resp_200, wantedRequest)
	req.url = parseRawURL(mockedServer.URL + "/api/v1.0/requests/:some-req-id")
	httpReq = convertMockedReqToHttpReq(req)

	// pull the status of update AddOn request from the newly modified server
	err := pullStatusOfUpdateAddOn(httpClient, &httpReq, "some-req-id", fromStatus)

	if err != nil {
		return err
	}

	return nil
}

func TestPrepareToGetLPAPIAccessToken(t *testing.T) {
	t.Parallel()
	t.Run("Succesfully generated a certificate with service key", func(t *testing.T) {
		err := prepareToGetLPAPIAccessToken(&updateAddOnConfig, &httpClientAT, &servKey)

		assert.Equal(t, nil, err)
		assert.Equal(t, "https://some-url.com", servKey.Url)
		assert.Equal(t, "some-client-id", servKey.Uaa.ClientId)
		assert.Equal(t, "https://some-cert-url.com", servKey.Uaa.CertUrl)
		assert.Equal(t, "-----BEGIN CERTIFICATE-----\nMIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw\nDgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow\nEjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d\n7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B\n5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr\nBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1\nNDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l\nWf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc\n6MF9+Yw1Yy0t\n-----END CERTIFICATE-----", servKey.Uaa.Certificate)
		assert.Equal(t, "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49\nAwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q\nEKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==\n-----END EC PRIVATE KEY-----", servKey.Uaa.Key)
	})

	t.Run("Error happened generating certificate with service key", func(t *testing.T) {
		lPAPIServiceKey_temp := `{
			"url": "https://some-url.com",
			"uaa": {
			    "clientid": "some-client-id",
			    "url": "https://some-uaa-url.com",
			    "certificate": "-----BEGIN CERTIFICATE-----\nsome-cert\n-----END CERTIFICATE-----",
			    "certurl": "https://some-cert-url.com",
			    "key": "-----BEGIN EC PRIVATE KEY-----\nsome-key\n-----END EC PRIVATE KEY-----"
			},
			"vendor": "SAP"
		    }`

		updateAddOnConfig_temp := updateAddOnConfig
		updateAddOnConfig_temp.LandscapePortalAPIServiceKey = lPAPIServiceKey_temp
		tempHttpClientAT := http.Client{}
		servKey_temp := serviceKey{}

		err := prepareToGetLPAPIAccessToken(&updateAddOnConfig_temp, &tempHttpClientAT, &servKey_temp)

		assert.NotEqual(t, nil, err)
	})
}

func TestGetLPAPIAccessToken(t *testing.T) {
	json.Unmarshal([]byte(lPAPIServiceKey), &servKey)
	req := mockGetLPAPIAccessTokenReq(servKey)
	servKey_temp := servKey

	t.Parallel()
	t.Run("Successfully return an access token", func(t *testing.T) {
		var accessToken string

		mockedServer := mockServer(&req, resp_200, wantedAccessToken)

		servKey_temp.Uaa.CertUrl = mockedServer.URL

		accessToken, err := getLPAPIAccessToken(httpClientAT, servKey_temp)

		assert.Equal(t, wantedAccessToken.AccessToken, accessToken)
		assert.Equal(t, nil, err)
	})

	t.Run("Non-200 status code returned when getting an access token", func(t *testing.T) {
		var accessToken string

		mockedServer := mockServer(&req, resp_400, nil)

		servKey_temp.Uaa.CertUrl = mockedServer.URL

		accessToken, err := getLPAPIAccessToken(httpClientAT, servKey_temp)

		expectedErr := fmt.Errorf("Unexpected response status 400 Bad Request received when getting access token of LP API.")
		accessToken, err = getLPAPIAccessToken(httpClientAT, servKey_temp)

		assert.Equal(t, "", accessToken)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Error returned when returned when getting an access token", func(t *testing.T) {
		var accessToken string

		accessToken, err := getLPAPIAccessToken(httpClientAT, servKey)

		assert.Equal(t, "", accessToken)
		assert.ErrorContains(t, err, "no such host")
	})
}

func TestGetSystemBySystemNumber(t *testing.T) {
	json.Unmarshal([]byte(lPAPIServiceKey), &servKey)
	accessToken, servKey_temp := mockGetLPAPIAccessToken(httpClientAT, servKey)
	rawURL := servKey.Url + "/api/v1.0/systems/:" + updateAddOnConfig.AbapSystemNumber

	req := mockedClientReq{
		url:    parseRawURL(rawURL),
		method: http.MethodGet,
		header: map[string][]string{
			"Authorization": {"Bearer" + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
	}

	wantedSystem := systemEntity{
		SystemId:     "some-system-id",
		SystemNumber: "some-system-number",
		ZoneId:       "some-zone-id",
	}

	t.Parallel()
	t.Run("Successfully return the id of a system", func(t *testing.T) {
		var systemId string

		mockedServer := mockServer(&req, resp_200, wantedSystem)

		servKey_temp.Url = mockedServer.URL

		err := getSystemBySystemNumber(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, &systemId)

		assert.Equal(t, systemId, wantedSystem.SystemId)
		assert.Equal(t, nil, err)
	})

	t.Run("Non-200 status code returned when getting the id of a system", func(t *testing.T) {
		var systemId string

		mockedServer := mockServer(&req, resp_400, nil)

		servKey_temp.Url = mockedServer.URL

		err := getSystemBySystemNumber(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, &systemId)

		expectedErr := fmt.Errorf("Unexpected response status 400 Bad Request received when getting system with systemNumber %v.", updateAddOnConfig.AbapSystemNumber)

		assert.Equal(t, "", systemId)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Error returned when returned when getting the id of a system", func(t *testing.T) {
		var systemId string

		err := getSystemBySystemNumber(&updateAddOnConfig, httpClient, httpClientAT, servKey, &systemId)

		assert.Equal(t, "", systemId)
		assert.ErrorContains(t, err, "no such host")
	})
}

func TestGetStatusOfUpdateAddOn(t *testing.T) {
	json.Unmarshal([]byte(lPAPIServiceKey), &servKey)
	reqId := "some-req-id"
	accessToken, servKey_temp := mockGetLPAPIAccessToken(httpClientAT, servKey)
	rawURL := servKey.Url + "/api/v1.0/requests/:" + reqId

	req := mockedClientReq{
		url:    parseRawURL(rawURL),
		method: http.MethodGet,
		header: map[string][]string{
			"Authorization": {"Bearer" + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
	}

	wantedRequest := reqEntity{
		RequestId: "some-req-id",
		ZoneId:    "some-zone-id",
		Status:    "C",
		SystemId:  "some-system-id",
	}

	t.Parallel()
	t.Run("Successfully return the status of a request and store the query request for later use", func(t *testing.T) {
		var getStatusReq http.Request
		var status string
		mockedServer := mockServer(&req, resp_200, wantedRequest)

		servKey_temp.Url = mockedServer.URL

		err := getStatusOfUpdateAddOn(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, reqId, &status, &getStatusReq)

		assert.Equal(t, wantedRequest.Status, status)
		assert.Equal(t, nil, err)
		assert.Contains(t, mockedServer.URL, getStatusReq.URL.Host)
		assert.Equal(t, "/api/v1.0/requests/:some-req-id", getStatusReq.URL.Path)
		assert.Equal(t, req.method, getStatusReq.Method)
		assert.Equal(t, req.header, getStatusReq.Header)
	})

	t.Run("Non-200 status code returned when getting the id of a request", func(t *testing.T) {
		var getStatusReq http.Request
		var status string

		mockedServer := mockServer(&req, resp_400, nil)

		servKey_temp.Url = mockedServer.URL

		expectedErr := fmt.Errorf("Unexpected response status 400 Bad Request received when getting status of request with id %v.", reqId)
		err := getStatusOfUpdateAddOn(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, reqId, &status, &getStatusReq)

		assert.Equal(t, "", status)
		assert.Equal(t, expectedErr, err)
		assert.Contains(t, mockedServer.URL, getStatusReq.URL.Host)
		assert.Equal(t, "/api/v1.0/requests/:some-req-id", getStatusReq.URL.Path)
		assert.Equal(t, req.method, getStatusReq.Method)
		assert.Equal(t, req.header, getStatusReq.Header)
	})

	t.Run("Error returned when returned when getting the id of a request", func(t *testing.T) {
		var getStatusReq http.Request
		var status string

		err := getStatusOfUpdateAddOn(&updateAddOnConfig, httpClient, httpClientAT, servKey, reqId, &status, &getStatusReq)

		assert.Equal(t, "", status)
		assert.Equal(t, http.Request{}, getStatusReq)
		assert.ErrorContains(t, err, "no such host")
	})
}

func TestUpdateAddOn(t *testing.T) {
	json.Unmarshal([]byte(lPAPIServiceKey), &servKey)
	systemId := "some-system-id"
	accessToken, servKey_temp := mockGetLPAPIAccessToken(httpClientAT, servKey)
	rawURL := servKey.Url + "/api/v1.0/systems/:" + systemId + "/deployProduct"
	updateAddOnReqBody := updateAddOnReq{
		ProductName:    "some-product-name",
		ProductVersion: "some-product-version",
	}
	httpReq := encodeReqBody[updateAddOnReq](updateAddOnReqBody, rawURL)

	req := mockedClientReq{
		url:    parseRawURL(rawURL),
		method: http.MethodPost,
		header: map[string][]string{
			"Authorization": {"Bearer" + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
		body: httpReq.Body,
	}

	wantedRequest := reqEntity{
		RequestId: "some-req-id",
		ZoneId:    "some-zone-id",
		Status:    "S",
		SystemId:  "some-system-id",
	}

	t.Parallel()
	t.Run("Successfully update AddOn in the system", func(t *testing.T) {
		var reqId string

		mockedServer := mockServer(&req, resp_200, wantedRequest)

		servKey_temp.Url = mockedServer.URL

		err := updateAddOn(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, systemId, &reqId)

		assert.Equal(t, wantedRequest.RequestId, reqId)
		assert.Equal(t, nil, err)
	})

	t.Run("Non-200 status code returned when updating AddOn in a system", func(t *testing.T) {
		var reqId string

		mockedServer := mockServer(&req, resp_400, nil)

		servKey_temp.Url = mockedServer.URL

		expectedErr := fmt.Errorf("Unexpected response status 400 Bad Request received when updating AddOn in system with systemId %v.", systemId)
		err := updateAddOn(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, systemId, &reqId)

		assert.Equal(t, "", reqId)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Error returned when returned when updating AddOn in a system", func(t *testing.T) {
		var reqId string

		err := updateAddOn(&updateAddOnConfig, httpClient, httpClientAT, servKey, systemId, &reqId)

		assert.Equal(t, "", reqId)
		assert.ErrorContains(t, err, "no such host")
	})
}

func TestCancelUpdateAddOn(t *testing.T) {
	json.Unmarshal([]byte(lPAPIServiceKey), &servKey)
	reqId := "some-req-id"
	accessToken, servKey_temp := mockGetLPAPIAccessToken(httpClientAT, servKey)
	rawURL := servKey.Url + "/api/v1.0/requests/" + reqId

	req := mockedClientReq{
		url:    parseRawURL(rawURL),
		method: http.MethodPost,
		header: map[string][]string{
			"Authorization": {"Bearer" + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
	}

	t.Parallel()
	t.Run("Successfully cancel update AddOn request", func(t *testing.T) {
		mockedServer := mockServer(&req, resp_204, nil)

		servKey_temp.Url = mockedServer.URL

		err := cancelUpdateAddOn(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, reqId)

		assert.Equal(t, nil, err)
	})

	t.Run("Non-204 status code returned when canceling update AddOn request", func(t *testing.T) {
		mockedServer := mockServer(&req, resp_400, nil)

		servKey_temp.Url = mockedServer.URL

		expectedErr := fmt.Errorf("Unexpected response status 400 Bad Request received when canceling addon update request with id %v.", reqId)
		err := cancelUpdateAddOn(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, reqId)

		assert.Equal(t, expectedErr, err)
	})

	t.Run("Error returned when returned when canceling update AddOn request", func(t *testing.T) {
		err := cancelUpdateAddOn(&updateAddOnConfig, httpClient, httpClientAT, servKey, reqId)

		assert.ErrorContains(t, err, "no such host")
	})
}

func TestRespondToUpdateAddOnFinalStatus(t *testing.T) {
	t.Run("Cancel update AddOn request if it failed", func(t *testing.T) {
		json.Unmarshal([]byte(lPAPIServiceKey), &servKey)
		reqId := "some-req-id"
		accessToken, servKey_temp := mockGetLPAPIAccessToken(httpClientAT, servKey)
		rawURL := servKey.Url + "/api/v1.0/requests/" + reqId
		status := "E"

		req := mockedClientReq{
			url:    parseRawURL(rawURL),
			method: http.MethodPost,
			header: map[string][]string{
				"Authorization": {"Bearer" + accessToken},
				"Content-Type":  {"application/json"},
				"Accept":        {"application/json"},
			},
		}

		mockedServer := mockServer(&req, resp_204, nil)

		servKey_temp.Url = mockedServer.URL

		expectedErr := fmt.Errorf("AddOn update failed.")
		err := respondToUpdateAddOnFinalStatus(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, reqId, status)

		assert.Equal(t, expectedErr, err)
	})
}

func TestRunAbapLandcapePortalUpdateAddOnProduct(t *testing.T) {
	// declare variables
	var systemId, reqId, reqStatus string
	var getStatusReq http.Request
	// var err error

	// mock server for getLPAPIAccessToken to get LP API access token
	json.Unmarshal([]byte(lPAPIServiceKey), &servKey)
	req1 := mockGetLPAPIAccessTokenReq(servKey)

	servKey_temp := servKey

	mockedServer := mockServer(&req1, resp_200, wantedAccessToken)
	servKey_temp.Uaa.CertUrl = mockedServer.URL

	accessToken, err1 := getLPAPIAccessToken(httpClientAT, servKey_temp)

	// mock server for getSystemBySystemNumber and get the system id
	wantedSystem := systemEntity{
		SystemId:     "some-system-id",
		SystemNumber: "some-system-number",
		ZoneId:       "some-zone-id",
	}

	req2 := mockedClientReq{
		url:    parseRawURL(servKey.Url + "/api/v1.0/systems/:" + updateAddOnConfig.AbapSystemNumber),
		method: http.MethodGet,
		header: map[string][]string{
			"Authorization": {"Bearer" + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
	}

	mockedServer = mockServer(&req2, resp_200, wantedSystem)

	servKey_temp.Url = mockedServer.URL

	err2 := getSystemBySystemNumber(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, &systemId)

	// mock server for updateAddOn and excute it
	updateAddOnReqBody := updateAddOnReq{
		ProductName:    "some-product-name",
		ProductVersion: "some-product-version",
	}
	httpReq := encodeReqBody[updateAddOnReq](updateAddOnReqBody, servKey.Url+"/api/v1.0/systems/:"+systemId+"/deployProduct")

	req3 := mockedClientReq{
		url:    parseRawURL(servKey.Url + "/api/v1.0/systems/:" + systemId + "/deployProduct"),
		method: http.MethodPost,
		header: map[string][]string{
			"Authorization": {"Bearer" + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
		body: httpReq.Body,
	}

	wantedRequest := reqEntity{
		RequestId: "some-req-id",
		ZoneId:    "some-zone-id",
		Status:    "S",
		SystemId:  "some-system-id",
	}

	mockedServer = mockServer(&req3, resp_200, wantedRequest)

	servKey_temp.Url = mockedServer.URL

	err3 := updateAddOn(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, systemId, &reqId)

	// mock server for getStatusOfUpdateAddOn and execute it
	req4 := mockedClientReq{
		url:    parseRawURL(servKey.Url + "/api/v1.0/requests/:" + reqId),
		method: http.MethodGet,
		header: map[string][]string{
			"Authorization": {"Bearer" + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
	}

	wantedRequest = reqEntity{
		RequestId: "some-req-id",
		ZoneId:    "some-zone-id",
		Status:    "C",
		SystemId:  "some-system-id",
	}

	mockedServer = mockServer(&req4, resp_200, wantedRequest)

	servKey_temp.Url = mockedServer.URL

	err4 := getStatusOfUpdateAddOn(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, reqId, &reqStatus, &getStatusReq)

	t.Parallel()
	t.Run("Successfully update AddOn", func(t *testing.T) {
		finalStatus := "C"

		// mock pulling status of update AddOn request until final status reaches
		mockedClientReq_temp := req4
		mockedClientReq_temp.url = parseRawURL(mockedServer.URL + "/api/v1.0/requests/:" + reqId)
		err5 := keepPullingUntilFinalStatusReached(&reqStatus, finalStatus, mockedServer, mockedClientReq_temp)

		// mock respond to completed update AddOn
		err6 := respondToUpdateAddOnFinalStatus(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, reqId, reqStatus)

		// assertions
		assert.Equal(t, "some-access-token", accessToken)
		assert.Equal(t, "some-system-id", systemId)
		assert.Equal(t, "some-req-id", reqId)
		assert.Equal(t, finalStatus, reqStatus)
		assert.Contains(t, mockedServer.URL, getStatusReq.URL.Host)
		assert.Equal(t, "/api/v1.0/requests/:some-req-id", getStatusReq.URL.Path)

		assert.Equal(t, nil, err1)
		assert.Equal(t, nil, err2)
		assert.Equal(t, nil, err3)
		assert.Equal(t, nil, err4)
		assert.Equal(t, nil, err5)
		assert.Equal(t, nil, err6)
		// assert.Equal(t, nil, err)
	})

	t.Run("Update AddOn request is aborted", func(t *testing.T) {
		finalStatus := "X"

		// mock pulling status of update AddOn request until final status reaches
		mockedClientReq_temp := req4
		mockedClientReq_temp.url = parseRawURL(mockedServer.URL + "/api/v1.0/requests/:" + reqId)
		err5 := keepPullingUntilFinalStatusReached(&reqStatus, finalStatus, mockedServer, mockedClientReq_temp)

		// mock respond to abort update AddOn
		expectedErr6 := fmt.Errorf("AddOn update is aborted.")
		err6 := respondToUpdateAddOnFinalStatus(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, reqId, reqStatus)

		// assertions
		assert.Equal(t, "some-access-token", accessToken)
		assert.Equal(t, "some-system-id", systemId)
		assert.Equal(t, "some-req-id", reqId)
		assert.Equal(t, finalStatus, reqStatus)
		assert.Contains(t, mockedServer.URL, getStatusReq.URL.Host)
		assert.Equal(t, "/api/v1.0/requests/:some-req-id", getStatusReq.URL.Path)

		assert.Equal(t, nil, err1)
		assert.Equal(t, nil, err2)
		assert.Equal(t, nil, err3)
		assert.Equal(t, nil, err4)
		assert.Equal(t, nil, err5)
		assert.Equal(t, expectedErr6, err6)
	})

	t.Run("Failed to update AddOn", func(t *testing.T) {
		finalStatus := "E"

		// mock pulling status of update AddOn request until final status reaches
		mockedClientReq_temp := req4
		mockedClientReq_temp.url = parseRawURL(mockedServer.URL + "/api/v1.0/requests/:" + reqId)
		err5 := keepPullingUntilFinalStatusReached(&reqStatus, finalStatus, mockedServer, mockedClientReq_temp)

		// mock respond to cancel update AddOn
		req5 := mockedClientReq{
			url:    parseRawURL(servKey.Url + "/api/v1.0/requests/" + reqId),
			method: http.MethodDelete,
			header: map[string][]string{
				"Authorization": {"Bearer" + accessToken},
				"Content-Type":  {"application/json"},
				"Accept":        {"application/json"},
			},
		}

		mockedServer_cancelUpdateAddOn := mockServer(&req5, resp_204, nil)

		servKey_temp.Url = mockedServer_cancelUpdateAddOn.URL

		expectedErr6 := fmt.Errorf("AddOn update failed.")
		err6 := respondToUpdateAddOnFinalStatus(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, reqId, reqStatus)

		// assertions
		assert.Equal(t, "some-access-token", accessToken)
		assert.Equal(t, "some-system-id", systemId)
		assert.Equal(t, "some-req-id", reqId)
		assert.Equal(t, finalStatus, reqStatus)
		assert.Contains(t, mockedServer.URL, getStatusReq.URL.Host)
		assert.Equal(t, "/api/v1.0/requests/:some-req-id", getStatusReq.URL.Path)

		assert.Equal(t, nil, err1)
		assert.Equal(t, nil, err2)
		assert.Equal(t, nil, err3)
		assert.Equal(t, nil, err4)
		assert.Equal(t, nil, err5)
		assert.Equal(t, expectedErr6, err6)
	})

	t.Run("Non-200 status code was returned", func(t *testing.T) {
		finalStatus := "E"

		// mock pulling status of update AddOn request until final status reaches
		mockedClientReq_temp := req4
		mockedClientReq_temp.url = parseRawURL(mockedServer.URL + "/api/v1.0/requests/:" + reqId)
		err5 := keepPullingUntilFinalStatusReached(&reqStatus, finalStatus, mockedServer, mockedClientReq_temp)

		// mock respond to cancel update AddOn
		req5 := mockedClientReq{
			url:    parseRawURL(servKey.Url + "/api/v1.0/requests/" + reqId),
			method: http.MethodDelete,
			header: map[string][]string{
				"Authorization": {"Bearer" + accessToken},
				"Content-Type":  {"application/json"},
				"Accept":        {"application/json"},
			},
		}

		mockedServer_cancelUpdateAddOn := mockServer(&req5, resp_400, nil)

		servKey_temp.Url = mockedServer_cancelUpdateAddOn.URL

		err6 := respondToUpdateAddOnFinalStatus(&updateAddOnConfig, httpClient, httpClientAT, servKey_temp, reqId, reqStatus)

		// assertions
		assert.Equal(t, "some-access-token", accessToken)
		assert.Equal(t, "some-system-id", systemId)
		assert.Equal(t, "some-req-id", reqId)
		assert.Equal(t, finalStatus, reqStatus)
		assert.Contains(t, mockedServer.URL, getStatusReq.URL.Host)
		assert.Equal(t, "/api/v1.0/requests/:some-req-id", getStatusReq.URL.Path)

		assert.Equal(t, nil, err1)
		assert.Equal(t, nil, err2)
		assert.Equal(t, nil, err3)
		assert.Equal(t, nil, err4)
		assert.Equal(t, nil, err5)
		assert.Equal(t, err6.Error(), "Failed to cancel addon update. Error: Unexpected response status 400 Bad Request received when canceling addon update request with id some-req-id.")
	})

	t.Run("Other error returned", func(t *testing.T) {
		finalStatus := "E"

		// mock pulling status of update AddOn request until final status reaches
		mockedClientReq_temp := req4
		mockedClientReq_temp.url = parseRawURL(mockedServer.URL + "/api/v1.0/requests/:" + reqId)
		err5 := keepPullingUntilFinalStatusReached(&reqStatus, finalStatus, mockedServer, mockedClientReq_temp)
		err6 := respondToUpdateAddOnFinalStatus(&updateAddOnConfig, httpClient, httpClientAT, servKey, reqId, reqStatus)

		// assertions
		assert.Equal(t, "some-access-token", accessToken)
		assert.Equal(t, "some-system-id", systemId)
		assert.Equal(t, "some-req-id", reqId)
		assert.Equal(t, finalStatus, reqStatus)
		assert.Contains(t, mockedServer.URL, getStatusReq.URL.Host)
		assert.Equal(t, "/api/v1.0/requests/:some-req-id", getStatusReq.URL.Path)

		assert.Equal(t, nil, err1)
		assert.Equal(t, nil, err2)
		assert.Equal(t, nil, err3)
		assert.Equal(t, nil, err4)
		assert.Equal(t, nil, err5)
		assert.Contains(t, err6.Error(), "no such host")
	})
}
