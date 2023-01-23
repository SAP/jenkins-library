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

func valueMappingArtifactUpload(config valueMappingArtifactUploadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}
	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runValueMappingArtifactUpload(&config, telemetryData, fileUtils, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runValueMappingArtifactUpload(config *valueMappingArtifactUploadOptions, telemetryData *telemetry.CustomData, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender) error {

	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}

	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/json")
	vMapStatusServiceURL := fmt.Sprintf("%s/api/v1/ValueMappingDesigntimeArtifacts(Id='%s',Version='%s')", serviceKey.OAuth.Host, config.ValueMappingID, "Active")
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "GET"

	//Check availability of value mapping artifact in CPI design time
	vMapStatusResp, httpErr := httpClient.SendRequest(httpMethod, vMapStatusServiceURL, nil, header, nil)

	if vMapStatusResp != nil && vMapStatusResp.Body != nil {
		defer vMapStatusResp.Body.Close()
	}
	if vMapStatusResp.StatusCode == 200 {
		return UploadValueMappingArtifact(config, httpClient, fileUtils, serviceKey.OAuth.Host)
	} else if httpErr != nil && vMapStatusResp.StatusCode == 404 {
		return UploadValueMappingArtifact(config, httpClient, fileUtils, serviceKey.OAuth.Host)
	}

	if vMapStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if httpErr != nil {
		responseBody, readErr := ioutil.ReadAll(vMapStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", vMapStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, vMapStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, vMapStatusServiceURL, string(responseBody))
	}
	return errors.Errorf("Failed to check value mapping availability, Response Status code: %v", vMapStatusResp.StatusCode)
}

// UploadValueMappingArtifact - Upload new value mapping artifact
func UploadValueMappingArtifact(config *valueMappingArtifactUploadOptions, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils, apiHost string) error {
	httpMethod := "POST"
	uploadVmapStatusURL := fmt.Sprintf("%s/api/v1/ValueMappingDesigntimeArtifacts", apiHost)
	header := make(http.Header)
	header.Add("content-type", "application/json")
	payload, jsonError := GetJSONPayloadAsByteArray2(config, "create", fileUtils)
	if jsonError != nil {
		return errors.Wrapf(jsonError, "Failed to get json payload for file %v, failed with error", config.FilePath)
	}

	uploadVmapStatusResp, httpErr := httpClient.SendRequest(httpMethod, uploadVmapStatusURL, payload, header, nil)

	if uploadVmapStatusResp != nil && uploadVmapStatusResp.Body != nil {
		defer uploadVmapStatusResp.Body.Close()
	}

	if uploadVmapStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if uploadVmapStatusResp.StatusCode == http.StatusCreated {
		log.Entry().
			WithField("ValueMappingID", config.ValueMappingID).
			Info("Successfully created value mapping artifact in CPI designtime")
		return nil
	}
	if httpErr != nil {
		responseBody, readErr := ioutil.ReadAll(uploadVmapStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", uploadVmapStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, uploadVmapStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, uploadVmapStatusURL, string(responseBody))
	}
	return errors.Errorf("Failed to create Value mapping artifact, Response Status code: %v", uploadVmapStatusResp.StatusCode)
}

// UpdateValueMappingArtifact - Update existing value mapping artifact
func UpdateValueMappingArtifact(config *valueMappingArtifactUploadOptions, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils, apiHost string) error {
	httpMethod := "PUT"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	updateVmapStatusURL := fmt.Sprintf("%s/api/v1/ValueMappingDesigntimeArtifacts(Id='%s',Version='%s')", apiHost, config.ValueMappingID, "Active")
	payload, jsonError := GetJSONPayloadAsByteArray2(config, "update", fileUtils)
	if jsonError != nil {
		return errors.Wrapf(jsonError, "Failed to get json payload for file %v, failed with error", config.FilePath)
	}
	updateVmapStatusResp, httpErr := httpClient.SendRequest(httpMethod, updateVmapStatusURL, payload, header, nil)

	if updateVmapStatusResp != nil && updateVmapStatusResp.Body != nil {
		defer updateVmapStatusResp.Body.Close()
	}

	if updateVmapStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if updateVmapStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("ValueMappingID", config.ValueMappingID).
			Info("Successfully updated Value Mapping artifact in CPI designtime")
		return nil
	}
	if httpErr != nil {
		responseBody, readErr := ioutil.ReadAll(updateVmapStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", updateVmapStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), updateVmapStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, updateVmapStatusURL, string(responseBody))
	}
	return errors.Errorf("Failed to update Value Mapping artifact, Response Status code: %v", updateVmapStatusResp.StatusCode)
}

// GetJSONPayloadAsByteArray2 -return http payload as byte array
func GetJSONPayloadAsByteArray2(config *valueMappingArtifactUploadOptions, mode string, fileUtils piperutils.FileUtils) (*bytes.Buffer, error) {
	fileContent, readError := fileUtils.FileRead(config.FilePath)
	if readError != nil {
		return nil, errors.Wrapf(readError, "Error reading file")
	}
	jsonObj := gabs.New()
	if mode == "create" {
		jsonObj.Set(config.ValueMappingName, "Name")
		jsonObj.Set(config.ValueMappingID, "Id")
		jsonObj.Set(config.PackageID, "PackageId")
		jsonObj.Set(b64.StdEncoding.EncodeToString(fileContent), "ArtifactContent")
	} else if mode == "update" {
		jsonObj.Set(config.ValueMappingName, "Name")
		jsonObj.Set(b64.StdEncoding.EncodeToString(fileContent), "ArtifactContent")
	} else {
		return nil, fmt.Errorf("Unkown node: '%s'", mode)
	}

	jsonBody, jsonErr := json.Marshal(jsonObj)

	if jsonErr != nil {
		return nil, errors.Wrapf(jsonErr, "json payload is invalid for value mapping artifact %q", config.ValueMappingID)
	}
	return bytes.NewBuffer(jsonBody), nil
}
