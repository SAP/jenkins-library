package cmd

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func integrationArtifactDownload(config integrationArtifactDownloadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runIntegrationArtifactDownload(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runIntegrationArtifactDownload(config *integrationArtifactDownloadOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {
	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/zip")
	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}
	downloadArtifactURL := fmt.Sprintf("%s/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')/$value", serviceKey.OAuth.Host, config.IntegrationFlowID, config.IntegrationFlowVersion)
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return fmt.Errorf("failed to fetch Bearer Token: %w", err)
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "GET"
	downloadResp, httpErr := httpClient.SendRequest(httpMethod, downloadArtifactURL, nil, header, nil)
	if httpErr != nil {
		return fmt.Errorf("HTTP %v request to %v failed with error: %w", httpMethod, downloadArtifactURL, httpErr)
	}
	if downloadResp == nil {
		return fmt.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}
	contentDisposition := downloadResp.Header.Get("Content-Disposition")
	disposition, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return fmt.Errorf("failed to read filename from http response headers, Content-Disposition %s: %w", disposition, err)
	}
	filename := params["filename"]

	if downloadResp != nil && downloadResp.Body != nil {
		defer downloadResp.Body.Close()
	}

	if downloadResp.StatusCode == 200 {
		workspaceRelativePath := config.DownloadPath
		err = os.MkdirAll(workspaceRelativePath, 0755)
		if err != nil {
			return fmt.Errorf("Failed to create workspace directory: %w", err)
		}
		zipFileName := filepath.Join(workspaceRelativePath, filename)
		file, err := os.Create(zipFileName)
		if err != nil {
			return fmt.Errorf("Failed to create integration flow artifact file: %w", err)
		}
		if _, err := io.Copy(file, downloadResp.Body); err != nil {
			return err
		}
		return nil
	}
	responseBody, readErr := io.ReadAll(downloadResp.Body)

	if readErr != nil {
		return fmt.Errorf("HTTP response body could not be read, Response status code : %v: %w", downloadResp.StatusCode, readErr)
	}

	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code : %v", string(responseBody), downloadResp.StatusCode)
	return fmt.Errorf("Integration Flow artifact download failed, Response Status code: %v", downloadResp.StatusCode)
}
