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

const retryCountVM = 14

type valueMappingDeployUtils interface {
	command.ExecRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The valueMappingDeployUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type valueMappingDeployUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to valueMappingDeployUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// valueMappingDeployUtilsBundle and forward to the implementation of the dependency.
}

func newValueMappingDeployUtils() valueMappingDeployUtils {
	utils := valueMappingDeployUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func valueMappingDeploy(config valueMappingDeployOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newValueMappingDeployUtils()
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
	err := runValueMappingDeploy(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runValueMappingDeploy(config *valueMappingDeployOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {
	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/json")
	log.Entry().Info(config, "serviceKey")
	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)

	if err != nil {
		return err
	}
	fmt.Println(serviceKey, "serviceKey")
	deployURL := fmt.Sprintf("%s/api/v1/DeployValueMappingDesigntimeArtifact?Id='%s'&Version='%s'", serviceKey.OAuth.Host, config.ValueMappingID, "Active")
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
			WithField("ValueMappingID", config.ValueMappingID).
			Info("successfully deployed into CPI runtime")
		// taskId, readErr := ioutil.ReadAll(deployResp.Body)
		// log.Entry().Info(deployResp.Body, "deployBodyResponse")
		// log.Entry().Info(taskId, "taskId")
		// if readErr != nil {
		// 	return errors.Wrap(readErr, "Task Id not found. HTTP response body could not be read.")
		// }
		// deploymentError := pollValueMappingDeploymentStatus(string(taskId), retryCountVM, config, httpClient, serviceKey.OAuth.Host)
		// return deploymentError
	}
	responseBody, readErr := ioutil.ReadAll(deployResp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, response status code: %v", deployResp.StatusCode)
	}
	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code : %v", string(responseBody), deployResp.StatusCode)
	return errors.Errorf("value mapping deployment failed, response Status code: %v", deployResp.StatusCode)
}

// pollValueMappingDeploymentStatus - Poll the value mapping deployment status, return status or error details
func pollValueMappingDeploymentStatus(taskId string, retryCountVM int, config *valueMappingDeployOptions, httpClient piperhttp.Sender, apiHost string) error {

	if retryCountVM <= 0 {
		return errors.New("failed to start value mapping artifact after retrying several times")
	}
	deployStatus, err := getValueMappingDeployStatus(config, httpClient, apiHost, taskId)
	if err != nil {
		return err
	}

	//if artifact starting, then retry based on provided retry count
	//with specific delay between each retry
	if deployStatus == "DEPLOYING" {
		// Calling Sleep method
		sleepTime := int(retryCountVM * 3)
		time.Sleep(time.Duration(sleepTime) * time.Second)
		retryCountVM--
		return pollValueMappingDeploymentStatus(taskId, retryCountVM, config, httpClient, apiHost)
	}

	//if artifact started, then just return
	if deployStatus == "SUCCESS" {
		return nil
	}

	//if error return immediately with error details
	if deployStatus == "FAIL" || deployStatus == "FAIL_ON_LICENSE_ERROR" {
		resp, err := getValueMappingDeployError(config, httpClient, apiHost)
		if err != nil {
			return err
		}
		resp = "Error"
		return errors.New(resp)
	}
	return nil
}

// GetHTTPErrorMessage - Return HTTP failure message
func getVMHTTPErrorMessage(httpErr error, response *http.Response, httpMethod, statusURL string) (string, error) {
	responseBody, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return "", errors.Wrapf(readErr, "HTTP response body could not be read, response status code: %v", response.StatusCode)
	}
	log.Entry().Errorf("a HTTP error occurred! Response body: %v, response status code: %v", string(responseBody), response.StatusCode)
	return "", errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, statusURL, responseBody)
}

// getValueMappingDeployStatus - Get integration artifact Deploy Status
func getValueMappingDeployStatus(config *valueMappingDeployOptions, httpClient piperhttp.Sender, apiHost string, taskId string) (string, error) {
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
			WithField("ValueMappingID", config.ValueMappingID).
			Info("Successfully started value mapping artefact in CPI runtime")

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
		return getVMHTTPErrorMessage(httpErr, deployStatusResp, httpMethod, deployStatusURL)
	}
	return "", errors.Errorf("failed to get Value Mapping artefact runtime status, response Status code: %v", deployStatusResp.StatusCode)
}

// getIntegrationArtifactDeployError - Get integration artifact deploy error details
func getValueMappingDeployError(config *valueMappingDeployOptions, httpClient piperhttp.Sender, apiHost string) (string, error) {
	httpMethod := "GET"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	errorStatusURL := fmt.Sprintf("%s/api/v1/IntegrationRuntimeArtifacts('%s')/ErrorInformation/$value", apiHost, config.ValueMappingID)
	errorStatusResp, httpErr := httpClient.SendRequest(httpMethod, errorStatusURL, nil, header, nil)

	if errorStatusResp != nil && errorStatusResp.Body != nil {
		defer errorStatusResp.Body.Close()
	}

	if errorStatusResp == nil {
		return "", errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if errorStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("ValueMappingID", config.ValueMappingID).
			Info("Successfully retrieved value mapping artefact deploy error details")
		responseBody, readErr := ioutil.ReadAll(errorStatusResp.Body)
		if readErr != nil {
			return "", errors.Wrapf(readErr, "HTTP response body could not be read, response status code: %v", errorStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), errorStatusResp.StatusCode)
		errorDetails := string(responseBody)
		return errorDetails, nil
	}
	if httpErr != nil {
		return getHTTPErrorMessage(httpErr, errorStatusResp, httpMethod, errorStatusURL)
	}
	return "", errors.Errorf("failed to get value mapping artefact deploy error details, response Status code: %v", errorStatusResp.StatusCode)
}
