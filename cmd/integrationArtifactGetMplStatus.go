package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func integrationArtifactGetMplStatus(config integrationArtifactGetMplStatusOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *integrationArtifactGetMplStatusCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}
	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runIntegrationArtifactGetMplStatus(&config, telemetryData, httpClient, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runIntegrationArtifactGetMplStatus(
	config *integrationArtifactGetMplStatusOptions,
	telemetryData *telemetry.CustomData,
	httpClient piperhttp.Sender,
	commonPipelineEnvironment *integrationArtifactGetMplStatusCommonPipelineEnvironment) error {

	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}

	clientOptions := piperhttp.ClientOptions{}
	httpClient.SetOptions(clientOptions)
	header := make(http.Header)
	header.Add("Accept", "application/json")
	mplStatusEncodedURL := fmt.Sprintf("%s/api/v1/MessageProcessingLogs?$filter=IntegrationArtifact/Id"+url.QueryEscape(" eq ")+"'%s'"+
		url.QueryEscape(" and Status ne ")+"'DISCARDED'"+"&$orderby="+url.QueryEscape("LogEnd desc")+"&$top=1", serviceKey.OAuth.Host, config.IntegrationFlowID)
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "GET"
	mplStatusResp, httpErr := httpClient.SendRequest(httpMethod, mplStatusEncodedURL, nil, header, nil)
	if httpErr != nil {
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error", httpMethod, mplStatusEncodedURL)
	}

	if mplStatusResp != nil && mplStatusResp.Body != nil {
		defer mplStatusResp.Body.Close()
	}

	if mplStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if mplStatusResp.StatusCode == 200 {
		bodyText, readErr := ioutil.ReadAll(mplStatusResp.Body)
		if readErr != nil {
			return errors.Wrap(readErr, "HTTP response body could not be read")
		}
		jsonResponse, parsingErr := gabs.ParseJSON([]byte(bodyText))
		if parsingErr != nil {
			return errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
		}
		if jsonResponse == nil {
			return errors.Errorf("Empty json response: %v", string(bodyText))
		}
		if jsonResponse.Exists("d", "results", "0") {
			mplStatus := jsonResponse.Path("d.results.0.Status").Data().(string)
			commonPipelineEnvironment.custom.integrationFlowMplStatus = mplStatus

			//if error, then return immediately with the error details
			if mplStatus == "FAILED" {
				mplID := jsonResponse.Path("d.results.0.MessageGuid").Data().(string)
				resp, err := getIntegrationArtifactMPLError(commonPipelineEnvironment, mplID, httpClient, serviceKey.OAuth.Host)
				if err != nil {
					return err
				}
				return errors.New(resp)
			}
		}
		return nil
	}
	responseBody, readErr := ioutil.ReadAll(mplStatusResp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", mplStatusResp.StatusCode)
	}

	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), mplStatusResp.StatusCode)
	return errors.Errorf("Unable to get integration flow MPL status, Response Status code: %v", mplStatusResp.StatusCode)
}

// getIntegrationArtifactMPLError - Get integration artifact MPL error details
func getIntegrationArtifactMPLError(commonPipelineEnvironment *integrationArtifactGetMplStatusCommonPipelineEnvironment, mplID string, httpClient piperhttp.Sender, apiHost string) (string, error) {
	httpMethod := "GET"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	errorStatusURL := fmt.Sprintf("%s/api/v1/MessageProcessingLogs('%s')/ErrorInformation/$value", apiHost, mplID)
	errorStatusResp, httpErr := httpClient.SendRequest(httpMethod, errorStatusURL, nil, header, nil)

	if errorStatusResp != nil && errorStatusResp.Body != nil {
		defer errorStatusResp.Body.Close()
	}

	if errorStatusResp == nil {
		return "", errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if errorStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("MPLID", mplID).
			Info("Successfully retrieved Integration Flow artefact message processing error")
		responseBody, readErr := ioutil.ReadAll(errorStatusResp.Body)
		if readErr != nil {
			return "", errors.Wrapf(readErr, "HTTP response body could not be read, response status code: %v", errorStatusResp.StatusCode)
		}
		mplErrorDetails := string(responseBody)
		commonPipelineEnvironment.custom.integrationFlowMplError = mplErrorDetails
		return mplErrorDetails, nil
	}
	if httpErr != nil {
		return getHTTPErrorMessage(httpErr, errorStatusResp, httpMethod, errorStatusURL)
	}
	return "", errors.Errorf("failed to get Integration Flow artefact message processing error, response Status code: %v", errorStatusResp.StatusCode)
}
