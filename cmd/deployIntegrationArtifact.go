package cmd

import (
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"

	"github.com/Jeffail/gabs/v2"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type deployIntegrationArtifactUtils interface {
	command.ExecRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The deployIntegrationArtifactUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type deployIntegrationArtifactUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to deployIntegrationArtifactUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// deployIntegrationArtifactUtilsBundle and forward to the implementation of the dependency.
}

func newDeployIntegrationArtifactUtils() deployIntegrationArtifactUtils {
	utils := deployIntegrationArtifactUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func deployIntegrationArtifact(config deployIntegrationArtifactOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newDeployIntegrationArtifactUtils()
	utils.Stdout(log.Writer())
	httpClient := &piperhttp.Client{}
	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runDeployIntegrationArtifact(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runDeployIntegrationArtifact(config *deployIntegrationArtifactOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {
	cookieJar, cookieErr := cookiejar.New(nil)
	if cookieErr != nil {
		return errors.Wrap(cookieErr, "creating a cookie jar failed")
	}
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
	}
	httpClient.SetOptions(clientOptions)

	header := make(http.Header)
	header.Add("Accept", "application/json")

	deployURL := config.Host +
		"/api/v1/DeployIntegrationDesigntimeArtifact?Id=" + "'" + config.IntegrationFlowID +
		"'" + "&Version=" + "'" + config.IntegrationFlowVersion + "'"

	finalResult, err := getBearerTokenforDeployIntegrationArtifactCall(config, httpClient)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = "Bearer " + finalResult
	httpClient.SetOptions(clientOptions)

	deployResp, httpErr := httpClient.SendRequest("POST", deployURL, nil, header, nil)

	defer func() {
		if deployResp != nil && deployResp.Body != nil {
			deployResp.Body.Close()
		}
	}()

	if deployResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if deployResp.StatusCode == 202 {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("successfully deployed in to CPI runtime")
		return nil
	}

	log.Entry().Errorf("a HTTP error occurred! Response Status Code: %v", deployResp.StatusCode)
	return errors.Wrap(httpErr, "Deploying the integration flow failed")
}

func getBearerTokenforDeployIntegrationArtifactCall(config *deployIntegrationArtifactOptions, httpClient piperhttp.Sender) (string, error) {

	clientOptions := piperhttp.ClientOptions{
		Username: config.Username,
		Password: config.Password,
	}
	httpClient.SetOptions(clientOptions)

	header := make(http.Header)
	header.Add("Accept", "application/json")
	tokenURL := config.OAuthTokenProviderURL + "?grant_type=client_credentials"
	method := "POST"
	resp, httpErr := httpClient.SendRequest(method, tokenURL, nil, header, nil)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp == nil {
		return "", errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	// for supporting tests
	// with httpMockGcts we want to try only on actual odata API calls but not for OAuth token fetch calls
	// so we pass Oauth token in advance and skip OAuth call for mock tests
	if resp.Header.Get("Authorization") != "" {
		result := resp.Header.Get("Authorization")
		return result, nil
	}

	if resp.StatusCode != 200 {
		return "", errors.Errorf("did not retrieve a valid HTTP response code: %v", httpErr)
	}

	bodyText, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return "", errors.Wrap(readErr, "HTTP response body could not be read")
	}
	jsonresponse, parsingErr := gabs.ParseJSON([]byte(bodyText))
	if parsingErr != nil {
		return "", errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %v", string(bodyText))
	}
	finalResult := jsonresponse.Path("access_token").Data().(string)
	return finalResult, nil
}
