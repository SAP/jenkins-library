package cmd

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type integrationArtifactResourceData struct {
	Method     string
	URL        string
	IFlowID    string
	ScsMessage string
	FlrMessage string
	StatusCode int
}

func integrationArtifactResource(config integrationArtifactResourceOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runIntegrationArtifactResource(&config, telemetryData, fileUtils, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runIntegrationArtifactResource(config *integrationArtifactResourceOptions, telemetryData *telemetry.CustomData, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender) error {
	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}

	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/json")
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	mode := strings.ToLower(strings.TrimSpace(config.Operation))
	switch mode {
	case "create":
		return UploadIntegrationArtifactResource(config, httpClient, fileUtils, serviceKey.OAuth.Host)
	case "update":
		return UpdateIntegrationArtifactResource(config, httpClient, fileUtils, serviceKey.OAuth.Host)
	case "delete":
		return DeleteIntegrationArtifactResource(config, httpClient, fileUtils, serviceKey.OAuth.Host)
	default:
		return errors.New("invalid input for resource operation")
	}
}

// UploadIntegrationArtifactResource - Upload new resource file to existing integration flow design time artefact
func UploadIntegrationArtifactResource(config *integrationArtifactResourceOptions, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils, apiHost string) error {
	httpMethod := "POST"
	uploadIflowStatusURL := fmt.Sprintf("%s/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')/Resources", apiHost, config.IntegrationFlowID, "Active")
	header := make(http.Header)
	header.Add("content-type", "application/json")
	payload, jsonError := GetJSONPayload(config, "create", fileUtils)
	if jsonError != nil {
		return errors.Wrapf(jsonError, "Failed to get json payload for file %v, failed with error", config.ResourcePath)
	}

	uploadIflowStatusResp, httpErr := httpClient.SendRequest(httpMethod, uploadIflowStatusURL, payload, header, nil)

	successMessage := "Successfully create a new resource file in the integration flow artefact"
	failureMessage := "Failed to create a new resource file in the integration flow artefact"
	integrationArtifactResourceData := integrationArtifactResourceData{
		Method:     httpMethod,
		URL:        uploadIflowStatusURL,
		IFlowID:    config.IntegrationFlowID,
		ScsMessage: successMessage,
		FlrMessage: failureMessage,
		StatusCode: http.StatusCreated,
	}

	return HttpResponseHandler(uploadIflowStatusResp, httpErr, &integrationArtifactResourceData)
}

// UpdateIntegrationArtifactResource - Update integration artifact resource file
func UpdateIntegrationArtifactResource(config *integrationArtifactResourceOptions, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils, apiHost string) error {
	httpMethod := "PUT"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	fileName := filepath.Base(config.ResourcePath)
	fileExt := GetResourceFileExtension(fileName)
	if fileExt == "" {
		return errors.New("invalid file extension in resource file")
	}
	updateIflowStatusURL := fmt.Sprintf("%s/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')/$links/Resources(Name='%s',ResourceType='%s')", apiHost, config.IntegrationFlowID, "Active", fileName, fileExt)
	payload, jsonError := GetJSONPayload(config, "update", fileUtils)
	if jsonError != nil {
		return errors.Wrapf(jsonError, "Failed to get json payload for file %v, failed with error", config.ResourcePath)
	}
	updateIflowStatusResp, httpErr := httpClient.SendRequest(httpMethod, updateIflowStatusURL, payload, header, nil)

	successMessage := "Successfully updated resource file of the integration flow artefact"
	failureMessage := "Failed to update rsource file of the integration flow artefact"
	integrationArtifactResourceData := integrationArtifactResourceData{
		Method:     httpMethod,
		URL:        updateIflowStatusURL,
		IFlowID:    config.IntegrationFlowID,
		ScsMessage: successMessage,
		FlrMessage: failureMessage,
		StatusCode: http.StatusOK,
	}

	return HttpResponseHandler(updateIflowStatusResp, httpErr, &integrationArtifactResourceData)
}

// DeleteIntegrationArtifactResource - Delete integration artifact resource file
func DeleteIntegrationArtifactResource(config *integrationArtifactResourceOptions, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils, apiHost string) error {
	httpMethod := "DELETE"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	fileName := filepath.Base(config.ResourcePath)
	fileExt := GetResourceFileExtension(fileName)
	if fileExt == "" {
		return errors.New("invalid file extension in resource file")
	}
	deleteIflowResourceStatusURL := fmt.Sprintf("%s/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')/$links/Resources(Name='%s',ResourceType='%s')", apiHost, config.IntegrationFlowID, "Active", fileName, fileExt)
	deleteIflowResourceStatusResp, httpErr := httpClient.SendRequest(httpMethod, deleteIflowResourceStatusURL, nil, header, nil)

	successMessage := "Successfully deleted a resource file in the integration flow artefact"
	failureMessage := "Failed to delete a resource file in the integration flow artefact"
	integrationArtifactResourceData := integrationArtifactResourceData{
		Method:     httpMethod,
		URL:        deleteIflowResourceStatusURL,
		IFlowID:    config.IntegrationFlowID,
		ScsMessage: successMessage,
		FlrMessage: failureMessage,
		StatusCode: http.StatusOK,
	}

	return HttpResponseHandler(deleteIflowResourceStatusResp, httpErr, &integrationArtifactResourceData)
}

// GetJSONPayload -return http payload as byte array
func GetJSONPayload(config *integrationArtifactResourceOptions, mode string, fileUtils piperutils.FileUtils) (*bytes.Buffer, error) {
	fileContent, readError := fileUtils.FileRead(config.ResourcePath)
	if readError != nil {
		return nil, errors.Wrapf(readError, "Error reading file")
	}
	fileName := filepath.Base(config.ResourcePath)
	jsonObj := gabs.New()
	if mode == "create" {
		jsonObj.Set(fileName, "Name")
		jsonObj.Set(GetResourceFileExtension(fileName), "ResourceType")
		jsonObj.Set(b64.StdEncoding.EncodeToString(fileContent), "ResourceContent")
	} else if mode == "update" {
		jsonObj.Set(b64.StdEncoding.EncodeToString(fileContent), "ResourceContent")
	} else {
		return nil, fmt.Errorf("Unkown node: '%s'", mode)
	}

	jsonBody, jsonErr := json.Marshal(jsonObj)

	if jsonErr != nil {
		return nil, errors.Wrapf(jsonErr, "json payload is invalid for integration flow artifact %q", config.IntegrationFlowID)
	}
	return bytes.NewBuffer(jsonBody), nil
}

// GetResourceFileExtension -return resource file extension
func GetResourceFileExtension(filename string) string {
	fileExtension := filepath.Ext(filename)
	switch fileExtension {
	case ".xsl":
		return "xslt"
	case ".gsh", ".groovy":
		return "groovy"
	case ".js":
		return "js"
	case ".jar":
		return "jar"
	default:
		return ""
	}
}

// HttpResponseHandler - handle http response object
func HttpResponseHandler(resp *http.Response, httpErr error, integrationArtifactResourceData *integrationArtifactResourceData) error {

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if resp.StatusCode == integrationArtifactResourceData.StatusCode {
		log.Entry().
			WithField("IntegrationFlowID", integrationArtifactResourceData.IFlowID).
			Info(integrationArtifactResourceData.ScsMessage)
		return nil
	}
	if httpErr != nil {
		responseBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", resp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), resp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", integrationArtifactResourceData.Method, integrationArtifactResourceData.URL, string(responseBody))
	}
	return errors.Errorf("%s, Response Status code: %v", integrationArtifactResourceData.FlrMessage, resp.StatusCode)
}
