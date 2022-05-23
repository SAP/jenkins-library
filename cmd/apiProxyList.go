package cmd

import (
	"fmt"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/apim"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type apiProxyListUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The apiProxyListUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

func apiProxyList(config apiProxyListOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *apiProxyListCommonPipelineEnvironment) {

	httpClient := &piperhttp.Client{}
	err := runApiProxyList(&config, telemetryData, httpClient, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runApiProxyList(config *apiProxyListOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender, commonPipelineEnvironment *apiProxyListCommonPipelineEnvironment) error {
	apimData := apim.Bundle{APIServiceKey: config.APIServiceKey, Client: httpClient}
	err := apim.Utils.InitAPIM(&apimData)
	if err != nil {
		return err
	}
	return getApiProxyList(config, apimData, commonPipelineEnvironment)
}

func getApiProxyList(config *apiProxyListOptions, apistruct apim.Bundle, commonPipelineEnvironment *apiProxyListCommonPipelineEnvironment) error {
	httpClient := apistruct.Client
	httpMethod := http.MethodGet
	odataFilterInputs := apim.OdataParameters{Filter: config.Filter, Search: config.Search,
		Top: config.Top, Skip: config.Skip, Orderby: config.Orderby,
		Select: config.Select, Expand: config.Expand}
	odataFilters, urlErr := apim.OdataUtils.InitOdataQuery(&odataFilterInputs)
	if urlErr != nil {

		return errors.New("failed to create odata filter")
	}
	getApiPrxyListURL := fmt.Sprintf("%s/apiportal/api/1.0/Management.svc/APIProxies%s", apistruct.Host, odataFilters)
	header := make(http.Header)
	header.Add("Content-Type", "application/json")
	header.Add("Accept", "application/json")
	apiProxyListResp, httpErr := httpClient.SendRequest(httpMethod, getApiPrxyListURL, nil, header, nil)
	failureMessage := "Failed to create API provider artefact"
	successMessage := "Successfully created api provider artefact in API Portal"
	httpGetRequestParameters := cpi.HttpFileUploadRequestParameters{
		ErrMessage:     failureMessage,
		Response:       apiProxyListResp,
		HTTPMethod:     httpMethod,
		HTTPURL:        getApiPrxyListURL,
		HTTPErr:        httpErr,
		SuccessMessage: successMessage}
	resp, err := cpi.HTTPUploadUtils.HandleHTTPGetRequestResponse(httpGetRequestParameters)
	commonPipelineEnvironment.custom.APIProxyList = resp
	return err
}
