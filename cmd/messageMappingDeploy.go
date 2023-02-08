package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

const retryCountMM = 14

type messageMappingDeployUtils interface {
	command.ExecRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The messageMappingDeployUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type messageMappingDeployUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to messageMappingDeployUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// messageMappingDeployUtilsBundle and forward to the implementation of the dependency.
}

func newMessageMappingDeployUtils() messageMappingDeployUtils {
	utils := messageMappingDeployUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func messageMappingDeploy(config messageMappingDeployOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newMessageMappingDeployUtils()
	utils.Stdout(log.Writer())
	httpClient := &piperhttp.Client{}
	log.Entry().Info(httpClient, "httpClient")
	log.Entry().Info(config, "config")
	log.Entry().Info(utils, "utils")
	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runMessageMappingDeploy(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runMessageMappingDeploy(config *messageMappingDeployOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {
	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/json")
	log.Entry().Info(config, "serviceKey")
	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)

	if err != nil {
		return err
	}
	fmt.Println(serviceKey, "serviceKey")
	deployURL := fmt.Sprintf("%s/api/v1/DeployMessageMappingDesigntimeArtifact?Id='%s'&Version='%s'", serviceKey.OAuth.Host, config.MessageMappingID, "Active")
	fmt.Println(deployURL, "deployURL")
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	fmt.Println(token, "token")

	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "POST"
	deployResp, httpErr := httpClient.SendRequest(httpMethod, deployURL, nil, header, nil)
	if httpErr != nil {
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error", httpMethod, deployURL)
	}
	log.Entry().Info(deployResp, "deployResponse1")
	log.Entry().Info(deployResp.Body, "deployBodyResponse1")
	if deployResp != nil && deployResp.Body != nil {
		defer deployResp.Body.Close()
	}

	if deployResp == nil {
		return errors.Errorf("did not retrieve a HTTP response")
	}

	if deployResp.StatusCode == http.StatusAccepted {
		log.Entry().
			WithField("MessageMappingID", config.MessageMappingID).
			Info("successfully deployed into CPI runtime")
		// taskId, readErr := ioutil.ReadAll(deployResp.Body)
		// log.Entry().Info(deployResp.Body, "deployBodyResponse")
		// log.Entry().Info(taskId, "taskId")
		// if readErr != nil {
		// 	return errors.Wrap(readErr, "Task Id not found. HTTP response body could not be read.")
		// }
		// deploymentError := pollMessageMappingDeploymentStatus(string(taskId), retryCountMM, config, httpClient, serviceKey.OAuth.Host)
		return nil
	}
	responseBody, readErr := ioutil.ReadAll(deployResp.Body)
	log.Entry().Info(readErr, "REEEEEEEEEEAAAAAAAAAAAADDDDDDD")
	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, response status code: %v", deployResp.StatusCode)
	}
	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code : %v", string(responseBody), deployResp.StatusCode)
	return errors.Errorf("message mapping deployment failed, response Status code: %v", deployResp.StatusCode)
}

// pollMessageMappingDeploymentStatus - Poll the message mapping deployment status, return status or error details
func pollMessageMappingDeploymentStatus(taskId string, retryCountMM int, config *messageMappingDeployOptions, httpClient piperhttp.Sender, apiHost string) error {

	if retryCountMM <= 0 {
		return errors.New("failed to start message mapping artifact after retrying several times")
	}
	deployStatus, err := getMessageMappingDeployStatus(config, httpClient, apiHost, taskId)
	if err != nil {
		return err
	}

	//if artifact starting, then retry based on provided retry count
	//with specific delay between each retry
	if deployStatus == "DEPLOYING" {
		// Calling Sleep method
		sleepTime := int(retryCountMM * 3)
		time.Sleep(time.Duration(sleepTime) * time.Second)
		retryCountMM--
		return pollMessageMappingDeploymentStatus(taskId, retryCountMM, config, httpClient, apiHost)
	}

	//if artifact started, then just return
	if deployStatus == "SUCCESS" {
		return nil
	}

	//if error return immediately with error details
	if deployStatus == "FAIL" || deployStatus == "FAIL_ON_LICENSE_ERROR" {
		resp, err := getMessageMappingDeployError(config, httpClient, apiHost)
		if err != nil {
			return err
		}
		resp = "Error"
		return errors.New(resp)
	}
	return nil
}

// GetHTTPErrorMessage - Return HTTP failure message
func getMMHTTPErrorMessage(httpErr error, response *http.Response, httpMethod, statusURL string) (string, error) {
	responseBody, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return "", errors.Wrapf(readErr, "HTTP response body could not be read, response status code: %v", response.StatusCode)
	}
	log.Entry().Errorf("a HTTP error occurred! Response body: %v, response status code: %v", string(responseBody), response.StatusCode)
	return "", errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, statusURL, responseBody)
}

// getMessageMappingDeployStatus - Get integration artifact Deploy Status
func getMessageMappingDeployStatus(config *messageMappingDeployOptions, httpClient piperhttp.Sender, apiHost string, taskId string) (string, error) {
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
		return "", errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if deployStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("MessageMappingID", config.MessageMappingID).
			Info("Successfully started message mapping artefact in CPI runtime")

		bodyText, readErr := ioutil.ReadAll(deployStatusResp.Body)
		if readErr != nil {
			return "", errors.Wrapf(readErr, "HTTP response body could not be read, response status code: %v", deployStatusResp.StatusCode)
		}
		jsonResponse, parsingErr := gabs.ParseJSON([]byte(bodyText))
		if parsingErr != nil {
			return "", errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
		}
		deployStatus := jsonResponse.Path("d.Status").Data().(string)
		return deployStatus, nil
	}
	if httpErr != nil {
		return getMMHTTPErrorMessage(httpErr, deployStatusResp, httpMethod, deployStatusURL)
	}
	return "", errors.Errorf("failed to get message Mapping artefact runtime status, response Status code: %v", deployStatusResp.StatusCode)
}

// getIntegrationArtifactDeployError - Get integration artifact deploy error details
func getMessageMappingDeployError(config *messageMappingDeployOptions, httpClient piperhttp.Sender, apiHost string) (string, error) {
	httpMethod := "GET"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	errorStatusURL := fmt.Sprintf("%s/api/v1/IntegrationRuntimeArtifacts('%s')/ErrorInformation/$value", apiHost, config.MessageMappingID)
	errorStatusResp, httpErr := httpClient.SendRequest(httpMethod, errorStatusURL, nil, header, nil)

	if errorStatusResp != nil && errorStatusResp.Body != nil {
		defer errorStatusResp.Body.Close()
	}

	if errorStatusResp == nil {
		return "", errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if errorStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("MessageMappingID", config.MessageMappingID).
			Info("Successfully retrieved message mapping artefact deploy error details")
		responseBody, readErr := ioutil.ReadAll(errorStatusResp.Body)
		if readErr != nil {
			return "", errors.Wrapf(readErr, "HTTP response body could not be read, response status code: %v", errorStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), errorStatusResp.StatusCode)
		errorDetails := string(responseBody)
		return errorDetails, nil
	}
	if httpErr != nil {
		return getMMHTTPErrorMessage(httpErr, errorStatusResp, httpMethod, errorStatusURL)
	}
	return "", errors.Errorf("failed to get message mapping artefact deploy error details, response Status code: %v", errorStatusResp.StatusCode)
}
