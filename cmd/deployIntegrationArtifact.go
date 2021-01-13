package cmd

import (
	"fmt"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/command"
	cpi "github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
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
	clientOptions := piperhttp.ClientOptions{}
	httpClient.SetOptions(clientOptions)
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

	if deployResp.StatusCode == 202 {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("successfully deployed into CPI runtime")
		return nil
	}

	log.Entry().Errorf("a HTTP error occurred! Response Status Code: %v", deployResp.StatusCode)
	return errors.Errorf("Integration Flow deployment failed, Response Status code: %v", deployResp.StatusCode)
}
