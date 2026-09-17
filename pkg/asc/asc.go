package asc

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"errors"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/sirupsen/logrus"
)

// DeployRequest holds the data used to deploy an app to ASC (and therewith to Jamf)
type DeployRequest struct {
	BundleID     string
	FilePath     string
	Version      string
	Description  string
	ReleaseDate  string
	Visible      bool
	TargetSystem string
	User         string
}

// SystemInstance is the client communicating with the ASC backend
type SystemInstance struct {
	serverURL string
	token     string
	client    *piperHttp.Client
	logger    *logrus.Entry
}

type System interface {
	DeployApp(request DeployRequest) error
}

// NewSystemInstance returns a new ASC client for communicating with the backend
func NewSystemInstance(client *piperHttp.Client, serverURL, token string, timeout time.Duration) (*SystemInstance, error) {
	loggerInstance := log.Entry().WithField("package", "SAP/jenkins-library/pkg/asc")

	if len(serverURL) == 0 {
		return nil, errors.New("serverUrl is not set but required")
	}

	if len(token) == 0 {
		return nil, errors.New("AppToken is not set but required")
	}

	sys := &SystemInstance{
		serverURL: strings.TrimSuffix(serverURL, "/"),
		token:     token,
		client:    client,
		logger:    loggerInstance,
	}

	log.RegisterSecret(token)

	options := piperHttp.ClientOptions{
		Token:              fmt.Sprintf("Bearer %s", sys.token),
		TransportTimeout:   timeout,
		MaxRequestDuration: timeout,
	}
	sys.client.SetOptions(options)

	return sys, nil
}

// DeployApp uploads the app binary together with the release information to ASC in a single request.
// ASC creates the release note and forwards the binary to Jamf.
func (sys *SystemInstance) DeployApp(request DeployRequest) error {
	releaseDate := request.ReleaseDate
	if len(releaseDate) == 0 {
		releaseDate = time.Now().UTC().Format("2006-01-02")
	} else if t, err := time.Parse("01/02/2006", releaseDate); err == nil {
		// auto-convert legacy MM/DD/YYYY format to YYYY-MM-DD
		releaseDate = t.Format("2006-01-02")
	}

	fileHandle, err := piperutils.Files{}.Open(request.FilePath)
	if err != nil {
		return fmt.Errorf("unable to locate file %v: %w", request.FilePath, err)
	}
	defer fileHandle.Close()

	url := fmt.Sprintf("%v/api/public/apps/%v/deploy", sys.serverURL, request.BundleID)

	formFields := map[string]string{
		"version":      request.Version,
		"description":  request.Description,
		"release_date": releaseDate,
		"visible":      strconv.FormatBool(request.Visible),
		"system":       request.TargetSystem,
		"user":         request.User,
	}

	sys.logger.Infof("Deploying app to ASC")
	sys.logger.Infof("  URL:          %v", url)
	sys.logger.Infof("  bundleId:     %v", request.BundleID)
	sys.logger.Infof("  file:         %v", request.FilePath)
	sys.logger.Infof("  version:      %v", request.Version)
	sys.logger.Infof("  description:  %v", request.Description)
	sys.logger.Infof("  release_date: %v", releaseDate)
	sys.logger.Infof("  visible:      %v", request.Visible)
	sys.logger.Infof("  system:       %v", request.TargetSystem)
	sys.logger.Infof("  user:         %v", request.User)

	response, err := sys.client.Upload(piperHttp.UploadRequestData{
		Method:        http.MethodPost,
		URL:           url,
		File:          request.FilePath,
		FileFieldName: "file",
		FormFields:    formFields,
		FileContent:   fileHandle,
		UploadType:    "form",
	})
	if err != nil {
		return fmt.Errorf("failed to deploy app to asc: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("deploy request failed with status %v: %v", response.StatusCode, string(body))
	}

	return nil
}
