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
	deployURL := fmt.Sprintf("%s/api/v1/DeployIntegrationDesigntimeArtifact?Id='%s'&Version='%s'", config.Host, config.IntegrationFlowID, config.IntegrationFlowVersion)
	tokenParameters := cpi.TokenParameters{TokenURL: config.OAuthTokenProviderURL, Username: config.Username, Password: config.Password, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
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

	if deployResp != nil && deployResp.Body != nil {
		defer deployResp.Body.Close()
	}

	if deployResp == nil {
		return errors.Errorf("did not retrieve a HTTP response")
	}

	if deployResp.StatusCode == http.StatusAccepted {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("successfully deployed into CPI runtime")
		runs := 14
		error := PolliFlowDeploymentStatus(runs, config, httpClient)
		return error
	}
	responseBody, readErr := ioutil.ReadAll(deployResp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code : %v", deployResp.StatusCode)
	}

	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code : %v", responseBody, deployResp.StatusCode)
	return errors.Errorf("Integration Flow deployment failed, Response Status code: %v", deployResp.StatusCode)
}

//PolliFlowDeploymentStatus - Poll the integration flow deployment status,retrun status or error details
func PolliFlowDeploymentStatus(iter int, config *integrationArtifactDeployOptions, httpClient piperhttp.Sender) error {

	if iter <= 0 {
		return nil
	}
	deployStatus, err := GetIntegrationArtifactDeployStatus(config, httpClient)
	if err != nil {
		return err
	}

	//if artifact starting, then retry based on provided retry count
	//with specific delay between each retry
	if deployStatus == "STARTING" {
		// Calling Sleep method
		sleepTime := int(iter * 3)
		time.Sleep(time.Duration(sleepTime) * time.Second)
		newItrCount := iter - 1
		return PolliFlowDeploymentStatus(newItrCount, config, httpClient)
	}

	//if artifact started, then just return
	if deployStatus == "STARTED" {
		return nil
	}

	//if error return immediately with error details
	if deployStatus == "ERROR" {
		resp, err := GetIntegrationArtifactDeployError(config, httpClient)
		if err != nil {
			return err
		}
		return errors.New(resp)
	}
	return nil
}

//GetIntegrationArtifactDeployStatus - Get integration artifact Deploy Status
func GetIntegrationArtifactDeployStatus(config *integrationArtifactDeployOptions, httpClient piperhttp.Sender) (string, error) {
	httpMethod := "GET"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	header.Add("Accept", "application/json")
	deployStatusURL := fmt.Sprintf("%s/api/v1/IntegrationRuntimeArtifacts('%s')", config.Host, config.IntegrationFlowID)
	deployStatusResp, httpErr := httpClient.SendRequest(httpMethod, deployStatusURL, nil, header, nil)

	if deployStatusResp != nil && deployStatusResp.Body != nil {
		defer deployStatusResp.Body.Close()
	}

	if deployStatusResp == nil {
		return "", errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if deployStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("Successfully started integration flow artefact in CPI runtime")

		bodyText, readErr := ioutil.ReadAll(deployStatusResp.Body)
		if readErr != nil {
			return "", errors.Wrapf(readErr, "HTTP response body could not be read, Response status code : %v", deployStatusResp.StatusCode)
		}
		jsonResponse, parsingErr := gabs.ParseJSON([]byte(bodyText))
		if parsingErr != nil {
			return "", errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
		}
		deployStatus := jsonResponse.Path("d.Status").Data().(string)
		return deployStatus, nil
	}
	if httpErr != nil {
		responseBody, readErr := ioutil.ReadAll(deployStatusResp.Body)
		if readErr != nil {
			return "", errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", deployStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, deployStatusResp.StatusCode)
		return "", errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, deployStatusURL, responseBody)
	}
	return "", errors.Errorf("Failed to get Integration Flow artefact runtime status, Response Status code: %v", deployStatusResp.StatusCode)
}

//GetIntegrationArtifactDeployError - Get integration artifact deploy error details
func GetIntegrationArtifactDeployError(config *integrationArtifactDeployOptions, httpClient piperhttp.Sender) (string, error) {
	httpMethod := "GET"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	errorStatusURL := fmt.Sprintf("%s/api/v1/IntegrationRuntimeArtifacts('%s')/ErrorInformation/$value", config.Host, config.IntegrationFlowID)
	errorStatusResp, httpErr := httpClient.SendRequest(httpMethod, errorStatusURL, nil, header, nil)

	if errorStatusResp != nil && errorStatusResp.Body != nil {
		defer errorStatusResp.Body.Close()
	}

	if errorStatusResp == nil {
		return "", errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if errorStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("Successfully retrieved Integration Flow artefact deploy error details")
		responseBody, readErr := ioutil.ReadAll(errorStatusResp.Body)
		if readErr != nil {
			return "", errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", errorStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, errorStatusResp.StatusCode)
		errorDetails := string(responseBody)
		return errorDetails, nil
	}
	if httpErr != nil {
		responseBody, readErr := ioutil.ReadAll(errorStatusResp.Body)
		if readErr != nil {
			return "", errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", errorStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, errorStatusResp.StatusCode)
		return "", errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, errorStatusURL, responseBody)
	}
	return "", errors.Errorf("Failed to get Integration Flow artefact deploy error details, Response Status code: %v", errorStatusResp.StatusCode)
}
