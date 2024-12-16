//go:build unit

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	resBodyJSON_token = `{"access_token": "some-access-token", "token_type": "bearer", "expires_in": 86400, "scope": "some-scope"}`
	resBodyJSON_sys   = `{"SystemId": "some-system-id", "SystemNumber": "some-system-number", "zone_id": "some-zone-id"}`
	resBodyJSON_req_S = `{"RequestId": "some-request-id","zone_id": "some-zone-id", "Status": "S", "SystemId": "some-system-id"}`
	resBodyJSON_req_I = `{"RequestId": "some-request-id","zone_id": "some-zone-id", "Status": "I", "SystemId": "some-system-id"}`
	resBodyJSON_req_C = `{"RequestId": "some-request-id","zone_id": "some-zone-id", "Status": "C", "SystemId": "some-system-id"}`
	resBodyJSON_req_E = `{"RequestId": "some-request-id","zone_id": "some-zone-id", "Status": "E", "SystemId": "some-system-id"}`
	resBodyJSON_req_X = `{"RequestId": "some-request-id","zone_id": "some-zone-id", "Status": "X", "SystemId": "some-system-id"}`
)

type mockClient struct {
	DoFunc func(*http.Request) (*http.Response, error)
}

var GetDoFunc func(req *http.Request) (*http.Response, error)

var testUaa = uaa{
	CertUrl:     "https://some-cert-url.com",
	ClientId:    "some-client-id",
	Certificate: "-----BEGIN CERTIFICATE-----\nMIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw\nDgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow\nEjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d\n7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B\n5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr\nBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1\nNDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l\nWf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc\n6MF9+Yw1Yy0t\n-----END CERTIFICATE-----",
	Key:         "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49\nAwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q\nEKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==\n-----END EC PRIVATE KEY-----",
}
var mockServKey = serviceKey{
	Url: "https://some-url.com",
	Uaa: testUaa,
}

var mockServiceKeyJSON = `{
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

var mockUpdateAddOnConfig = abapLandscapePortalUpdateAddOnProductOptions{
	LandscapePortalAPIServiceKey: mockServiceKeyJSON,
	AbapSystemNumber:             "some-system-number",
	AddonDescriptorFileName:      "addon.yml",
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

func init() {
	client = &mockClient{}
	clientToken = &mockClient{}
}

func TestParseServiceKeyAndPrepareAccessTokenHttpClient(t *testing.T) {
	t.Run("Successfully parsed service key", func(t *testing.T) {
		var testServKey serviceKey
		clientParseServKey := clientToken

		err := parseServiceKeyAndPrepareAccessTokenHttpClient(mockUpdateAddOnConfig.LandscapePortalAPIServiceKey, &clientParseServKey, &testServKey)

		assert.Equal(t, nil, err)
		assert.Equal(t, "https://some-url.com", testServKey.Url)
		assert.Equal(t, "some-client-id", testServKey.Uaa.ClientId)
		assert.Equal(t, "https://some-cert-url.com", testServKey.Uaa.CertUrl)
	})
}

func TestGetLPAPIAccessToken(t *testing.T) {
	t.Run("Successfully got LP API access token", func(t *testing.T) {
		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			resBodyReader := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
			return &http.Response{
				StatusCode: 200,
				Body:       resBodyReader,
			}, nil
		}

		res, err := getLPAPIAccessToken(clientToken, mockServKey)

		assert.Equal(t, "some-access-token", res)
		assert.Equal(t, nil, err)
	})

	t.Run("Failed to get LP API access token", func(t *testing.T) {
		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("Failed to get access token.")
		}
		res, err := getLPAPIAccessToken(clientToken, mockServKey)

		assert.Equal(t, "", res)
		assert.Equal(t, fmt.Errorf("Failed to get access token."), err)
	})
}

func TestGetSystemBySystemNumber(t *testing.T) {
	reqUrl_token := mockServKey.Uaa.CertUrl + "/oauth/token"
	reqUrl_sys := mockServKey.Url + "/api/systems/" + mockUpdateAddOnConfig.AbapSystemNumber

	t.Run("Successfully got ABAP system", func(t *testing.T) {
		var testSysId string

		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == reqUrl_token {
				resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_sys {
				resBodyReader_sys := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_sys)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_sys,
				}, nil
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		err := getSystemBySystemNumber(&mockUpdateAddOnConfig, client, clientToken, mockServKey, &testSysId)

		assert.Equal(t, "some-system-id", testSysId)
		assert.Equal(t, nil, err)
	})

	t.Run("Failed to get ABAP system", func(t *testing.T) {
		var testSysId string

		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == reqUrl_token {
				resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_sys {
				return nil, fmt.Errorf("Failed to get ABAP system.")
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		err := getSystemBySystemNumber(&mockUpdateAddOnConfig, client, clientToken, mockServKey, &testSysId)

		assert.Equal(t, "", testSysId)
		assert.Equal(t, fmt.Errorf("Failed to get ABAP system."), err)
	})
}

func TestUpdateAddOn(t *testing.T) {
	testSysId := "some-system-id"
	reqUrl_token := mockServKey.Uaa.CertUrl + "/oauth/token"
	reqUrl_update := mockServKey.Url + "/api/systems/" + testSysId + "/deployProduct"

	t.Run("Successfully updated addon", func(t *testing.T) {
		// write addon.yml
		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		addonYML := `addonProduct: some-addon-product
addonVersion: 1.0.0
`
		addonYMLBytes := []byte(addonYML)
		os.WriteFile("addon.yml", addonYMLBytes, 0644)

		// mock Do func
		var testReqId string

		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == reqUrl_token {
				resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_update {
				resBodyReader_req_S := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_req_S)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_req_S,
				}, nil
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		err := updateAddOn(mockUpdateAddOnConfig.AddonDescriptorFileName, client, clientToken, mockServKey, testSysId, &testReqId)

		assert.Equal(t, "some-request-id", testReqId)
		assert.Equal(t, nil, err)
	})

	t.Run("Failed to update addon", func(t *testing.T) {
		// write addon.yml
		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		addonYML := `addonProduct: some-addon-product
addonVersion: 1.0.0
`
		addonYMLBytes := []byte(addonYML)
		os.WriteFile("addon.yml", addonYMLBytes, 0644)

		// mock Do func
		var testReqId string

		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == reqUrl_token {
				resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_update {
				return nil, fmt.Errorf("Failed to update addon.")
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		err := updateAddOn(mockUpdateAddOnConfig.AddonDescriptorFileName, client, clientToken, mockServKey, testSysId, &testReqId)

		assert.Equal(t, "", testReqId)
		assert.Equal(t, fmt.Errorf("Failed to update addon."), err)
	})
}

func TestPollStatusOfUpdateAddOn(t *testing.T) {
	var testReq http.Request

	testReqId := "some-request-id"
	reqUrl_token := mockServKey.Uaa.CertUrl + "/oauth/token"
	reqUrl_pollAndCancel := mockServKey.Url + "/api/requests/" + testReqId

	t.Run("Successfully polled request status", func(t *testing.T) {
		var testStatus string

		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == reqUrl_token {
				resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_pollAndCancel {
				resBodyReader_pollStatus := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_req_I)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_pollStatus,
				}, nil
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		err1 := prepareGetStatusHttpRequest(clientToken, mockServKey, testReqId, &testReq)
		err2 := pollStatusOfUpdateAddOn(client, &testReq, testReqId, &testStatus)

		assert.Equal(t, "I", testStatus)
		assert.Equal(t, nil, err1)
		assert.Equal(t, nil, err2)
	})

	t.Run("Failed to poll request status", func(t *testing.T) {
		var testStatus string

		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == reqUrl_token {
				resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_pollAndCancel {
				return nil, fmt.Errorf("Failed to poll status.")
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		err1 := prepareGetStatusHttpRequest(clientToken, mockServKey, testReqId, &testReq)
		err2 := pollStatusOfUpdateAddOn(client, &testReq, testReqId, &testStatus)

		assert.Equal(t, "", testStatus)
		assert.Equal(t, nil, err1)
		assert.Equal(t, fmt.Errorf("Failed to poll status."), err2)
	})
}

func TestCancelUpdateAddOn(t *testing.T) {
	testReqId := "some-request-id"
	reqUrl_token := mockServKey.Uaa.CertUrl + "/oauth/token"
	reqUrl_pollAndCancel := mockServKey.Url + "/api/requests/" + testReqId

	t.Run("Successfully canceled addon update", func(t *testing.T) {
		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
			if req.URL.String() == reqUrl_token {
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_pollAndCancel {
				resBodyReader_cancelUpdate := io.NopCloser(nil)
				return &http.Response{
					StatusCode: 204,
					Body:       resBodyReader_cancelUpdate,
				}, nil
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		err := cancelUpdateAddOn(client, clientToken, mockServKey, testReqId)

		assert.Equal(t, nil, err)
	})

	t.Run("Failed to cancel addon update", func(t *testing.T) {
		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == reqUrl_token {
				resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_pollAndCancel {
				return nil, fmt.Errorf("Failed to cancel addon update.")
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		err := cancelUpdateAddOn(client, clientToken, mockServKey, testReqId)

		assert.Equal(t, fmt.Errorf("Failed to cancel addon update."), err)
	})
}

func TestRunAbapLandscapePortalUpdateAddOnProduct(t *testing.T) {
	reqUrl_token := mockServKey.Uaa.CertUrl + "/oauth/token"
	reqUrl_sys := mockServKey.Url + "/api/systems/" + mockUpdateAddOnConfig.AbapSystemNumber
	reqUrl_update := mockServKey.Url + "/api/systems/" + "some-system-id" + "/deployProduct"
	reqUrl_pollAndCancel := mockServKey.Url + "/api/requests/" + "some-request-id"

	t.Run("Successfully ran update addon in ABAP system", func(t *testing.T) {
		// write addon.yml
		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		addonYML := `addonProduct: some-addon-product
addonVersion: 1.0.0
`
		addonYMLBytes := []byte(addonYML)
		os.WriteFile("addon.yml", addonYMLBytes, 0644)

		// mock Do func
		maxRuntimeInMinute := time.Duration(1) * time.Minute
		pollIntervalInSecond := time.Duration(1) * time.Second

		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == reqUrl_token {
				resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_sys {
				resBodyReader_sys := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_sys)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_sys,
				}, nil
			}
			if req.URL.String() == reqUrl_update {
				resBodyReader_update := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_req_S)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_update,
				}, nil
			}

			if req.URL.String() == reqUrl_pollAndCancel {
				resBodyReader_pollStatus_C := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_req_C)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_pollStatus_C,
				}, nil
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		// execution and assertion
		err := runAbapLandscapePortalUpdateAddOnProduct(&mockUpdateAddOnConfig, client, clientToken, mockServKey, maxRuntimeInMinute, pollIntervalInSecond)

		assert.Equal(t, nil, err)
	})

	t.Run("Update addon ended in error", func(t *testing.T) {
		// write addon.yml
		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		addonYML := `addonProduct: some-addon-product
addonVersion: 1.0.0
`
		addonYMLBytes := []byte(addonYML)
		os.WriteFile("addon.yml", addonYMLBytes, 0644)

		// mock Do func
		maxRuntimeInMinute := time.Duration(1) * time.Minute
		pollIntervalInSecond := time.Duration(1) * time.Second

		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == reqUrl_token {
				resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_sys {
				resBodyReader_sys := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_sys)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_sys,
				}, nil
			}
			if req.URL.String() == reqUrl_update {
				resBodyReader_update := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_req_S)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_update,
				}, nil
			}

			if req.URL.String() == reqUrl_pollAndCancel && req.Method == "GET" {
				resBodyReader_pollStatus_E := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_req_E)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_pollStatus_E,
				}, nil
			}

			if req.URL.String() == reqUrl_pollAndCancel && req.Method == "DELETE" {
				resBodyReader_cancelUpdate := io.NopCloser(nil)
				return &http.Response{
					StatusCode: 204,
					Body:       resBodyReader_cancelUpdate,
				}, nil
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		// execution and assertion
		expectedErr1 := fmt.Errorf("Addon update failed.\n")
		expectedErr2 := fmt.Errorf("The final status of addon update is E. Error: %v\n", expectedErr1)

		err := runAbapLandscapePortalUpdateAddOnProduct(&mockUpdateAddOnConfig, client, clientToken, mockServKey, maxRuntimeInMinute, pollIntervalInSecond)

		assert.Equal(t, expectedErr2, err)
	})

	t.Run("Update addon was aborted", func(t *testing.T) {
		// write addon.yml
		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		addonYML := `addonProduct: some-addon-product
addonVersion: 1.0.0
`
		addonYMLBytes := []byte(addonYML)
		os.WriteFile("addon.yml", addonYMLBytes, 0644)

		// mock Do func
		maxRuntimeInMinute := time.Duration(1) * time.Minute
		pollIntervalInSecond := time.Duration(1) * time.Second

		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == reqUrl_token {
				resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_sys {
				resBodyReader_sys := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_sys)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_sys,
				}, nil
			}
			if req.URL.String() == reqUrl_update {
				resBodyReader_update := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_req_S)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_update,
				}, nil
			}

			if req.URL.String() == reqUrl_pollAndCancel && req.Method == "GET" {
				resBodyReader_pollStatus_X := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_req_X)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_pollStatus_X,
				}, nil
			}

			if req.URL.String() == reqUrl_pollAndCancel && req.Method == "DELETE" {
				resBodyReader_cancelUpdate := io.NopCloser(nil)
				return &http.Response{
					StatusCode: 204,
					Body:       resBodyReader_cancelUpdate,
				}, nil
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		// execution and assertion
		expectedErr1 := fmt.Errorf("Addon update was aborted.\n")
		expectedErr2 := fmt.Errorf("The final status of addon update is X. Error: %v\n", expectedErr1)

		err := runAbapLandscapePortalUpdateAddOnProduct(&mockUpdateAddOnConfig, client, clientToken, mockServKey, maxRuntimeInMinute, pollIntervalInSecond)

		assert.Equal(t, expectedErr2, err)
	})

	t.Run("Update addon reached timeout", func(t *testing.T) {
		// write addon.yml
		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		addonYML := `addonProduct: some-addon-product
addonVersion: 1.0.0
`
		addonYMLBytes := []byte(addonYML)
		os.WriteFile("addon.yml", addonYMLBytes, 0644)

		// mock Do func
		maxRuntimeInMinute := time.Duration(3) * time.Second
		pollIntervalInSecond := time.Duration(1) * time.Second

		GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == reqUrl_token {
				resBodyReader_token := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_token)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_token,
				}, nil
			}

			if req.URL.String() == reqUrl_sys {
				resBodyReader_sys := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_sys)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_sys,
				}, nil
			}
			if req.URL.String() == reqUrl_update {
				resBodyReader_update := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_req_S)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_update,
				}, nil
			}

			if req.URL.String() == reqUrl_pollAndCancel && req.Method == "GET" {
				resBodyReader_pollStatus_I := io.NopCloser(bytes.NewReader([]byte(resBodyJSON_req_I)))
				return &http.Response{
					StatusCode: 200,
					Body:       resBodyReader_pollStatus_I,
				}, nil
			}

			return nil, fmt.Errorf("some-unknown-error")
		}

		// execution and assertion
		expectedErr1 := fmt.Errorf("Timed out: max runtime %v reached.", maxRuntimeInMinute)
		expectedErr2 := fmt.Errorf("Error occurred before a final status can be reached. Error: %v\n", expectedErr1)

		err := runAbapLandscapePortalUpdateAddOnProduct(&mockUpdateAddOnConfig, client, clientToken, mockServKey, maxRuntimeInMinute, pollIntervalInSecond)

		assert.Equal(t, expectedErr2, err)
	})
}
