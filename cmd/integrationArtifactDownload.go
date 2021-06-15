package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type integrationArtifactDownloadUtils interface {
	command.ExecRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The integrationArtifactDownloadUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type integrationArtifactDownloadUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to integrationArtifactDownloadUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// integrationArtifactDownloadUtilsBundle and forward to the implementation of the dependency.
}

func newIntegrationArtifactDownloadUtils() integrationArtifactDownloadUtils {
	utils := integrationArtifactDownloadUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func integrationArtifactDownload(config integrationArtifactDownloadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runIntegrationArtifactDownload(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runIntegrationArtifactDownload(config *integrationArtifactDownloadOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {
	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/zip")
	serviceKey, err := cpi.ReadCpiServiceKey(config.ServiceKey)
	if err != nil {
		return err
	}
	downloadArtifactURL := fmt.Sprintf("%s/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')/$value", serviceKey.Host, config.IntegrationFlowID, config.IntegrationFlowVersion)
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.Uaa.OAuthTokenProviderURL, Username: serviceKey.Uaa.ClientId, Password: serviceKey.Uaa.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "GET"
	downloadResp, httpErr := httpClient.SendRequest(httpMethod, downloadArtifactURL, nil, header, nil)
	if httpErr != nil {
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error", httpMethod, downloadArtifactURL)
	}
	if downloadResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}
	contentDisposition := downloadResp.Header.Get("Content-Disposition")
	disposition, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return errors.Wrapf(err, "failed to read filename from http response headers, Content-Disposition "+disposition)
	}
	filename := params["filename"]

	if downloadResp != nil && downloadResp.Body != nil {
		defer downloadResp.Body.Close()
	}

	if downloadResp.StatusCode == 200 {
		workspaceRelativePath := config.DownloadPath
		err = os.MkdirAll(workspaceRelativePath, 0755)
		if err != nil {
			return errors.Wrapf(err, "Failed to create workspace directory")
		}
		zipFileName := filepath.Join(workspaceRelativePath, filename)
		file, err := os.Create(zipFileName)
		if err != nil {
			return errors.Wrapf(err, "Failed to create integration flow artifact file")
		}
		io.Copy(file, downloadResp.Body)
		return nil
	}
	responseBody, readErr := ioutil.ReadAll(downloadResp.Body)

	if readErr != nil {
		return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code : %v", downloadResp.StatusCode)
	}

	log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code : %v", responseBody, downloadResp.StatusCode)
	return errors.Errorf("Integration Flow artifact download failed, Response Status code: %v", downloadResp.StatusCode)
}
