package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const (
	StatusComplete   = "C"
	StatusError      = "E"
	StatusInProgress = "I"
	StatusScheduled  = "S"
	StatusAborted    = "X"
)

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

func abapEnvironmentUpdateAddOnProduct(config abapEnvironmentUpdateAddOnProductOptions, telemetryData *telemetry.CustomData) {
	// define a http client
	client := http.Client{}
	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runAbapEnvironmentUpdateAddOnProduct(&config, client)
	if err != nil {
		log.Fatal("step execution failed")
	}
}

func runAbapEnvironmentUpdateAddOnProduct(config *abapEnvironmentUpdateAddOnProductOptions, client http.Client) error {
	// declare variables
	var systemId, reqId, reqStatus string
	var clientAT http.Client
	var servKey serviceKey
	var getStatusReq http.Request
	var err error

	// prepare to get access token
	prepareErr := prepareToGetLPAPIAccessToken(config, &clientAT, &servKey)
	if prepareErr != nil {
		err = fmt.Errorf("Failed to prepare credentials to get access token of LP API. Error: %v", prepareErr)
		return err
	}

	// get system
	getSystemErr := getSystemBySystemNumber(config, client, clientAT, servKey, &systemId)
	if getSystemErr != nil {
		err = fmt.Errorf("Failed to get system with systemNumber %v. Error: %v", config.AbapSystemNumber, getSystemErr)
		return err
	}

	// update addon in the system
	updateAddOnErr := updateAddOn(config, client, clientAT, servKey, systemId, &reqId)
	if updateAddOnErr != nil {
		err = fmt.Errorf("Failed to update addon in the system with systemNumber %v. Error: %v", config.AbapSystemNumber, updateAddOnErr)
		return err
	}

	// query status of request
	getStatusOfUpdateAddOnErr := getStatusOfUpdateAddOn(config, client, clientAT, servKey, reqId, &reqStatus, &getStatusReq)
	if getStatusOfUpdateAddOnErr != nil {
		err = fmt.Errorf("Failed to get status of update addon request with id %v. Error: %v", reqId, getStatusOfUpdateAddOnErr)
		return err
	}

	// keep pulling status of addon update request until it reaches a final status (C/E/X)
	for reqStatus == StatusInProgress || reqStatus == StatusScheduled {
		// pull status every 30s
		time.Sleep(30 * time.Second)
		pullStatusOfUpdateAddOnErr := pullStatusOfUpdateAddOn(client, &getStatusReq, reqId, &reqStatus)

		if pullStatusOfUpdateAddOnErr != nil {
			err = fmt.Errorf("Error happened when waiting for the update addon with request id %v to reach a final status. Error: %v", reqId, pullStatusOfUpdateAddOnErr)
			return err
		}
	}

	// respond to the final status of addon update
	respondToUpdateAddOnFinalStatusErr := respondToUpdateAddOnFinalStatus(config, client, clientAT, servKey, reqId, reqStatus)

	if respondToUpdateAddOnFinalStatusErr != nil {
		err = fmt.Errorf("Failed to respond to the final status %v of addon update. Error: %v", reqStatus, respondToUpdateAddOnFinalStatusErr)
		return err
	}

	return nil
}

// this function is used to parse service key JSON
func prepareToGetLPAPIAccessToken(config *abapEnvironmentUpdateAddOnProductOptions, clientAT *http.Client, servKey *serviceKey) error {
	// parse the service key from JSON string to struct
	servKeyJSON := config.LandscapePortalAPIServiceKey
	parseServiceKeyErr := json.Unmarshal([]byte(servKeyJSON), servKey)

	if parseServiceKeyErr != nil {
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

	*clientAT = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{certificate},
			},
		},
	}

	return nil
}

// this function is used to get access token of Landscape Portal API
func getLPAPIAccessToken(config *abapEnvironmentUpdateAddOnProductOptions, clientAT http.Client, servKey serviceKey) (string, error) {
	// define the raw url of the request
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

	// send request and get response
	resp, getAccessTokenErr := clientAT.Do(req)

	if getAccessTokenErr != nil {
		return "", getAccessTokenErr
	}

	defer resp.Body.Close()

	// error case of response status code being non 200
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Unexpected response status %v received when getting access token of LP API.", resp.Status)
		return "", err
	}

	// read and parse response body
	var respBody accessTokenResp
	parseRespBody[accessTokenResp](resp, &respBody)

	return respBody.AccessToken, nil
}

// this function is used to check the existence of integration test system
func getSystemBySystemNumber(config *abapEnvironmentUpdateAddOnProductOptions, client http.Client, clientAT http.Client, servKey serviceKey, systemId *string) error {
	// get access token
	accessToken, getAccessTokenErr := getLPAPIAccessToken(config, clientAT, servKey)

	if getAccessTokenErr != nil {
		return getAccessTokenErr
	}

	// define the raw url of the request and parse it into required form used in http.Request
	getSystemRawURL := servKey.Url + "/api/v1.0/systems/:" + config.AbapSystemNumber
	getSystemURL, urlParseErr := url.Parse(getSystemRawURL)

	if urlParseErr != nil {
		return urlParseErr
	}

	// define the request
	req := http.Request{
		Method: http.MethodGet,
		URL:    getSystemURL,
		Header: map[string][]string{
			"Authorization": {"Bearer" + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
	}

	// send request and get response
	resp, getSystemErr := client.Do(&req)

	if getSystemErr != nil {
		return getSystemErr
	}

	defer resp.Body.Close()

	// error case of response status code being non 200
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Unexpected response status %v received when getting system with systemNumber %v.", resp.Status, config.AbapSystemNumber)
		return err
	}

	// read and parse response body
	respBody := systemEntity{}
	parseRespBodyErr := parseRespBody[systemEntity](resp, &respBody)

	if parseRespBodyErr != nil {
		return parseRespBodyErr
	}

	*systemId = respBody.SystemId

	return nil
}

// this function is used to define and maintain the request body of querying status of addon update request, and send request to pull the status of request
func getStatusOfUpdateAddOn(config *abapEnvironmentUpdateAddOnProductOptions, client http.Client, clientAT http.Client, servKey serviceKey, reqId string, status *string, getStatusReq *http.Request) error {
	// get access token
	accessToken, getAccessTokenErr := getLPAPIAccessToken(config, clientAT, servKey)

	if getAccessTokenErr != nil {
		return getAccessTokenErr
	}

	// define the raw url of the request and parse it into required form used in http.Request
	getStatusRawURL := servKey.Url + "/api/v1.0/requests/:" + reqId
	getStatusURL, urlParseErr := url.Parse(getStatusRawURL)

	if urlParseErr != nil {
		return urlParseErr
	}

	// define the request
	req := http.Request{
		Method: http.MethodGet,
		URL:    getStatusURL,
		Header: map[string][]string{
			"Authorization": {"Bearer" + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
	}

	// store the req in the global variable for later usage
	*getStatusReq = req

	// call function to pull status of request
	pullStatusOfUpdateAddOnErr := pullStatusOfUpdateAddOn(client, getStatusReq, reqId, status)

	if pullStatusOfUpdateAddOnErr != nil {
		return pullStatusOfUpdateAddOnErr
	}

	return nil
}

// this function is used to pull status of addon update request and maintain the status
func pullStatusOfUpdateAddOn(client http.Client, req *http.Request, reqId string, status *string) error {
	// send request and get response
	resp, getStatusErr := client.Do(req)

	if getStatusErr != nil {
		return getStatusErr
	}

	defer resp.Body.Close()

	// error case of response status code being non 200
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Unexpected response status %v received when getting status of request with id %v.", resp.Status, reqId)
		return err
	}

	// read and parse response body
	respBody := reqEntity{}
	parseRespBodyErr := parseRespBody[reqEntity](resp, &respBody)

	if parseRespBodyErr != nil {
		return parseRespBodyErr
	}

	*status = respBody.Status

	return nil
}

// this function is used to update addon
func updateAddOn(config *abapEnvironmentUpdateAddOnProductOptions, client http.Client, clientAT http.Client, servKey serviceKey, systemId string, reqId *string) error {
	// get access token
	accessToken, getAccessTokenErr := getLPAPIAccessToken(config, clientAT, servKey)

	if getAccessTokenErr != nil {
		return getAccessTokenErr
	}

	// read productName and productVersion from addon.yml
	addOnDescriptor, readAddOnErr := abaputils.ReadAddonDescriptor(config.AddonDescriptorFileName)

	if readAddOnErr != nil {
		return readAddOnErr
	}

	// define the raw url of the request and parse it into required form used in http.Request
	updateAddOnRawURL := servKey.Url + "/api/v1.0/systems/:" + systemId + "/deployProduct"

	// define the request body as a struct
	reqBody := updateAddOnReq{
		ProductName:    addOnDescriptor.AddonProduct,
		ProductVersion: addOnDescriptor.AddonVersionYAML,
	}

	// encode the request body to JSON
	var reqBuff bytes.Buffer

	json.NewEncoder(&reqBuff).Encode(reqBody)

	// define the http request
	req, reqErr := http.NewRequest(http.MethodPost, updateAddOnRawURL, &reqBuff)

	if reqErr != nil {
		return reqErr
	}

	req.Header = map[string][]string{
		"Authorization": {"Bearer" + accessToken},
		"Content-Type":  {"application/json"},
		"Accept":        {"application/json"},
	}

	// send request and get response
	resp, updateAddOnErr := client.Do(req)

	if updateAddOnErr != nil {
		return updateAddOnErr
	}

	defer resp.Body.Close()

	// error case of response status code being non 200
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Unexpected response status %v received when updating AddOn in system with systemNumber %v.", resp.Status, config.AbapSystemNumber)
		return err
	}

	respBody := updateAddOnResp{}
	parseRespBodyErr := parseRespBody[updateAddOnResp](resp, &respBody)

	if parseRespBodyErr != nil {
		return parseRespBodyErr
	}

	*reqId = respBody.RequestId

	return nil
}

// this function is used to cancel addon update
func cancelUpdateAddOn(config *abapEnvironmentUpdateAddOnProductOptions, client http.Client, clientAT http.Client, servKey serviceKey, reqId string) error {
	// get access token
	accessToken, getAccessTokenErr := getLPAPIAccessToken(config, clientAT, servKey)

	if getAccessTokenErr != nil {
		return getAccessTokenErr
	}

	// define the raw url of the request and parse it into required form used in http.Request
	cancelUpdateAddOnRawURL := servKey.Url + "/api/v1.0/requests/" + reqId
	cancelUpdateAddOnURL, urlParseErr := url.Parse(cancelUpdateAddOnRawURL)

	if urlParseErr != nil {
		return urlParseErr
	}

	// define the http request
	req := http.Request{
		Method: http.MethodDelete,
		URL:    cancelUpdateAddOnURL,
		Header: map[string][]string{
			"Authorization": {"Bearer" + accessToken},
			"Content-Type":  {"application/json"},
			"Accept":        {"application/json"},
		},
	}

	// send request and get response
	resp, cancelUpdateAddOnErr := client.Do(&req)

	if cancelUpdateAddOnErr != nil {
		return cancelUpdateAddOnErr
	}

	defer resp.Body.Close()

	// error case of response status code being non 204
	if resp.StatusCode != http.StatusNoContent {
		err := fmt.Errorf("Unexpected response status %v received when canceling addon update request with id %v", resp.Status, reqId)
		return err
	}

	return nil
}

// this function is used to respond to a final status of addon update
func respondToUpdateAddOnFinalStatus(config *abapEnvironmentUpdateAddOnProductOptions, client http.Client, clientAT http.Client, servKey serviceKey, reqId string, status string) error {
	switch status {
	case StatusComplete:
		fmt.Println("AddOn update succeeded.")
	case StatusError:
		fmt.Println("AddOn update failed and will be canceled.")

		cancelUpdateAddOnErr := cancelUpdateAddOn(config, client, clientAT, servKey, reqId)
		if cancelUpdateAddOnErr != nil {
			err := fmt.Errorf("Failed to cancel addon update. Error: %v", cancelUpdateAddOnErr)
			return err
		}
	case StatusAborted:
		fmt.Println("AddOn update is aborted.")
	}

	return nil
}

// this function is used to parse response body of http request
func parseRespBody[T comparable](resp *http.Response, respBody *T) error {
	// read response body
	respBodyRaw, readRespErr := io.ReadAll(resp.Body)

	if readRespErr != nil {
		return readRespErr
	}

	decodeRespBodyErr := json.Unmarshal(respBodyRaw, &respBody)

	if decodeRespBodyErr != nil {
		return decodeRespBodyErr
	}

	return nil
}
