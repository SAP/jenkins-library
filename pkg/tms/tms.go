package tms

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// AuthToken - structure to store OAuth token
type AuthToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// CommunicationInstance - structure to store data and objects (including http client) required for communication with TMS backend
type CommunicationInstance struct {
	uaaUrl       string
	tmsUrl       string
	clientId     string
	clientSecret string
	httpClient   piperHttp.Uploader
	logger       *logrus.Entry
	isVerbose    bool
}

// NewCommunicationInstance returns CommunicationInstance structure with http client prepared for communication with TMS backend
func NewCommunicationInstance(httpClient piperHttp.Uploader, uaaUrl, clientId, clientSecret string, isVerbose bool) (*CommunicationInstance, error) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/tms")

	communicationInstance := &CommunicationInstance{
		uaaUrl:       uaaUrl,
		clientId:     clientId,
		clientSecret: clientSecret,
		httpClient:   httpClient,
		logger:       logger,
		isVerbose:    isVerbose,
	}

	token, err := communicationInstance.getOAuthToken()
	if err != nil {
		return communicationInstance, errors.Wrap(err, "Error fetching OAuth token")
	}

	log.RegisterSecret(token)

	options := piperHttp.ClientOptions{
		Token: token,
	}
	communicationInstance.httpClient.SetOptions(options)

	return communicationInstance, nil
}

func (communicationInstance *CommunicationInstance) getOAuthToken() (string, error) {
	communicationInstance.logger.Info("OAuth token retrieval started")

	if communicationInstance.isVerbose {
		communicationInstance.logger.Infof("uaaUrl: %v, clientId: %v", communicationInstance.uaaUrl, communicationInstance.clientId)
	}

	encodedUsernameColonPassword := b64.StdEncoding.EncodeToString([]byte(communicationInstance.clientId + ":" + communicationInstance.clientSecret))
	header := http.Header{}
	header.Add("Content-type", "application/x-www-form-urlencoded")
	header.Add("authorization", "Basic "+encodedUsernameColonPassword)

	// TODO: somewhere here the proxy should be considered as well

	// TODO: should one need to replace '%20' with '+' for username and passowrd, as it was done in groovy?
	urlFormData := url.Values{
		"username":   {communicationInstance.clientId},
		"password":   {communicationInstance.clientSecret},
		"grant_type": {"password"},
	}

	data, err := sendRequest(communicationInstance, http.MethodPost, "/oauth/token/?grant_type=client_credentials&response_type=token", strings.NewReader(urlFormData.Encode()), header, true)
	if err != nil {
		return "", err
	}

	var token AuthToken
	json.Unmarshal(data, &token)

	communicationInstance.logger.Info("OAuth Token retrieved successfully")
	return token.AccessToken, nil
}

func sendRequest(communicationInstance *CommunicationInstance, method, urlPathAndQuery string, body io.Reader, header http.Header, isTowardsUaa bool) ([]byte, error) {
	var requestBody io.Reader
	if body != nil {
		closer := ioutil.NopCloser(body)
		bodyBytes, _ := ioutil.ReadAll(closer)
		requestBody = bytes.NewBuffer(bodyBytes)
		defer closer.Close()
	}

	url := communicationInstance.tmsUrl
	if isTowardsUaa {
		url = communicationInstance.uaaUrl
	}
	url = strings.TrimSuffix(url, "/")

	response, err := communicationInstance.httpClient.SendRequest(method, fmt.Sprintf("%v%v", url, urlPathAndQuery), requestBody, header, nil)

	// TODO: how to check for accepted status code?
	if err != nil {
		communicationInstance.recordResponseDetailsInErrorCase(response)
		communicationInstance.logger.Errorf("HTTP request failed with error: %s", err)
		return nil, err
	}

	data, _ := ioutil.ReadAll(response.Body)
	if !isTowardsUaa {
		communicationInstance.logger.Debugf("Valid response body: %v", string(data))
	}
	defer response.Body.Close()
	return data, nil
}

func (communicationInstance *CommunicationInstance) recordResponseDetailsInErrorCase(response *http.Response) {
	if response != nil && response.Body != nil {
		data, _ := ioutil.ReadAll(response.Body)
		communicationInstance.logger.Errorf("Response body: %s", data)
		response.Body.Close()
	}
}
