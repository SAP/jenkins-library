package cmd

import (
	"fmt"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/apim"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func apiProviderList(config apiProviderListOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *apiProviderListCommonPipelineEnvironment) {
	httpClient := &piperhttp.Client{}
	err := runApiProviderList(&config, telemetryData, httpClient, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runApiProviderList(config *apiProviderListOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender, commonPipelineEnvironment *apiProviderListCommonPipelineEnvironment) error {
	apimData := apim.Bundle{APIServiceKey: config.APIServiceKey, Client: httpClient}
	err := apim.Utils.InitAPIM(&apimData)
	if err != nil {
		return err
	}
	return getApiProviderList(config, apimData, commonPipelineEnvironment)
}

func getApiProviderList(config *apiProviderListOptions, apistruct apim.Bundle, commonPipelineEnvironment *apiProviderListCommonPipelineEnvironment) error {
	httpClient := apistruct.Client
	httpMethod := http.MethodGet
	odataFilterInputs := apim.OdataParameters{Filter: config.Filter, Search: config.Search,
		Top: config.Top, Skip: config.Skip, Orderby: config.Orderby,
		Select: config.Select, Expand: config.Expand}
	odataFilters, urlErr := apim.OdataUtils.MakeOdataQuery(&odataFilterInputs)
	if urlErr != nil {
		return fmt.Errorf("failed to create odata filter: %w", urlErr)
	}
	getApiProviderListURL := fmt.Sprintf("%s/apiportal/api/1.0/Management.svc/APIProviders%s", apistruct.Host, odataFilters)
	header := make(http.Header)
	header.Add("Accept", "application/json")
	apiProviderListResp, httpErr := httpClient.SendRequest(httpMethod, getApiProviderListURL, nil, header, nil)
	failureMessage := "Failed to get List of API Providers"
	successMessage := "Successfully retrieved the api provider list from API Portal"
	httpGetRequestParameters := cpi.HttpFileUploadRequestParameters{
		ErrMessage:     failureMessage,
		Response:       apiProviderListResp,
		HTTPMethod:     httpMethod,
		HTTPURL:        getApiProviderListURL,
		HTTPErr:        httpErr,
		SuccessMessage: successMessage}
	resp, err := cpi.HTTPUploadUtils.HandleHTTPGetRequestResponse(httpGetRequestParameters)
	commonPipelineEnvironment.custom.APIProviderList = resp
	return err
}
