package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func apiKeyValueMapUpload(config apiKeyValueMapUploadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runApiKeyValueMapUpload(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runApiKeyValueMapUpload(config *apiKeyValueMapUploadOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {

	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}
	clientOptions := piperhttp.ClientOptions{}
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, tokenErr := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if tokenErr != nil {
		return errors.Wrap(tokenErr, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)

	httpMethod := "POST"
	uploadApiKeyValueMapStatusURL := fmt.Sprintf("%s/apiportal/api/1.0/Management.svc/KeyMapEntries", serviceKey.OAuth.Host)
	header := make(http.Header)
	header.Add("Content-Type", "application/json")
	header.Add("Accept", "application/json")
	jsonObj := gabs.New()
	jsonObj.Set(config.Key, "name")
	jsonObj.Set(config.KeyValueMapName, "map_name")
	jsonObj.Set(config.Value, "value")
	jsonRootObj := gabs.New()
	jsonRootObj.Set(config.KeyValueMapName, "name")
	jsonRootObj.Set(true, "encrypted")
	jsonRootObj.Set("ENV", "scope")
	jsonRootObj.ArrayAppend(jsonObj, "keyMapEntryValues")
	jsonBody, jsonErr := json.Marshal(jsonRootObj)
	if jsonErr != nil {
		return errors.Wrapf(jsonErr, "json payload is invalid for key value map %q", config.KeyValueMapName)
	}
	payload := bytes.NewBuffer([]byte(string(jsonBody)))
	apiProxyUploadStatusResp, httpErr := httpClient.SendRequest(httpMethod, uploadApiKeyValueMapStatusURL, payload, header, nil)

	if httpErr != nil {
		return errors.Wrapf(httpErr, "HTTP %q request to %q failed with error", httpMethod, uploadApiKeyValueMapStatusURL)
	}

	if apiProxyUploadStatusResp != nil && apiProxyUploadStatusResp.Body != nil {
		defer apiProxyUploadStatusResp.Body.Close()
	}

	if apiProxyUploadStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response")
	}

	if apiProxyUploadStatusResp.StatusCode == http.StatusCreated {
		log.Entry().
			WithField("KeyValueMap", config.KeyValueMapName).
			Info("Successfully created api key value map artefact in API Portal")
		return nil
	}
	response, readErr := ioutil.ReadAll(apiProxyUploadStatusResp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", apiProxyUploadStatusResp.StatusCode)
	}

	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(response), apiProxyUploadStatusResp.StatusCode)
	return errors.Errorf("Failed to upload API key value map artefact, Response Status code: %v", apiProxyUploadStatusResp.StatusCode)
}
