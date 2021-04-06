package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type integrationArtifactGetServiceEndpointUtils interface {
	command.ExecRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The integrationArtifactGetServiceEndpointUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type integrationArtifactGetServiceEndpointUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to integrationArtifactGetServiceEndpointUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// integrationArtifactGetServiceEndpointUtilsBundle and forward to the implementation of the dependency.
}

func newIntegrationArtifactGetServiceEndpointUtils() integrationArtifactGetServiceEndpointUtils {
	utils := integrationArtifactGetServiceEndpointUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func integrationArtifactGetServiceEndpoint(config integrationArtifactGetServiceEndpointOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *integrationArtifactGetServiceEndpointCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runIntegrationArtifactGetServiceEndpoint(&config, telemetryData, httpClient, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runIntegrationArtifactGetServiceEndpoint(config *integrationArtifactGetServiceEndpointOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender, commonPipelineEnvironment *integrationArtifactGetServiceEndpointCommonPipelineEnvironment) error {
	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/json")

	servieEndpointURL := fmt.Sprintf("%s/api/v1/ServiceEndpoints?$expand=EntryPoints", config.Host)
	tokenParameters := cpi.TokenParameters{TokenURL: config.OAuthTokenProviderURL, Username: config.Username, Password: config.Password, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "GET"
	serviceEndpointResp, httpErr := httpClient.SendRequest(httpMethod, servieEndpointURL, nil, header, nil)

	if httpErr != nil {
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error", httpMethod, servieEndpointURL)
	}

	if serviceEndpointResp != nil && serviceEndpointResp.Body != nil {
		defer serviceEndpointResp.Body.Close()
	}

	if serviceEndpointResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if serviceEndpointResp.StatusCode == 200 {
		bodyText, readErr := ioutil.ReadAll(serviceEndpointResp.Body)
		if readErr != nil {
			return errors.Wrap(readErr, "HTTP response body could not be read")
		}
		jsonResponse, parsingErr := gabs.ParseJSON([]byte(bodyText))
		if parsingErr != nil {
			return errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
		}

		for _, child := range jsonResponse.S("d", "results").Children() {
			iflowID := strings.ReplaceAll(child.Path("Name").String(), "\"", "")
			if iflowID == config.IntegrationFlowID {
				entryPoints := child.S("EntryPoints")
				finalEndpoint := entryPoints.Path("results.0.Url").Data().(string)
				commonPipelineEnvironment.custom.iFlowServiceEndpoint = finalEndpoint
				return nil
			}
		}
	}
	responseBody, readErr := ioutil.ReadAll(serviceEndpointResp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", serviceEndpointResp.StatusCode)
	}

	log.Entry().Errorf("a HTTP error occurred!  Response body: %v, Response status code: %v", responseBody, serviceEndpointResp.StatusCode)
	return errors.Errorf("Unable to get integration flow service endpoint, Response Status code: %v", serviceEndpointResp.StatusCode)
}
