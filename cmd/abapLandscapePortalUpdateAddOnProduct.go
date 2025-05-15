package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"golang.org/x/exp/slices"
)

const (
	StatusComplete       = "C"
	StatusError          = "E"
	StatusInProgress     = "I"
	StatusScheduled      = "S"
	StatusAborted        = "X"
	maxRuntimeInMinute   = time.Duration(120) * time.Minute
	pollIntervalInSecond = time.Duration(30) * time.Second
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type uaa struct {
	CertUrl     string `json:"certurl"`
	ClientId    string `json:"clientid"`
	Certificate string `json:"certificate"`
	Key         string `json:"key"`
}

type serviceKey struct {
	Url string `json:"url"`
	Uaa uaa    `json:"uaa"`
}

type accessTokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type systemEntity struct {
	SystemId     string `json:"SystemId"`
	SystemNumber string `json:"SystemNumber"`
	ZoneId       string `json:"zone_id"`
}

type reqEntity struct {
	RequestId string `json:"RequestId"`
	ZoneId    string `json:"zone_id"`
	Status    string `json:"Status"`
	SystemId  string `json:"SystemId"`
}

type updateAddOnReq struct {
	ProductName    string `json:"productName"`
	ProductVersion string `json:"productVersion"`
}

type updateAddOnResp struct {
	RequestId string `json:"requestId"`
	ZoneId    string `json:"zoneId"`
	Status    string `json:"status"`
	SystemId  string `json:"systemId"`
}

var client, clientToken httpClient
var servKey serviceKey

func abapLandscapePortalUpdateAddOnProduct(config abapLandscapePortalUpdateAddOnProductOptions, telemetryData *telemetry.CustomData) {
	client = &http.Client{}

	if prepareErr := parseServiceKeyAndPrepareAccessTokenHttpClient(config.LandscapePortalAPIServiceKey, &clientToken, &servKey); prepareErr != nil {
		err := fmt.Errorf("Failed to prepare credentials to get access token of LP API. Error: %v\n", prepareErr)
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runAbapLandscapePortalUpdateAddOnProduct(&config, client, clientToken, servKey, maxRuntimeInMinute, pollIntervalInSecond)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapLandscapePortalUpdateAddOnProduct(config *abapLandscapePortalUpdateAddOnProductOptions, client httpClient, clientToken httpClient, servKey serviceKey, maxRuntimeInMinute time.Duration, pollIntervalInSecond time.Duration) error {
	var systemId, reqId, reqStatus string
	var getStatusReq http.Request
	var err error

	// get system
	if getSystemErr := getSystemBySystemNumber(config, client, clientToken, servKey, &systemId); getSystemErr != nil {
		err = fmt.Errorf("Failed to get system with systemNumber %v. Error: %v\n", config.AbapSystemNumber, getSystemErr)
		return err
	}

	// update addon in the system
	if updateAddOnErr := updateAddOn(config.AddonDescriptorFileName, client, clientToken, servKey, systemId, &reqId); updateAddOnErr != nil {
		err = fmt.Errorf("Failed to update addon in the system with systemId %v. Error: %v\n", systemId, updateAddOnErr)
		return err
	}

	// prepare http request to poll status of addon update
	if prepareGetStatusHttpRequestErr := prepareGetStatusHttpRequest(clientToken, servKey, reqId, &getStatusReq); prepareGetStatusHttpRequestErr != nil {
		err = fmt.Errorf("Failed to prepare http request to poll status of addon update request %v. Error: %v\n", reqId, prepareGetStatusHttpRequestErr)
		return err
	}

	// keep polling request status until it reaches a final status or timeout
	if waitToBeFinishedErr := waitToBeFinished(maxRuntimeInMinute, pollIntervalInSecond, client, &getStatusReq, reqId, &reqStatus); waitToBeFinishedErr != nil {
		err = fmt.Errorf("Error occurred before a final status can be reached. Error: %v\n", waitToBeFinishedErr)
		return err
	}

	// respond to the final status of addon update
	if respondToUpdateAddOnFinalStatusErr := respondToUpdateAddOnFinalStatus(client, clientToken, servKey, reqId, reqStatus); respondToUpdateAddOnFinalStatusErr != nil {
		err = fmt.Errorf("The final status of addon update is %v. Error: %v\n", reqStatus, respondToUpdateAddOnFinalStatusErr)
		return err
	}

	return nil
}

// this function is used to parse service key JSON and prepare http client for access token
func parseServiceKeyAndPrepareAccessTokenHttpClient(servKeyJSON string, clientToken *httpClient, servKey *serviceKey) error {
	// parse the service key from JSON string to struct
	if parseServiceKeyErr := json.Unmarshal([]byte(servKeyJSON), servKey); parseServiceKeyErr != nil {
		return parseServiceKeyErr
	}

	// configure http client with certificate authorization for getLPAPIAccessToken
	certSource := servKey.Uaa.Certificate
	keySource := servKey.Uaa.Key

	certPem := strings.Replace(certSource, `\n`, "\n", -1)
	keyPem := strings.Replace(keySource, `\n`, "\n", -1)

	certificate, certErr := tls.X509KeyPair([]byte(certPem), []byte(keyPem))
	if certErr != nil {
		return certErr
	}

	*clientToken = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{certificate},
			},
		},
	}

	return nil
}

// this function is used to get access token of Landscape Portal API
func getLPAPIAccessToken(clientToken httpClient, servKey serviceKey) (string, error) {
	authRawURL := servKey.Uaa.CertUrl + "/oauth/token"

	// configure request body
	reqBody := url.Values{}
	reqBody.Set("grant_type", "client_credentials")
	reqBody.Set("client_id", servKey.Uaa.ClientId)

	encodedReqBody := reqBody.Encode()

	// generate http request and configure header
	req, reqErr := http.NewRequest(http.MethodPost, authRawURL, strings.NewReader(encodedReqBody))
	if reqErr != nil {
		return "", reqErr
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, getAccessTokenErr := clientToken.Do(req)
	if getAccessTokenErr != nil {
		return "", getAccessTokenErr
	}

	defer resp.Body.Close()

	// error case of response status code being non 200
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Unexpected response status %v received when getting access token of LP API.\n", resp.Status)
		return "", err
	}

	// read and parse response body
	respBody := accessTokenResp{}
	if parseRespBodyErr := parseRespBody[accessTokenResp](resp, &respBody); parseRespBodyErr != nil {
		return "", parseRespBodyErr
	}

	return respBody.AccessToken, nil
}

// this function is used to check the existence of integration test system
func getSystemBySystemNumber(config *abapLandscapePortalUpdateAddOnProductOptions, client httpClient, clientToken httpClient, servKey serviceKey, systemId *string) error {
	accessToken, getAccessTokenErr := getLPAPIAccessToken(clientToken, servKey)
	if getAccessTokenErr != nil {
		return getAccessTokenErr
	}

	// define the raw url of the request and parse it into required form used in http.Request
	getSystemRawURL := servKey.Url + "/api/systems/" + config.AbapSystemNumber
	getSystemURL, urlParseErr := url.Parse(getSystemRawURL)
	if urlParseErr != nil {
		return urlParseErr
	}

	req := http.Request{
		Method: http.MethodGet,
		URL:    getSystemURL,
		Header: map[string][]string{
			"Authorization": {"Bearer " + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
	}

	resp, getSystemErr := client.Do(&req)
	if getSystemErr != nil {
		return getSystemErr
	}

	defer resp.Body.Close()

	// error case of response status code being non 200
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Unexpected response status %v received when getting system with systemNumber %v.\n", resp.Status, config.AbapSystemNumber)
		return err
	}

	// read and parse response body
	respBody := systemEntity{}
	if parseRespBodyErr := parseRespBody[systemEntity](resp, &respBody); parseRespBodyErr != nil {
		return parseRespBodyErr
	}

	*systemId = respBody.SystemId

	fmt.Printf("Successfully got ABAP system with systemNumber %v and systemId %v.\n", respBody.SystemNumber, respBody.SystemId)
	return nil
}

// this function is used to define and maintain the request body of querying status of addon update request
func prepareGetStatusHttpRequest(clientToken httpClient, servKey serviceKey, reqId string, getStatusReq *http.Request) error {
	accessToken, getAccessTokenErr := getLPAPIAccessToken(clientToken, servKey)
	if getAccessTokenErr != nil {
		return getAccessTokenErr
	}

	// define the raw url of the request and parse it into required form used in http.Request
	getStatusRawURL := servKey.Url + "/api/requests/" + reqId
	getStatusURL, urlParseErr := url.Parse(getStatusRawURL)
	if urlParseErr != nil {
		return urlParseErr
	}

	req := http.Request{
		Method: http.MethodGet,
		URL:    getStatusURL,
		Header: map[string][]string{
			"Authorization": {"Bearer " + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
	}

	// store the req in the global variable for later usage
	*getStatusReq = req

	return nil
}

// this function is used to poll status of addon update request and maintain the status
func pollStatusOfUpdateAddOn(client httpClient, req *http.Request, reqId string, status *string) error {
	resp, getStatusErr := client.Do(req)
	if getStatusErr != nil {
		return getStatusErr
	}

	defer resp.Body.Close()

	// error case of response status code being non 200
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Unexpected response status %v received when polling status of request %v.\n", resp.Status, reqId)
		return err
	}

	// read and parse response body
	respBody := reqEntity{}
	if parseRespBodyErr := parseRespBody[reqEntity](resp, &respBody); parseRespBodyErr != nil {
		return parseRespBodyErr
	}

	*status = respBody.Status

	fmt.Printf("Successfully polled status %v of request %v.\n", respBody.Status, respBody.RequestId)
	return nil
}

// this function is used to update addon
func updateAddOn(addOnFileName string, client httpClient, clientToken httpClient, servKey serviceKey, systemId string, reqId *string) error {
	accessToken, getAccessTokenErr := getLPAPIAccessToken(clientToken, servKey)
	if getAccessTokenErr != nil {
		return getAccessTokenErr
	}

	// read productName and productVersion from addon.yml
	addOnDescriptor, readAddOnErr := abaputils.ReadAddonDescriptor(addOnFileName)
	if readAddOnErr != nil {
		return readAddOnErr
	}

	// define the raw url of the request and parse it into required form used in http.Request
	updateAddOnRawURL := servKey.Url + "/api/systems/" + systemId + "/deployProduct"

	// define the request body as a struct
	reqBody := updateAddOnReq{
		ProductName:    addOnDescriptor.AddonProduct,
		ProductVersion: addOnDescriptor.AddonVersionYAML,
	}

	// encode the request body to JSON
	var reqBuff bytes.Buffer
	json.NewEncoder(&reqBuff).Encode(reqBody)

	req, reqErr := http.NewRequest(http.MethodPost, updateAddOnRawURL, &reqBuff)
	if reqErr != nil {
		return reqErr
	}

	req.Header = map[string][]string{
		"Authorization": {"Bearer " + accessToken},
		"Content-Type":  {"application/json"},
		"Accept":        {"application/json"},
	}

	resp, updateAddOnErr := client.Do(req)
	if updateAddOnErr != nil {
		return updateAddOnErr
	}

	defer resp.Body.Close()

	// error case of response status code being non 200
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Unexpected response status %v received when updating addon in system with systemId %v.\n", resp.Status, systemId)
		return err
	}

	// read and parse response body
	respBody := updateAddOnResp{}
	if parseRespBodyErr := parseRespBody[updateAddOnResp](resp, &respBody); parseRespBodyErr != nil {
		return parseRespBodyErr
	}

	*reqId = respBody.RequestId

	fmt.Printf("Successfully triggered addon update in system with systemId %v, the returned request id is %v.\n", systemId, respBody.RequestId)
	return nil
}

// this function is used to cancel addon update
func cancelUpdateAddOn(client httpClient, clientToken httpClient, servKey serviceKey, reqId string) error {
	accessToken, getAccessTokenErr := getLPAPIAccessToken(clientToken, servKey)
	if getAccessTokenErr != nil {
		return getAccessTokenErr
	}

	// define the raw url of the request and parse it into required form used in http.Request
	cancelUpdateAddOnRawURL := servKey.Url + "/api/requests/" + reqId
	cancelUpdateAddOnURL, urlParseErr := url.Parse(cancelUpdateAddOnRawURL)
	if urlParseErr != nil {
		return urlParseErr
	}

	req := http.Request{
		Method: http.MethodDelete,
		URL:    cancelUpdateAddOnURL,
		Header: map[string][]string{
			"Authorization": {"Bearer " + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
	}

	resp, cancelUpdateAddOnErr := client.Do(&req)
	if cancelUpdateAddOnErr != nil {
		return cancelUpdateAddOnErr
	}

	defer resp.Body.Close()

	// error case of response status code being non 204
	if resp.StatusCode != http.StatusNoContent {
		err := fmt.Errorf("Unexpected response status %v received when canceling addon update request %v.\n", resp.Status, reqId)
		return err
	}

	fmt.Printf("Successfully canceled addon update request %v.\n", reqId)
	return nil
}

// this function is used to respond to a final status of addon update
func respondToUpdateAddOnFinalStatus(client httpClient, clientToken httpClient, servKey serviceKey, reqId string, status string) error {
	switch status {
	case StatusComplete:
		fmt.Println("Addon update succeeded.")
	case StatusError:
		fmt.Println("Addon update failed and will be canceled.")

		if cancelUpdateAddOnErr := cancelUpdateAddOn(client, clientToken, servKey, reqId); cancelUpdateAddOnErr != nil {
			err := fmt.Errorf("Failed to cancel addon update. Error: %v\n", cancelUpdateAddOnErr)
			return err
		}

		err := fmt.Errorf("Addon update failed.\n")
		return err

	case StatusAborted:
		fmt.Println("Addon update was aborted.")
		err := fmt.Errorf("Addon update was aborted.\n")
		return err
	}

	return nil
}

// this function is used to parse response body of http request
func parseRespBody[T comparable](resp *http.Response, respBody *T) error {
	respBodyRaw, readRespErr := io.ReadAll(resp.Body)
	if readRespErr != nil {
		return readRespErr
	}

	if decodeRespBodyErr := json.Unmarshal(respBodyRaw, &respBody); decodeRespBodyErr != nil {
		return decodeRespBodyErr
	}

	return nil
}

// this function is used to wait for a final status/timeout
func waitToBeFinished(maxRuntimeInMinute time.Duration, pollIntervalInSecond time.Duration, client httpClient, getStatusReq *http.Request, reqId string, reqStatus *string) error {
	timeout := time.After(maxRuntimeInMinute)
	ticker := time.Tick(pollIntervalInSecond)
	reqFinalStatus := []string{StatusComplete, StatusError, StatusAborted}
	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timed out: max runtime %v reached.", maxRuntimeInMinute)
		case <-ticker:
			if pollStatusOfUpdateAddOnErr := pollStatusOfUpdateAddOn(client, getStatusReq, reqId, reqStatus); pollStatusOfUpdateAddOnErr != nil {
				err := fmt.Errorf("Error happened when waiting for the addon update request %v to reach a final status. Error: %v\n", reqId, pollStatusOfUpdateAddOnErr)
				return err
			}
			if !slices.Contains(reqFinalStatus, *reqStatus) {
				fmt.Printf("Addon update request %v is still in progress, will poll the status in %v.\n", reqId, pollIntervalInSecond)
			} else {
				return nil
			}
		}
	}
}
