package cpi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"

	"github.com/Jeffail/gabs/v2"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
)

//CommonUtils for CPI
type CommonUtils interface {
	GetBearerToken() (string, error)
}

//HttpCPIUtils for CPI
type HttpCPIUtils interface {
	HandleHTTPFileDownloadResponse() error
}

//TokenParameters struct
type TokenParameters struct {
	TokenURL, Username, Password string
	Client                       piperhttp.Sender
}

//HttpParameters struct
type HttpFileDownloadRequestParameters struct {
	ErrMessage, FileDownloadPath string
	Response                     *http.Response
}

// ServiceKey contains information about a CPI service key
type ServiceKey struct {
	OAuth OAuth `json:"oauth"`
}

// OAuth is inside a CPI service key and contains more needed information
type OAuth struct {
	Host                  string `json:"url"`
	OAuthTokenProviderURL string `json:"tokenurl"`
	ClientID              string `json:"clientid"`
	ClientSecret          string `json:"clientsecret"`
}

// ReadCpiServiceKey unmarshalls the give json service key string.
func ReadCpiServiceKey(serviceKeyJSON string) (cpiServiceKey ServiceKey, err error) {
	// parse
	err = json.Unmarshal([]byte(serviceKeyJSON), &cpiServiceKey)
	if err != nil {
		err = errors.Wrap(err, "error unmarshalling serviceKey")
		return
	}

	log.Entry().Info("CPI serviceKey read successfully")
	return
}

// GetBearerToken -Provides the bearer token for making CPI OData calls
func (tokenParameters TokenParameters) GetBearerToken() (string, error) {

	httpClient := tokenParameters.Client

	clientOptions := piperhttp.ClientOptions{
		Username: tokenParameters.Username,
		Password: tokenParameters.Password,
	}
	httpClient.SetOptions(clientOptions)

	header := make(http.Header)
	header.Add("Accept", "application/json")
	tokenFinalURL := fmt.Sprintf("%s?grant_type=client_credentials", tokenParameters.TokenURL)
	method := "POST"
	resp, httpErr := httpClient.SendRequest(method, tokenFinalURL, nil, header, nil)
	if httpErr != nil {
		return "", errors.Wrapf(httpErr, "HTTP %v request to %v failed with error", method, tokenFinalURL)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp == nil {
		return "", errors.Errorf("did not retrieve a HTTP response")
	}

	if resp.StatusCode != 200 {
		return "", errors.Errorf("did not retrieve a valid HTTP response code: %v", httpErr)
	}

	bodyText, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return "", errors.Wrap(readErr, "HTTP response body could not be read")
	}
	jsonResponse, parsingErr := gabs.ParseJSON([]byte(bodyText))
	if parsingErr != nil {
		return "", errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
	}
	token := jsonResponse.Path("access_token").Data().(string)
	return token, nil
}

// HandleHTTPFileDownloadResponse - handle the file download response for http multipart response
func (httpFileDownloadRequestParameters HttpFileDownloadRequestParameters) HandleHTTPFileDownloadResponse() error {
	response := httpFileDownloadRequestParameters.Response
	contentDisposition := response.Header.Get("Content-Disposition")
	disposition, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return errors.Wrapf(err, "failed to read filename from http response headers, Content-Disposition "+disposition)
	}
	filename := params["filename"]

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if response.StatusCode == 200 {
		workspaceRelativePath := httpFileDownloadRequestParameters.FileDownloadPath
		err = os.MkdirAll(workspaceRelativePath, 0755)
		if err != nil {
			return errors.Wrapf(err, "Failed to create workspace directory")
		}
		zipFileName := filepath.Join(workspaceRelativePath, filename)
		file, err := os.Create(zipFileName)
		if err != nil {
			return errors.Wrapf(err, httpFileDownloadRequestParameters.ErrMessage)
		}
		io.Copy(file, response.Body)
		return nil
	}
	responseBody, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code : %v", response.StatusCode)
	}
	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code : %v", responseBody, response.StatusCode)
	return errors.Errorf("%s, Response Status code: %v", httpFileDownloadRequestParameters.ErrMessage, response.StatusCode)
}
