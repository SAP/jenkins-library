package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/apim"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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

	apimData := apim.APIMBundle{APIServiceKey: config.APIServiceKey, Client: httpClient}
	error := apim.APIMUtils.NewAPIM(&apimData)
	if error != nil {
		return error
	}
	return createApiProvider(config, apimData, ioutil.ReadFile)
}

func createApiProvider(config *apiProviderUploadOptions, apim apim.APIMBundle, readFile func(string) ([]byte, error)) error {
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
