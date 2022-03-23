package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/pkg/errors"
)

func apiProviderUpload(config apiProviderUploadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runApiProviderUpload(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runApiProviderUpload(config *apiProviderUploadOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {

	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}
	clientOptions := piperhttp.ClientOptions{}
	x := xsuaa.XSUAA{
		OAuthURL:     serviceKey.OAuth.OAuthTokenProviderURL,
		ClientID:     serviceKey.OAuth.ClientID,
		ClientSecret: serviceKey.OAuth.ClientSecret,
	}
	token, tokenErr := x.GetBearerToken()

	if tokenErr != nil {
		return errors.Wrap(tokenErr, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token.AccessToken)
	httpClient.SetOptions(clientOptions)
	return createApiProvider(config, telemetryData, httpClient, serviceKey.OAuth.Host, ioutil.ReadFile)
}

func createApiProvider(config *apiProviderUploadOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender, host string, readFile func(string) ([]byte, error)) error {
	httpMethod := http.MethodPost
	uploadApiProviderStatusURL := fmt.Sprintf("%s/apiportal/api/1.0/Management.svc/APIProviders", host)
	header := make(http.Header)
	header.Add("Content-Type", "application/json")
	header.Add("Accept", "application/json")
	payload, err := readFile(config.FilePath)
	if err != nil {
		return err
	}
	apiProviderUploadStatusResp, httpErr := httpClient.SendRequest(httpMethod, uploadApiProviderStatusURL, bytes.NewBuffer(payload), header, nil)

	if httpErr != nil {
		return errors.Wrapf(httpErr, "HTTP %q request to %q failed with error", httpMethod, uploadApiProviderStatusURL)
	}

	if apiProviderUploadStatusResp != nil && apiProviderUploadStatusResp.Body != nil {
		defer apiProviderUploadStatusResp.Body.Close()
	}

	if apiProviderUploadStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response")
	}

	if apiProviderUploadStatusResp.StatusCode == http.StatusCreated {
		log.Entry().
			WithField("apiProvider", config.FilePath).
			Info("Successfully created api provider artefact in API Portal")
		return nil
	}
	response, readErr := ioutil.ReadAll(apiProviderUploadStatusResp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", apiProviderUploadStatusResp.StatusCode)
	}

	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(response), apiProviderUploadStatusResp.StatusCode)
	return errors.Errorf("Failed to create API provider artefact, Response Status code: %v", apiProviderUploadStatusResp.StatusCode)
}
