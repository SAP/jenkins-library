package cmd

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func messageMappingUpload(config messageMappingUploadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}
	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runMessageMappingUpload(&config, telemetryData, fileUtils, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMessageMappingUpload(config *messageMappingUploadOptions, telemetryData *telemetry.CustomData, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender) error {

	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}

	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/json")
	iFlowStatusServiceURL := fmt.Sprintf("%s/api/v1/MessageMappingDesigntimeArtifacts(Id='%s',Version='%s')", serviceKey.OAuth.Host, config.MessageMappingID, "Active")
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "GET"

	//Check availability of message mapping in CPI design time
	iFlowStatusResp, httpErr := httpClient.SendRequest(httpMethod, iFlowStatusServiceURL, nil, header, nil)

	if iFlowStatusResp != nil && iFlowStatusResp.Body != nil {
		defer iFlowStatusResp.Body.Close()
	}
	if iFlowStatusResp.StatusCode == 200 {
		return UpdateMessageMapping(config, httpClient, fileUtils, serviceKey.OAuth.Host)
	} else if httpErr != nil && iFlowStatusResp.StatusCode == 404 {
		return UploadMessageMapping(config, httpClient, fileUtils, serviceKey.OAuth.Host)
	}

	if iFlowStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if httpErr != nil {
		responseBody, readErr := ioutil.ReadAll(iFlowStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", iFlowStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, iFlowStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, iFlowStatusServiceURL, string(responseBody))
	}
	return errors.Errorf("Failed to check message mapping availability, Response Status code: %v", iFlowStatusResp.StatusCode)
}

// UploadMessageMapping - Upload new message mapping
func UploadMessageMapping(config *messageMappingUploadOptions, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils, apiHost string) error {
	httpMethod := "POST"
	uploadIflowStatusURL := fmt.Sprintf("%s/api/v1/MessageMappingDesigntimeArtifacts", apiHost)
	header := make(http.Header)
	header.Add("content-type", "application/json")
	payload, jsonError := GetJSONPayloadAsByteArrayMM(config, "create", fileUtils)
	if jsonError != nil {
		return errors.Wrapf(jsonError, "Failed to get json payload for file %v, failed with error", config.FilePath)
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
			WithField("MessageMappingID", config.MessageMappingID).
			Info("Successfully created message mapping artefact in CPI designtime")
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
	return errors.Errorf("Failed to create message mapping artefact, Response Status code: %v", uploadIflowStatusResp.StatusCode)
}

// UpdateMessageMapping - Update existing message mapping
func UpdateMessageMapping(config *messageMappingUploadOptions, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils, apiHost string) error {
	httpMethod := "PUT"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	updateIflowStatusURL := fmt.Sprintf("%s/api/v1/MessageMappingDesigntimeArtifacts(Id='%s',Version='%s')", apiHost, config.MessageMappingID, "Active")
	payload, jsonError := GetJSONPayloadAsByteArrayMM(config, "update", fileUtils)
	if jsonError != nil {
		return errors.Wrapf(jsonError, "Failed to get json payload for file %v, failed with error", config.FilePath)
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
			WithField("MessageMappingID", config.MessageMappingID).
			Info("Successfully updated message mapping artefact in CPI designtime")
		return nil
	}
	if httpErr != nil {
		responseBody, readErr := ioutil.ReadAll(updateIflowStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", updateIflowStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), updateIflowStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, updateIflowStatusURL, string(responseBody))
	}
	return errors.Errorf("Failed to update message mapping artefact, Response Status code: %v", updateIflowStatusResp.StatusCode)
}

// GetJSONPayloadAsByteArray -return http payload as byte array
func GetJSONPayloadAsByteArrayMM(config *messageMappingUploadOptions, mode string, fileUtils piperutils.FileUtils) (*bytes.Buffer, error) {
	fileContent, readError := fileUtils.FileRead(config.FilePath)
	if readError != nil {
		return nil, errors.Wrapf(readError, "Error reading file")
	}
	jsonObj := gabs.New()
	if mode == "create" {
		jsonObj.Set(config.MessageMappingName, "Name")
		jsonObj.Set(config.MessageMappingID, "Id")
		jsonObj.Set(config.PackageID, "PackageId")
		jsonObj.Set(b64.StdEncoding.EncodeToString(fileContent), "ArtifactContent")
	} else if mode == "update" {
		jsonObj.Set(config.MessageMappingName, "Name")
		jsonObj.Set(b64.StdEncoding.EncodeToString(fileContent), "ArtifactContent")
	} else {
		return nil, fmt.Errorf("Unkown node: '%s'", mode)
	}

	jsonBody, jsonErr := json.Marshal(jsonObj)

	if jsonErr != nil {
		return nil, errors.Wrapf(jsonErr, "json payload is invalid for message mapping artifact %q", config.MessageMappingID)
	}
	return bytes.NewBuffer(jsonBody), nil
}
