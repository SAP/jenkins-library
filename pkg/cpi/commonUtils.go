package cpi

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"

	"github.com/Jeffail/gabs/v2"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
)

// CommonUtils for CPI
type CommonUtils interface {
	GetBearerToken() (string, error)
}

// HttpCPIUtils for CPI
type HttpCPIUtils interface {
	HandleHTTPFileDownloadResponse() error
}

// HTTPUploadUtils for CPI
type HTTPUploadUtils interface {
	HandleHTTPFileUploadResponse() error
	HandleHTTPGetRequestResponse() (string, error)
}

// TokenParameters struct
type TokenParameters struct {
	TokenURL, Username, Password string
	Client                       piperhttp.Sender
}

// HttpParameters struct
type HttpFileDownloadRequestParameters struct {
	ErrMessage, FileDownloadPath string
	Response                     *http.Response
}

// HTTPFileUploadRequestParameters struct
type HttpFileUploadRequestParameters struct {
	ErrMessage, FilePath, HTTPMethod, HTTPURL, SuccessMessage string
	Response                                                  *http.Response
	HTTPErr                                                   error
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

	bodyText, readErr := io.ReadAll(resp.Body)
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

// HandleHTTPFileDownloadResponse - Handle the file download response for http multipart response
func (httpFileDownloadRequestParameters HttpFileDownloadRequestParameters) HandleHTTPFileDownloadResponse() error {
	response := httpFileDownloadRequestParameters.Response
	contentDisposition := response.Header.Get("Content-Disposition")
	disposition, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return errors.Wrapf(err, "failed to read filename from http response headers, Content-Disposition %s", disposition)
	}
	filename := params["filename"]

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if response.StatusCode == 200 {
		workspaceRelativePath := httpFileDownloadRequestParameters.FileDownloadPath
		err = os.MkdirAll(workspaceRelativePath, 0755)
		// handling error while creating a workspce directoy for file download, if one not exist already!
		if err != nil {
			return errors.Wrapf(err, "Failed to create workspace directory")
		}
		zipFileName := filepath.Join(workspaceRelativePath, filename)
		file, err := os.Create(zipFileName)
		// handling error while creating a file in the filesystem
		if err != nil {
			return errors.Wrap(err, "failed to create zip archive of api proxy")
		}
		_, err = io.Copy(file, response.Body)
		if err != nil {
			return err
		}
		return nil
	}
	responseBody, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", response.StatusCode)
	}
	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code : %v", responseBody, response.StatusCode)
	return errors.Errorf("%s, Response Status code: %v", httpFileDownloadRequestParameters.ErrMessage, response.StatusCode)
}

// HandleHTTPFileUploadResponse - Handle the file upload response
func (httpFileUploadRequestParameters HttpFileUploadRequestParameters) HandleHTTPFileUploadResponse() error {
	response := httpFileUploadRequestParameters.Response
	httpErr := httpFileUploadRequestParameters.HTTPErr
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if response == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}
	responseCode := response.StatusCode

	if (responseCode == http.StatusOK) || (responseCode == http.StatusCreated) {
		log.Entry().
			WithField("Created Artifact", httpFileUploadRequestParameters.FilePath).
			Info(httpFileUploadRequestParameters.SuccessMessage)
		return nil
	}
	if httpErr != nil {
		responseBody, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", response.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), response.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpFileUploadRequestParameters.HTTPMethod, httpFileUploadRequestParameters.HTTPURL, string(responseBody))
	}
	return errors.Errorf("%s, Response Status code: %v", httpFileUploadRequestParameters.ErrMessage, response.StatusCode)
}

// HandleHTTPGetRequestResponse - Handle the GET Request response data
func (httpGetRequestParameters HttpFileUploadRequestParameters) HandleHTTPGetRequestResponse() (string, error) {
	response := httpGetRequestParameters.Response
	httpErr := httpGetRequestParameters.HTTPErr
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if response == nil {
		return "", errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}
	if response.StatusCode == http.StatusOK {
		responseBody, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return "", errors.Wrapf(readErr, "HTTP response body could not be read, response status code: %v", response.StatusCode)
		}
		return string(responseBody), nil
	}
	if httpErr != nil {
		responseBody, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return "", errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", response.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), response.StatusCode)
		return "", errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpGetRequestParameters.HTTPMethod, httpGetRequestParameters.HTTPURL, string(responseBody))
	}
	return "", errors.Errorf("%s, Response Status code: %v", httpGetRequestParameters.ErrMessage, response.StatusCode)
}
