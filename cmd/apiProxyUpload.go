package cmd

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func apiProxyUpload(config apiProxyUploadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}
	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runApiProxyUpload(&config, telemetryData, fileUtils, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runApiProxyUpload(config *apiProxyUploadOptions, telemetryData *telemetry.CustomData, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender) error {

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
	uploadApiProxyStatusURL := fmt.Sprintf("%s/apiportal/api/1.0/Transport.svc/APIProxies", serviceKey.OAuth.Host)
	header := make(http.Header)
	header.Add("Accept", "application/zip")
	fileContent, readError := fileUtils.FileRead(config.FilePath)
	if readError != nil {
		return errors.Wrapf(readError, "Error reading file")
	}
	payload := []byte(b64.StdEncoding.EncodeToString(fileContent))
	apiProxyUploadStatusResp, httpErr := httpClient.SendRequest(httpMethod, uploadApiProxyStatusURL, bytes.NewBuffer(payload), header, nil)

	if apiProxyUploadStatusResp != nil && apiProxyUploadStatusResp.Body != nil {
		defer apiProxyUploadStatusResp.Body.Close()
	}

	if apiProxyUploadStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if apiProxyUploadStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("Api Proxy artifact", config.FilePath).
			Info("Successfully created api proxy artefact in API Portal")
		return nil
	}
	if httpErr != nil {
		responseBody, readErr := ioutil.ReadAll(apiProxyUploadStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", apiProxyUploadStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), apiProxyUploadStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, uploadApiProxyStatusURL, string(responseBody))
	}
	return errors.Errorf("Failed to create api proxy artefact, Response Status code: %v", apiProxyUploadStatusResp.StatusCode)
}
