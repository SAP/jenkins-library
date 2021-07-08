package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type integrationArtifactGetMplStatusUtils interface {
	command.ExecRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The integrationArtifactGetMplStatusUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type integrationArtifactGetMplStatusUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to integrationArtifactGetMplStatusUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// integrationArtifactGetMplStatusUtilsBundle and forward to the implementation of the dependency.
}

func newIntegrationArtifactGetMplStatusUtils() integrationArtifactGetMplStatusUtils {
	utils := integrationArtifactGetMplStatusUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

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
		mplStatus := jsonResponse.Path("d.results.0.Status").Data().(string)
		commonPipelineEnvironment.custom.iFlowMplStatus = mplStatus
		return nil
	}
	responseBody, readErr := ioutil.ReadAll(mplStatusResp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", mplStatusResp.StatusCode)
	}

	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, mplStatusResp.StatusCode)
	return errors.Errorf("Unable to get integration flow MPL status, Response Status code: %v", mplStatusResp.StatusCode)
}
