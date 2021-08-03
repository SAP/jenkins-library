package cmd

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type integrationArtifactResourceUtils interface {
	command.ExecRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The integrationArtifactResourceUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type integrationArtifactResourceUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to integrationArtifactResourceUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// integrationArtifactResourceUtilsBundle and forward to the implementation of the dependency.
}

func newIntegrationArtifactResourceUtils() integrationArtifactResourceUtils {
	utils := integrationArtifactResourceUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
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
	mode := config.Operation
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

//UploadIntegrationArtifactResource - Upload new resource file to existing integration flow design time artefact
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

	if uploadIflowStatusResp != nil && uploadIflowStatusResp.Body != nil {
		defer uploadIflowStatusResp.Body.Close()
	}

	if uploadIflowStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if uploadIflowStatusResp.StatusCode == http.StatusCreated {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("Successfully uploaded a resource file in integration flow artefact in CPI designtime")
		return nil
	}
	if httpErr != nil {
		responseBody, readErr := ioutil.ReadAll(uploadIflowStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", uploadIflowStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, uploadIflowStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, uploadIflowStatusURL, string(responseBody))
	}
	return errors.Errorf("Failed to upload resource file in Integration Flow artefact, Response Status code: %v", uploadIflowStatusResp.StatusCode)
}

//UpdateIntegrationArtifactResource - Update integration artifact resource file
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

	if updateIflowStatusResp != nil && updateIflowStatusResp.Body != nil {
		defer updateIflowStatusResp.Body.Close()
	}

	if updateIflowStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if updateIflowStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("Successfully updated resource file of the integration flow artefact")
		return nil
	}
	if httpErr != nil {
		responseBody, readErr := ioutil.ReadAll(updateIflowStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", updateIflowStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, updateIflowStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, updateIflowStatusURL, string(responseBody))
	}
	return errors.Errorf("Failed to update rsource file of the Integration Flow artefact, Response Status code: %v", updateIflowStatusResp.StatusCode)
}

//DeleteIntegrationArtifactResource - Delete integration artifact resource file
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

	if deleteIflowResourceStatusResp != nil && deleteIflowResourceStatusResp.Body != nil {
		defer deleteIflowResourceStatusResp.Body.Close()
	}

	if deleteIflowResourceStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if deleteIflowResourceStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("Successfully updated integration flow artefact in CPI designtime")
		return nil
	}
	if httpErr != nil {
		responseBody, readErr := ioutil.ReadAll(deleteIflowResourceStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", deleteIflowResourceStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, deleteIflowResourceStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, deleteIflowResourceStatusURL, string(responseBody))
	}
	return errors.Errorf("Failed to update Integration Flow artefact, Response Status code: %v", deleteIflowResourceStatusResp.StatusCode)
}

//GetJSONPayload -return http payload as byte array
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

//GetResourceFileExtension -return resource file extension
func GetResourceFileExtension(filename string) string {
	fileExtension := filepath.Ext(filename)
	switch fileExtension {
	case ".xsl":
		return "xslt"
	case ".gsh":
		return "groovy"
	case ".groovy":
		return "groovy"
	case ".js":
		return "js"
	case ".jar":
		return "jar"
	default:
		return ""
	}
}
