package cmd

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"errors"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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
		return fmt.Errorf("failed to fetch Bearer Token: %w", tokenErr)
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)

	httpMethod := http.MethodPost
	uploadApiProxyStatusURL := fmt.Sprintf("%s/apiportal/api/1.0/Transport.svc/APIProxies", serviceKey.OAuth.Host)
	header := make(http.Header)
	header.Add("Accept", "application/zip")
	fileContent, readError := fileUtils.FileRead(config.FilePath)
	if readError != nil {
		return fmt.Errorf("Error reading file: %w", readError)
	}
	if !strings.Contains(config.FilePath, "zip") {
		return errors.New("not valid zip archive")
	}
	payload := []byte(b64.StdEncoding.EncodeToString(fileContent))
	apiProxyUploadStatusResp, httpErr := httpClient.SendRequest(httpMethod, uploadApiProxyStatusURL, bytes.NewBuffer(payload), header, nil)

	failureMessage := "Failed to upload API Proxy artefact"
	successMessage := "Successfully created api proxy artefact in API Portal"
	httpFileUploadRequestParameters := cpi.HttpFileUploadRequestParameters{
		ErrMessage:     failureMessage,
		FilePath:       config.FilePath,
		Response:       apiProxyUploadStatusResp,
		HTTPMethod:     httpMethod,
		HTTPURL:        uploadApiProxyStatusURL,
		HTTPErr:        httpErr,
		SuccessMessage: successMessage}
	return cpi.HTTPUploadUtils.HandleHTTPFileUploadResponse(httpFileUploadRequestParameters)
}
