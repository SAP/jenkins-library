package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type integrationArtifactTriggerIntegrationTestUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The integrationArtifactTriggerIntegrationTestUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type integrationArtifactTriggerIntegrationTestUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to integrationArtifactTriggerIntegrationTestUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// integrationArtifactTriggerIntegrationTestUtilsBundle and forward to the implementation of the dependency.
}

func newIntegrationArtifactTriggerIntegrationTestUtils() integrationArtifactTriggerIntegrationTestUtils {
	utils := integrationArtifactTriggerIntegrationTestUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func integrationArtifactTriggerIntegrationTest(config integrationArtifactTriggerIntegrationTestOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *integrationArtifactTriggerIntegrationTestCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newIntegrationArtifactTriggerIntegrationTestUtils()
	httpClient := &piperhttp.Client{}
	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runIntegrationArtifactTriggerIntegrationTest(&config, utils, httpClient, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runIntegrationArtifactTriggerIntegrationTest(config *integrationArtifactTriggerIntegrationTestOptions, utils integrationArtifactTriggerIntegrationTestUtils, httpClient piperhttp.Sender, commonPipelineEnvironment *integrationArtifactTriggerIntegrationTestCommonPipelineEnvironment) error {
	var getServiceEndpointCommonPipelineEnvironment integrationArtifactGetServiceEndpointCommonPipelineEnvironment
	var serviceEndpointUrl string
	if len(config.IntegrationFlowServiceEndpointURL) > 0 {
		serviceEndpointUrl = config.IntegrationFlowServiceEndpointURL
	} else {
		serviceEndpointUrl = getServiceEndpointCommonPipelineEnvironment.custom.integrationFlowServiceEndpoint
		if len(serviceEndpointUrl) == 0 {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("IFlowServiceEndpointURL not set")
		}
	}
	log.Entry().Info("The Service URL : ", serviceEndpointUrl)

	// Here we trigger the iFlow Service Endpoint.
	IFlowErr := callIFlowURL(config, utils, httpClient, serviceEndpointUrl, commonPipelineEnvironment)
	if IFlowErr != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("failed to execute iFlow: %w", IFlowErr)
	}

	return nil
}

func callIFlowURL(
	config *integrationArtifactTriggerIntegrationTestOptions,
	utils integrationArtifactTriggerIntegrationTestUtils,
	httpIFlowClient piperhttp.Sender,
	serviceEndpointUrl string,
	commonPipelineEnvironment *integrationArtifactTriggerIntegrationTestCommonPipelineEnvironment) error {

	var fileBody []byte
	var httpMethod string
	var header http.Header
	if len(config.MessageBodyPath) > 0 {
		if len(config.ContentType) == 0 {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("message body file %s given, but no ContentType", config.MessageBodyPath)
		}
		exists, err := utils.FileExists(config.MessageBodyPath)
		if err != nil {
			log.SetErrorCategory(log.ErrorUndefined)
			// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
			return fmt.Errorf("failed to check message body file %s: %w", config.MessageBodyPath, err)
		}
		if !exists {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("message body file %s configured, but not found", config.MessageBodyPath)
		}

		var fileErr error
		fileBody, fileErr = os.ReadFile(config.MessageBodyPath)
		if fileErr != nil {
			log.SetErrorCategory(log.ErrorUndefined)
			return fmt.Errorf("failed to read file %s: %w", config.MessageBodyPath, fileErr)
		}
		httpMethod = "POST"
		header = make(http.Header)
		header.Add("Content-Type", config.ContentType)
	} else {
		httpMethod = "GET"
	}

	serviceKey, err := cpi.ReadCpiServiceKey(config.IntegrationFlowServiceKey)
	if err != nil {
		return err
	}
	clientOptions := piperhttp.ClientOptions{}
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpIFlowClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return fmt.Errorf("failed to fetch Bearer Token: %w", err)
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	clientOptions.MaxRetries = -1
	httpIFlowClient.SetOptions(clientOptions)
	iFlowResp, httpErr := httpIFlowClient.SendRequest(httpMethod, serviceEndpointUrl, bytes.NewBuffer(fileBody), header, nil)

	if httpErr != nil {
		return fmt.Errorf("HTTP %q request to %q failed with error: %w", httpMethod, serviceEndpointUrl, httpErr)
	}

	if iFlowResp == nil {
		return fmt.Errorf("did not retrieve any HTTP response")
	}

	if iFlowResp.StatusCode < 400 {
		log.Entry().
			WithField(config.IntegrationFlowID, serviceEndpointUrl).
			Infof("successfully triggered %s with status code %d", serviceEndpointUrl, iFlowResp.StatusCode)
		bodyText, readErr := io.ReadAll(iFlowResp.Body)
		if readErr != nil {
			log.Entry().Warnf("HTTP response body could not be read. Error: %s", readErr.Error())
		} else if len(bodyText) > 0 {
			commonPipelineEnvironment.custom.integrationFlowTriggerIntegrationTestResponseBody = string(bodyText)
		}
		headersJson, err := json.Marshal(iFlowResp.Header)
		if err != nil {
			log.Entry().Warnf("HTTP response headers could not be marshalled. Error: %s", err.Error())
		} else {
			commonPipelineEnvironment.custom.integrationFlowTriggerIntegrationTestResponseHeaders = string(headersJson)
		}
	} else {
		return fmt.Errorf("request %s failed with response code %d", serviceEndpointUrl, iFlowResp.StatusCode)
	}

	return nil
}
