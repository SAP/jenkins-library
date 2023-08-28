package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/SAP/jenkins-library/pkg/apim"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func apiProviderUpload(config apiProviderUploadOptions, telemetryData *telemetry.CustomData) {
	httpClient := &piperhttp.Client{}
	err := runApiProviderUpload(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runApiProviderUpload(config *apiProviderUploadOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {

	apimData := apim.Bundle{APIServiceKey: config.APIServiceKey, Client: httpClient}
	err := apim.Utils.InitAPIM(&apimData)
	if err != nil {
		return err
	}
	return createApiProvider(config, apimData, os.ReadFile)
}

func createApiProvider(config *apiProviderUploadOptions, apim apim.Bundle, readFile func(string) ([]byte, error)) error {
	httpClient := apim.Client
	httpMethod := http.MethodPost
	uploadApiProviderStatusURL := fmt.Sprintf("%s/apiportal/api/1.0/Management.svc/APIProviders", apim.Host)
	header := make(http.Header)
	header.Add("Content-Type", "application/json")
	header.Add("Accept", "application/json")

	exists, _ := piperutils.FileExists(config.FilePath)
	if !exists {
		return errors.New("Missing API Provider input file")
	}

	payload, err := readFile(config.FilePath)
	if err != nil {
		return err
	}
	apim.Payload = string(payload)

	if !apim.IsPayloadJSON() {
		return errors.New("invalid JSON content in the input file")
	}

	apiProviderUploadStatusResp, httpErr := httpClient.SendRequest(httpMethod, uploadApiProviderStatusURL, bytes.NewBuffer(payload), header, nil)
	failureMessage := "Failed to create API provider artefact"
	successMessage := "Successfully created api provider artefact in API Portal"
	httpFileUploadRequestParameters := cpi.HttpFileUploadRequestParameters{
		ErrMessage:     failureMessage,
		FilePath:       config.FilePath,
		Response:       apiProviderUploadStatusResp,
		HTTPMethod:     httpMethod,
		HTTPURL:        uploadApiProviderStatusURL,
		HTTPErr:        httpErr,
		SuccessMessage: successMessage}
	return cpi.HTTPUploadUtils.HandleHTTPFileUploadResponse(httpFileUploadRequestParameters)
}
