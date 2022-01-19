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
	"os"
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

type FileInfo struct {
	Id   int64  `json:"fileId"`
	Name string `json:"fileName"`
}

type NodeUploadResponseEntity struct {
	TransportRequestId          int64        `json:"transportRequestId"`
	TransportRequestDescription string       `json:"transportRequestDescription"`
	QueueEntries                []QueueEntry `json:"queueEntries"`
}

type QueueEntry struct {
	Id       int64  `json:"queueId"`
	NodeId   int64  `json:"nodeId"`
	NodeName string `json:"nodeName"`
}

type NodeUploadRequestEntity struct {
	ContentType string  `json:"contentType"`
	StorageType string  `json:"storageType"`
	NodeName    string  `json:"nodeName"`
	Description string  `json:"description"`
	NamedUser   string  `json:"namedUser"`
	Entries     []Entry `json:"entries"`
}

type Entry struct {
	Uri string `json:"uri"`
}

type CommunicationInterface interface {
	GetNodes() ([]Node, error)
	GetMtaExtDescriptor(nodeId int64, mtaId, mtaVersion string) (MtaExtDescriptor, error)
	UpdateMtaExtDescriptor(nodeId, idOfMtaExtDescriptor int64, file, mtaVersion, description, namedUser string) (MtaExtDescriptor, error)
	UploadMtaExtDescriptorToNode(nodeId int64, file, mtaVersion, description, namedUser string) (MtaExtDescriptor, error)
	UploadFile(file, namedUser string) (FileInfo, error)
	UploadFileToNode(nodeName, fileId, description, namedUser string) (NodeUploadResponseEntity, error)
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
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("OAuth token retrieval started")
		communicationInstance.logger.Infof("uaaUrl: %v, clientId: %v", communicationInstance.uaaUrl, communicationInstance.clientId)
	}

	encodedUsernameColonPassword := b64.StdEncoding.EncodeToString([]byte(communicationInstance.clientId + ":" + communicationInstance.clientSecret))
	header := http.Header{}
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	header.Add("Authorization", "Basic "+encodedUsernameColonPassword)

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

	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("OAuth Token retrieved successfully")
	}
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

	var mtaExtDescriptor MtaExtDescriptor
	var data []byte
	data, err := sendRequest(communicationInstance, http.MethodGet, fmt.Sprintf("/v2/nodes/%v/mtaExtDescriptors?mtaId=%v&mtaVersion=%v", nodeId, mtaId, mtaVersion), nil, header, http.StatusOK, false)
	if err != nil {
		return mtaExtDescriptor, err
	}

	var getMtaExtDescriptorsResponse mtaExtDescriptors
	json.Unmarshal(data, &getMtaExtDescriptorsResponse)
	if len(getMtaExtDescriptorsResponse.MtaExtDescriptors) > 0 {
		mtaExtDescriptor = getMtaExtDescriptorsResponse.MtaExtDescriptors[0]
	}

	if communicationInstance.isVerbose {
		if mtaExtDescriptor != (MtaExtDescriptor{}) {
			communicationInstance.logger.Info("MTA extension descriptor obtained successfully")
		} else {
			communicationInstance.logger.Warn("No MTA extension descriptor found")
		}
	}
	return mtaExtDescriptor, nil

}

func (communicationInstance *CommunicationInstance) UploadFileToNode(nodeName, fileId, description, namedUser string) (NodeUploadResponseEntity, error) {
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("Node upload started")
		communicationInstance.logger.Infof("tmsUrl: %v, nodeName: %v, fileId: %v, description: %v, namedUser: %v", communicationInstance.tmsUrl, nodeName, fileId, description, namedUser)
	}

	header := http.Header{}
	header.Add("Content-Type", "application/json")

	var nodeUploadResponseEntity NodeUploadResponseEntity
	entry := Entry{Uri: fileId}
	body := NodeUploadRequestEntity{ContentType: "MTA", StorageType: "FILE", NodeName: nodeName, Description: description, NamedUser: namedUser, Entries: []Entry{entry}}
	bodyBytes, errMarshaling := json.Marshal(body)
	if errMarshaling != nil {
		return nodeUploadResponseEntity, errors.Wrapf(errMarshaling, "unable to marshal request body %v", body)
	}

	data, errSendRequest := sendRequest(communicationInstance, http.MethodPost, "/v2/nodes/upload", bytes.NewReader(bodyBytes), header, http.StatusOK, false)
	if errSendRequest != nil {
		return nodeUploadResponseEntity, errSendRequest
	}

	json.Unmarshal(data, &nodeUploadResponseEntity)
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("Node upload executed successfully")
	}
	return nodeUploadResponseEntity, nil

}

func (communicationInstance *CommunicationInstance) UpdateMtaExtDescriptor(nodeId, idOfMtaExtDescriptor int64, file, mtaVersion, description, namedUser string) (MtaExtDescriptor, error) {
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("Update of MTA extension descriptor started")
		communicationInstance.logger.Infof("tmsUrl: %v, nodeId: %v, mtaExtDescriptorId: %v, file: %v, mtaVersion: %v, description: %v, namedUser: %v", communicationInstance.tmsUrl, nodeId, idOfMtaExtDescriptor, file, mtaVersion, description, namedUser)
	}

	header := http.Header{}
	header.Add("tms-named-user", namedUser)

	tmsUrl := strings.TrimSuffix(communicationInstance.tmsUrl, "/")
	url := fmt.Sprintf("%v/v2/nodes/%v/mtaExtDescriptors/%v", tmsUrl, nodeId, idOfMtaExtDescriptor)
	formFields := map[string]string{"mtaVersion": mtaVersion, "description": description}

	var mtaExtDescriptor MtaExtDescriptor
	fileHandle, errOpenFile := os.Open(file)
	if errOpenFile != nil {
		return mtaExtDescriptor, errors.Wrapf(errOpenFile, "unable to locate file %v", file)
	}
	defer fileHandle.Close()

	uploadRequestData := piperHttp.UploadRequestData{Method: http.MethodPut, URL: url, File: file, FileFieldName: "file", FormFields: formFields, FileContent: fileHandle, Header: header, Cookies: nil}

	var data []byte
	data, errUpload := upload(communicationInstance, uploadRequestData, http.StatusOK)
	if errUpload != nil {
		return mtaExtDescriptor, errUpload
	}

	json.Unmarshal(data, &mtaExtDescriptor)
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("MTA extension descriptor updated successfully")
	}
	return mtaExtDescriptor, nil

}

func (communicationInstance *CommunicationInstance) UploadMtaExtDescriptorToNode(nodeId int64, file, mtaVersion, description, namedUser string) (MtaExtDescriptor, error) {
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("Upload of MTA extension descriptor started")
		communicationInstance.logger.Infof("tmsUrl: %v, nodeId: %v, file: %v, mtaVersion: %v, description: %v, namedUser: %v", communicationInstance.tmsUrl, nodeId, file, mtaVersion, description, namedUser)
	}

	header := http.Header{}
	header.Add("tms-named-user", namedUser)

	tmsUrl := strings.TrimSuffix(communicationInstance.tmsUrl, "/")
	url := fmt.Sprintf("%v/v2/nodes/%v/mtaExtDescriptors", tmsUrl, nodeId)
	formFields := map[string]string{"mtaVersion": mtaVersion, "description": description}

	var mtaExtDescriptor MtaExtDescriptor
	fileHandle, errOpenFile := os.Open(file)
	if errOpenFile != nil {
		return mtaExtDescriptor, errors.Wrapf(errOpenFile, "unable to locate file %v", file)
	}
	defer fileHandle.Close()

	uploadRequestData := piperHttp.UploadRequestData{Method: http.MethodPost, URL: url, File: file, FileFieldName: "file", FormFields: formFields, FileContent: fileHandle, Header: header, Cookies: nil}

	var data []byte
	data, errUpload := upload(communicationInstance, uploadRequestData, http.StatusCreated)
	if errUpload != nil {
		return mtaExtDescriptor, errUpload
	}

	json.Unmarshal(data, &mtaExtDescriptor)
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("MTA extension descriptor uploaded successfully")
	}
	return mtaExtDescriptor, nil

}

func (communicationInstance *CommunicationInstance) UploadFile(file, namedUser string) (FileInfo, error) {
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("Upload of file started")
		communicationInstance.logger.Infof("tmsUrl: %v, file: %v, namedUser: %v", communicationInstance.tmsUrl, file, namedUser)
	}

	tmsUrl := strings.TrimSuffix(communicationInstance.tmsUrl, "/")
	url := fmt.Sprintf("%v/v2/files/upload", tmsUrl)
	formFields := map[string]string{"namedUser": namedUser}

	var fileInfo FileInfo
	fileHandle, errOpenFile := os.Open(file)
	if errOpenFile != nil {
		return fileInfo, errors.Wrapf(errOpenFile, "unable to locate file %v", file)
	}
	defer fileHandle.Close()

	uploadRequestData := piperHttp.UploadRequestData{Method: http.MethodPost, URL: url, File: file, FileFieldName: "file", FormFields: formFields, FileContent: fileHandle, Header: http.Header{}, Cookies: nil}

	var data []byte
	data, errUpload := upload(communicationInstance, uploadRequestData, http.StatusCreated)
	if errUpload != nil {
		return fileInfo, errUpload
	}

	json.Unmarshal(data, &fileInfo)
	if communicationInstance.isVerbose {
		communicationInstance.logger.Info("File uploaded successfully")
	}
	return fileInfo, nil

}

func upload(communicationInstance *CommunicationInstance, uploadRequestData piperHttp.UploadRequestData, expectedStatusCode int) ([]byte, error) {
	response, err := communicationInstance.httpClient.Upload(uploadRequestData)

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
	if communicationInstance.isVerbose {
		communicationInstance.logger.Debugf("Valid response body: %v", string(data))
	}
	defer response.Body.Close()
	return data, nil
}
