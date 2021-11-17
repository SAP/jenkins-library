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

type AuthToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type CommunicationInstance struct {
	tmsUrl       string
	uaaUrl       string
	clientId     string
	clientSecret string
	httpClient   piperHttp.Uploader
	logger       *logrus.Entry
	isVerbose    bool
}

type Node struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type nodes struct {
	Nodes []Node `json:"nodes"`
}

type MtaExtDescriptor struct {
	Id            int64  `json:"id"`
	Description   string `json:"description"`
	MtaId         string `json:"mtaId"`
	MtaExtId      string `json:"mtaExtId"`
	MtaVersion    string `json:"mtaVersion"`
	LastChangedAt string `json:"lastChangedAt"`
}

type mtaExtDescriptors struct {
	MtaExtDescriptors []MtaExtDescriptor `json:"mtaExtDescriptors"`
}

type CommunicationInterface interface {
	GetNodes() ([]Node, error)
}

// NewCommunicationInstance returns CommunicationInstance structure with http client prepared for communication with TMS backend
func NewCommunicationInstance(httpClient piperHttp.Uploader, tmsUrl, uaaUrl, clientId, clientSecret string, isVerbose bool) (*CommunicationInstance, error) {
	logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/tms")

	communicationInstance := &CommunicationInstance{
		tmsUrl:       tmsUrl,
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
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	header.Add("Authorization", "Basic "+encodedUsernameColonPassword)

	// TODO: somewhere here the proxy should be considered as well

	urlFormData := url.Values{
		"username":   {communicationInstance.clientId},
		"password":   {communicationInstance.clientSecret},
		"grant_type": {"password"},
	}

	data, err := sendRequest(communicationInstance, http.MethodPost, "/oauth/token/?grant_type=client_credentials&response_type=token", strings.NewReader(urlFormData.Encode()), header, http.StatusOK, true)
	if err != nil {
		return "", err
	}

	var token AuthToken
	json.Unmarshal(data, &token)

	communicationInstance.logger.Info("OAuth Token retrieved successfully")
	return token.TokenType + " " + token.AccessToken, nil
}

func sendRequest(communicationInstance *CommunicationInstance, method, urlPathAndQuery string, body io.Reader, header http.Header, expectedStatusCode int, isTowardsUaa bool) ([]byte, error) {
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

	// err is not nil for HTTP status codes >= 300
	if err != nil {
		communicationInstance.logger.Errorf("HTTP request failed with error: %s", err)
		communicationInstance.logResponseBody(response)
		return nil, err
	}

	if response.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf("unexpected positive HTTP status code %v, while it was expected %v", response.StatusCode, expectedStatusCode)
	}

	data, _ := ioutil.ReadAll(response.Body)
	if !isTowardsUaa && communicationInstance.isVerbose {
		communicationInstance.logger.Debugf("Valid response body: %v", string(data))
	}
	defer response.Body.Close()
	return data, nil
}

func (communicationInstance *CommunicationInstance) logResponseBody(response *http.Response) {
	if response != nil && response.Body != nil {
		data, _ := ioutil.ReadAll(response.Body)
		communicationInstance.logger.Errorf("Response body: %s", data)
		response.Body.Close()
	}
}

func (communicationInstance *CommunicationInstance) GetNodes() ([]Node, error) {
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("Obtaining nodes started")
		communicationInstance.logger.Infof("tmsUrl: %v", communicationInstance.tmsUrl)
	}

	header := http.Header{}
	header.Add("Content-Type", "application/json")

	// TODO: somewhere here the proxy should be considered as well

	var aNodes []Node
	var data []byte
	data, err := sendRequest(communicationInstance, http.MethodGet, "/v2/nodes", nil, header, http.StatusOK, false)
	if err != nil {
		return aNodes, err
	}

	var getNodesResponse nodes
	json.Unmarshal(data, &getNodesResponse)
	aNodes = getNodesResponse.Nodes
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("Nodes obtained successfully")
	}
	return aNodes, nil
}

func (communicationInstance *CommunicationInstance) GetMtaExtDescriptor(nodeId int64, mtaId, mtaVersion string) (MtaExtDescriptor, error) {
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("Get MTA extension descriptor started")
		communicationInstance.logger.Infof("tmsUrl: %v, nodeId: %v, mtaId: %v, mtaVersion: %v", communicationInstance.tmsUrl, nodeId, mtaId, mtaVersion)
	}

	header := http.Header{}
	header.Add("Content-Type", "application/json")

	// TODO: somewhere here the proxy should be considered as well

	var mtaExtDescriptor MtaExtDescriptor
	var data []byte
	data, err := sendRequest(communicationInstance, http.MethodGet, fmt.Sprintf("/v2/nodes/%v/mtaExtDescriptors?mtaId=%v&mtaVersion=%v", nodeId, mtaId, mtaVersion), nil, header, http.StatusOK, false)
	if err != nil {
		return mtaExtDescriptor, err
	}

	var getMtaExtDescriptorsResponse mtaExtDescriptors
	json.Unmarshal(data, &getMtaExtDescriptorsResponse)
	if len(getMtaExtDescriptorsResponse.MtaExtDescriptors) != 0 {
		mtaExtDescriptor = getMtaExtDescriptorsResponse.MtaExtDescriptors[0]
	}

	if communicationInstance.isVerbose {
		if mtaExtDescriptor.Id != int64(0) { // the struct is initialized
			communicationInstance.logger.Info("MTA extension descriptor obtained successfully")
		} else {
			communicationInstance.logger.Warn("No MTA extension descriptor found")
		}
	}
	return mtaExtDescriptor, nil

}
