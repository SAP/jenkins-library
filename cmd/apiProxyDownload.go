package cmd

import (
	"fmt"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func apiProxyDownload(config apiProxyDownloadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runApiProxyDownload(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runApiProxyDownload(config *apiProxyDownloadOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {
	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/zip")
	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}
	downloadArtifactURL := fmt.Sprintf("%s/apiportal/api/1.0/Transport.svc/APIProxies?name=%s", serviceKey.OAuth.Host, config.APIProxyName)
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL,
		Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := http.MethodGet
	downloadResp, httpErr := httpClient.SendRequest(httpMethod, downloadArtifactURL, nil, header, nil)
	if httpErr != nil {
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error", httpMethod, downloadArtifactURL)
	}
	if downloadResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}
	failureMessage := "Failed to download API Proxy artefact"
	httpFileDownloadRequestParameters := cpi.HttpFileDownloadRequestParameters{ErrMessage: failureMessage, FileDownloadPath: config.DownloadPath, Response: downloadResp}
	return cpi.HttpCPIUtils.HandleHTTPFileDownloadResponse(httpFileDownloadRequestParameters)
}
