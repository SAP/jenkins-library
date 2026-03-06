package cmd

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"errors"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const retryCount = 14

type integrationArtifactDeployUtils interface {
	command.ExecRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The integrationArtifactDeployUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type integrationArtifactDeployUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to integrationArtifactDeployUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// integrationArtifactDeployUtilsBundle and forward to the implementation of the dependency.
}

func newIntegrationArtifactDeployUtils() integrationArtifactDeployUtils {
	utils := integrationArtifactDeployUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func integrationArtifactDeploy(config integrationArtifactDeployOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newIntegrationArtifactDeployUtils()
	utils.Stdout(log.Writer())
	httpClient := &piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runIntegrationArtifactDeploy(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runIntegrationArtifactDeploy(config *integrationArtifactDeployOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {
	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/json")
	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}
	deployURL := fmt.Sprintf("%s/api/v1/DeployIntegrationDesigntimeArtifact?Id='%s'&Version='%s'", serviceKey.OAuth.Host, config.IntegrationFlowID, "Active")

	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return fmt.Errorf("failed to fetch Bearer Token: %w", err)
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "POST"
	deployResp, httpErr := httpClient.SendRequest(httpMethod, deployURL, nil, header, nil)
	if httpErr != nil {
		return fmt.Errorf("HTTP %v request to %v failed with error: %w", httpMethod, deployURL, httpErr)
	}

	if deployResp != nil && deployResp.Body != nil {
		defer deployResp.Body.Close()
	}

	if deployResp == nil {
		return fmt.Errorf("did not retrieve a HTTP response")
	}

	if deployResp.StatusCode == http.StatusAccepted {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("successfully deployed into CPI runtime")
		taskId, readErr := io.ReadAll(deployResp.Body)
		if readErr != nil {
			return fmt.Errorf("Task Id not found. HTTP response body could not be read.: %w", readErr)
		}
		deploymentError := pollIFlowDeploymentStatus(string(taskId), retryCount, config, httpClient, serviceKey.OAuth.Host)
		return deploymentError
	}
	responseBody, readErr := io.ReadAll(deployResp.Body)

	if readErr != nil {
		return fmt.Errorf("HTTP response body could not be read, response status code: %v: %w", deployResp.StatusCode, readErr)
	}
	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code : %v", string(responseBody), deployResp.StatusCode)
	return fmt.Errorf("integration flow deployment failed, response Status code: %v", deployResp.StatusCode)
}

// pollIFlowDeploymentStatus - Poll the integration flow deployment status, return status or error details
func pollIFlowDeploymentStatus(taskId string, retryCount int, config *integrationArtifactDeployOptions, httpClient piperhttp.Sender, apiHost string) error {

	if retryCount <= 0 {
		return errors.New("failed to start integration artifact after retrying several times")
	}
	deployStatus, err := getIntegrationArtifactDeployStatus(config, httpClient, apiHost, taskId)
	if err != nil {
		return err
	}

	//if artifact starting, then retry based on provided retry count
	//with specific delay between each retry
	if deployStatus == "DEPLOYING" {
		// Calling Sleep method
		sleepTime := int(retryCount * 3)
		time.Sleep(time.Duration(sleepTime) * time.Second)
		retryCount--
		return pollIFlowDeploymentStatus(taskId, retryCount, config, httpClient, apiHost)
	}

	//if artifact started, then just return
	if deployStatus == "SUCCESS" {
		return nil
	}

	//if error return immediately with error details
	if deployStatus == "FAIL" || deployStatus == "FAIL_ON_LICENSE_ERROR" {
		resp, err := getIntegrationArtifactDeployError(config, httpClient, apiHost)
		if err != nil {
			return err
		}
		return errors.New(resp)
	}
	return nil
}

// GetHTTPErrorMessage - Return HTTP failure message
func getHTTPErrorMessage(httpErr error, response *http.Response, httpMethod, statusURL string) (string, error) {
	responseBody, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return "", fmt.Errorf("HTTP response body could not be read, response status code: %v: %w", response.StatusCode, readErr)
	}
	log.Entry().Errorf("a HTTP error occurred! Response body: %v, response status code: %v", string(responseBody), response.StatusCode)
	return "", fmt.Errorf("HTTP %v request to %v failed with error: %v: %w", httpMethod, statusURL, responseBody, httpErr)
}

// getIntegrationArtifactDeployStatus - Get integration artifact Deploy Status
func getIntegrationArtifactDeployStatus(config *integrationArtifactDeployOptions, httpClient piperhttp.Sender, apiHost string, taskId string) (string, error) {
	httpMethod := "GET"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	header.Add("Accept", "application/json")
	deployStatusURL := fmt.Sprintf("%s/api/v1/BuildAndDeployStatus(TaskId='%s')", apiHost, taskId)
	deployStatusResp, httpErr := httpClient.SendRequest(httpMethod, deployStatusURL, nil, header, nil)

	if deployStatusResp != nil && deployStatusResp.Body != nil {
		defer deployStatusResp.Body.Close()
	}

	if deployStatusResp == nil {
		return "", fmt.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if deployStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("Successfully started integration flow artefact in CPI runtime")

		bodyText, readErr := io.ReadAll(deployStatusResp.Body)
		if readErr != nil {
			return "", fmt.Errorf("HTTP response body could not be read, response status code: %v: %w", deployStatusResp.StatusCode, readErr)
		}
		jsonResponse, parsingErr := gabs.ParseJSON([]byte(bodyText))
		if parsingErr != nil {
			return "", fmt.Errorf("HTTP response body could not be parsed as JSON: %v: %w", string(bodyText), parsingErr)
		}
		deployStatus := jsonResponse.Path("d.Status").Data().(string)
		return deployStatus, nil
	}
	if httpErr != nil {
		return getHTTPErrorMessage(httpErr, deployStatusResp, httpMethod, deployStatusURL)
	}
	return "", fmt.Errorf("failed to get Integration Flow artefact runtime status, response Status code: %v", deployStatusResp.StatusCode)
}

// getIntegrationArtifactDeployError - Get integration artifact deploy error details
func getIntegrationArtifactDeployError(config *integrationArtifactDeployOptions, httpClient piperhttp.Sender, apiHost string) (string, error) {
	httpMethod := "GET"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	errorStatusURL := fmt.Sprintf("%s/api/v1/IntegrationRuntimeArtifacts('%s')/ErrorInformation/$value", apiHost, config.IntegrationFlowID)
	errorStatusResp, httpErr := httpClient.SendRequest(httpMethod, errorStatusURL, nil, header, nil)

	if errorStatusResp != nil && errorStatusResp.Body != nil {
		defer errorStatusResp.Body.Close()
	}

	if errorStatusResp == nil {
		return "", fmt.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if errorStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("Successfully retrieved Integration Flow artefact deploy error details")
		responseBody, readErr := io.ReadAll(errorStatusResp.Body)
		if readErr != nil {
			return "", fmt.Errorf("HTTP response body could not be read, response status code: %v: %w", errorStatusResp.StatusCode, readErr)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), errorStatusResp.StatusCode)
		errorDetails := string(responseBody)
		return errorDetails, nil
	}
	if httpErr != nil {
		return getHTTPErrorMessage(httpErr, errorStatusResp, httpMethod, errorStatusURL)
	}
	return "", fmt.Errorf("failed to get Integration Flow artefact deploy error details, response Status code: %v", errorStatusResp.StatusCode)
}
