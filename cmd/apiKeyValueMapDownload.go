package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func apiKeyValueMapDownload(config apiKeyValueMapDownloadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runApiKeyValueMapDownload(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runApiKeyValueMapDownload(config *apiKeyValueMapDownloadOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {
	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/json")
	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}
	downloadkeyValueMapArtifactURL := fmt.Sprintf("%s/apiportal/api/1.0/Management.svc/KeyMapEntries('%s')", serviceKey.OAuth.Host, config.KeyValueMapName)
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL,
		Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return fmt.Errorf("failed to fetch Bearer Token: %w", err)
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := http.MethodGet
	downloadResp, httpErr := httpClient.SendRequest(httpMethod, downloadkeyValueMapArtifactURL, nil, header, nil)
	if httpErr != nil {
		return fmt.Errorf("HTTP %v request to %v failed with error: %w", httpMethod, downloadkeyValueMapArtifactURL, httpErr)
	}
	if downloadResp == nil {
		return fmt.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}
	if downloadResp != nil && downloadResp.Body != nil {
		defer downloadResp.Body.Close()
	}
	if downloadResp.StatusCode == 200 {
		csvFilePath := config.DownloadPath
		file, err := os.Create(csvFilePath)
		if err != nil {
			return fmt.Errorf("Failed to create api key value map CSV file: %w", err)
		}
		_, err = io.Copy(file, downloadResp.Body)
		if err != nil {
			return err
		}
		return nil
	}
	responseBody, readErr := io.ReadAll(downloadResp.Body)

	if readErr != nil {
		return fmt.Errorf("HTTP response body could not be read, Response status code : %v: %w", downloadResp.StatusCode, readErr)
	}
	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code : %v", responseBody, downloadResp.StatusCode)
	return fmt.Errorf("api Key value map download failed, Response Status code: %v", downloadResp.StatusCode)
}
