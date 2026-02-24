package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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
		return fmt.Errorf("failed to fetch Bearer Token: %w", tokenErr)
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)

	httpMethod := http.MethodPost
	uploadApiKeyValueMapStatusURL := fmt.Sprintf("%s/apiportal/api/1.0/Management.svc/KeyMapEntries", serviceKey.OAuth.Host)
	header := make(http.Header)
	header.Add("Content-Type", "application/json")
	header.Add("Accept", "application/json")
	payload, jsonErr := createJSONPayload(config)
	if jsonErr != nil {
		return jsonErr
	}
	apiProxyUploadStatusResp, httpErr := httpClient.SendRequest(httpMethod, uploadApiKeyValueMapStatusURL, payload, header, nil)

	if httpErr != nil {
		return fmt.Errorf("HTTP %q request to %q failed with error: %w", httpMethod, uploadApiKeyValueMapStatusURL, httpErr)
	}

	if apiProxyUploadStatusResp != nil && apiProxyUploadStatusResp.Body != nil {
		defer apiProxyUploadStatusResp.Body.Close()
	}

	if apiProxyUploadStatusResp == nil {
		return fmt.Errorf("did not retrieve a HTTP response")
	}

	if apiProxyUploadStatusResp.StatusCode == http.StatusCreated {
		log.Entry().
			WithField("KeyValueMap", config.KeyValueMapName).
			Info("Successfully created api key value map artefact in API Portal")
		return nil
	}
	response, readErr := io.ReadAll(apiProxyUploadStatusResp.Body)

	if readErr != nil {
		return fmt.Errorf("HTTP response body could not be read, Response status code: %v: %w", apiProxyUploadStatusResp.StatusCode, readErr)
	}

	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(response), apiProxyUploadStatusResp.StatusCode)
	return fmt.Errorf("Failed to upload API key value map artefact, Response Status code: %v", apiProxyUploadStatusResp.StatusCode)
}

// createJSONPayload -return http payload as byte array
func createJSONPayload(config *apiKeyValueMapUploadOptions) (*bytes.Buffer, error) {
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
		return nil, fmt.Errorf("json payload is invalid for key value map %q: %w", config.KeyValueMapName, jsonErr)
	}
	payload := bytes.NewBuffer([]byte(jsonBody))
	return payload, nil
}
