package cmd

import (
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
	"github.com/pkg/errors"
)

type apiProviderDownloadUtils interface {
	command.ExecRunner
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileExists(filename string) (bool, error)
}

type apiProviderDownloadUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newApiProviderDownloadUtils() apiProviderDownloadUtils {
	utils := apiProviderDownloadUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func apiProviderDownload(config apiProviderDownloadOptions, telemetryData *telemetry.CustomData) {
	utils := newApiProviderDownloadUtils()
	httpClient := &piperhttp.Client{}
	err := runApiProviderDownload(&config, telemetryData, httpClient, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runApiProviderDownload(config *apiProviderDownloadOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender, utils apiProviderDownloadUtils) error {
	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/json")
	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}
	downloadArtifactURL := fmt.Sprintf("%s/apiportal/api/1.0/Management.svc/APIProviders('%s')", serviceKey.OAuth.Host, config.APIProviderName)
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
	if downloadResp != nil && downloadResp.Body != nil {
		defer downloadResp.Body.Close()
	}
	if downloadResp.StatusCode == 200 {
		jsonFilePath := config.DownloadPath
		content, err := io.ReadAll(downloadResp.Body)
		if err != nil {
			return err
		}
		err = utils.FileWrite(jsonFilePath, content, 0775)
		if err != nil {
			return err
		}
	}
	return nil
}
