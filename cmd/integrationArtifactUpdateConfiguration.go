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
	"github.com/pkg/errors"
)

func integrationArtifactUpdateConfiguration(config integrationArtifactUpdateConfigurationOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runIntegrationArtifactUpdateConfiguration(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runIntegrationArtifactUpdateConfiguration(config *integrationArtifactUpdateConfigurationOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {
	clientOptions := piperhttp.ClientOptions{}
	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}
	configUpdateURL := fmt.Sprintf("%s/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')/$links/Configurations('%s')", serviceKey.OAuth.Host, config.IntegrationFlowID, config.IntegrationFlowVersion, config.ParameterKey)
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "PUT"
	header := make(http.Header)
	header.Add("Content-Type", "application/json")
	header.Add("Accept", "application/json")
	jsonObj := gabs.New()
	jsonObj.Set(config.ParameterValue, "ParameterValue")
	jsonBody, jsonErr := json.Marshal(jsonObj)

	if jsonErr != nil {
		return errors.Wrapf(jsonErr, "input json body is invalid for parameterValue %q", config.ParameterValue)
	}
	configUpdateResp, httpErr := httpClient.SendRequest(httpMethod, configUpdateURL, bytes.NewBuffer(jsonBody), header, nil)
	if httpErr != nil {
		return errors.Wrapf(httpErr, "HTTP %q request to %q failed with error", httpMethod, configUpdateURL)
	}

	if configUpdateResp != nil && configUpdateResp.Body != nil {
		defer configUpdateResp.Body.Close()
	}

	if configUpdateResp == nil {
		return errors.Errorf("did not retrieve a HTTP response")
	}

	if configUpdateResp.StatusCode == http.StatusAccepted {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("successfully updated the integration flow configuration parameter")
		return nil
	}
	response, readErr := io.ReadAll(configUpdateResp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", configUpdateResp.StatusCode)
	}

	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(response), configUpdateResp.StatusCode)
	return errors.Errorf("Failed to update the integration flow configuration parameter, Response Status code: %v", configUpdateResp.StatusCode)
}
