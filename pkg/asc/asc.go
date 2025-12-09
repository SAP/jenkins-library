package asc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
	"strconv"
	"strings"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type App struct {
	AppId    int    `json:"app_id"`
	AppName  string `json:"app_name"`
	BundleId string `json:"bundle_id"`
	JamfId   string `json:"jamf_id"`
}

type JamfAppInformationResponse struct {
	MobileDeviceApplication JamfMobileDeviceApplication `json:"mobile_device_application"`
}

type JamfMobileDeviceApplication struct {
	General JamfMobileDeviceApplicationGeneral `json:"general"`
}

type JamfMobileDeviceApplicationGeneral struct {
	Id int `json:"id"`
}

type CreateReleaseResponse struct {
	Status  string  `json:"status"`
	Message string  `json:"message"`
	LastID  int     `json:"lastID"`
	Data    Release `json:"data"`
}

type Release struct {
	ReleaseID    int       `json:"release_id"`
	AppID        int       `json:"app_id"`
	Version      string    `json:"version"`
	Description  string    `json:"description"`
	ReleaseDate  time.Time `json:"release_date"`
	SortOrder    any       `json:"sort_order"`
	Visible      bool      `json:"visible"`
	Created      time.Time `json:"created"`
	FileMetadata any       `json:"file_metadata"`
}

// SystemInstance is the client communicating with the ASC backend
type SystemInstance struct {
	serverURL string
	token     string
	client    *piperHttp.Client
	logger    *logrus.Entry
}

type System interface {
	GetAppById(appId string) (App, error)
	CreateRelease(ascAppId int, version string, description string, releaseDate string, visible bool) (CreateReleaseResponse, error)
	GetJamfAppInfo(bundleId string, jamfTargetSystem string) (JamfAppInformationResponse, error)
	UploadIpa(path string, jamfAppId int, jamfTargetSystem string, bundleId string, ascRelease Release) error
}

// NewSystemInstance returns a new ASC client for communicating with the backend
func NewSystemInstance(client *piperHttp.Client, serverURL, token string) (*SystemInstance, error) {
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
		Token:            fmt.Sprintf("Bearer %s", sys.token),
		TransportTimeout: time.Second * 15,
	}
	sys.client.SetOptions(options)

	return sys, nil
}

func sendRequest(sys *SystemInstance, method, url string, body io.Reader, header http.Header) ([]byte, error) {
	var requestBody io.Reader
	if body != nil {
		closer := io.NopCloser(body)
		bodyBytes, _ := io.ReadAll(closer)
		requestBody = bytes.NewBuffer(bodyBytes)
		defer closer.Close()
	}
	response, err := sys.client.SendRequest(method, fmt.Sprintf("%v/%v", sys.serverURL, url), requestBody, header, nil)
	if err != nil && (response == nil) {
		sys.logger.Errorf("HTTP request failed with error: %s", err)
		return nil, err
	}

	data, _ := io.ReadAll(response.Body)
	sys.logger.Debugf("Valid response body: %v", string(data))
	defer response.Body.Close()
	return data, nil
}

// GetAppById returns the app addressed by appId from the ASC backend
func (sys *SystemInstance) GetAppById(appId string) (App, error) {
	sys.logger.Debugf("Getting ASC App with ID %v...", appId)
	var app App

	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("api/v1/apps/%v", appId), nil, nil)
	if err != nil {
		return app, errors.Wrapf(err, "fetching app %v failed", appId)
	}

	json.Unmarshal(data, &app)
	return app, nil
}

// CreateRelease creates a release in ASC
func (sys *SystemInstance) CreateRelease(ascAppId int, version string, description string, releaseDate string, visible bool) (CreateReleaseResponse, error) {

	var createReleaseResponse CreateReleaseResponse

	if len(releaseDate) == 0 {
		currentTime := time.Now()
		releaseDate = currentTime.Format("01/02/2006")
	}

	jsonData := map[string]string{
		"version":      version,
		"description":  description,
		"release_date": releaseDate,
		"visible":      strconv.FormatBool(visible),
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return createReleaseResponse, errors.Wrap(err, "error marshalling release payload")
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")

	response, err := sendRequest(sys, http.MethodPost, fmt.Sprintf("api/v1/apps/%v/releases", ascAppId), bytes.NewBuffer(jsonValue), header)
	if err != nil {
		return createReleaseResponse, errors.Wrap(err, "creating release")
	}

	json.Unmarshal(response, &createReleaseResponse)
	return createReleaseResponse, nil
}

// GetJamfAppInfo fetches information about the app from Jamf
func (sys *SystemInstance) GetJamfAppInfo(bundleId string, jamfTargetSystem string) (JamfAppInformationResponse, error) {

	sys.logger.Debugf("Getting Jamf App Info by ID %v from jamf %v system...", bundleId, jamfTargetSystem)
	var jamfAppInformationResponse JamfAppInformationResponse

	data, err := sendRequest(sys, http.MethodPost, fmt.Sprintf("api/v1/jamf/%v/info?system=%v", bundleId, url2.QueryEscape(jamfTargetSystem)), nil, nil)
	if err != nil {
		return jamfAppInformationResponse, errors.Wrapf(err, "fetching jamf %v app info for %v failed", jamfTargetSystem, bundleId)
	}

	json.Unmarshal(data, &jamfAppInformationResponse)
	return jamfAppInformationResponse, nil

}

// UploadIpa uploads the ipa to ASC and therewith to Jamf
func (sys *SystemInstance) UploadIpa(path string, jamfAppId int, jamfTargetSystem string, bundleId string, ascRelease Release) error {

	url := fmt.Sprintf("%v/api/v1/jamf/%v/ipa?app_id=%v&version=%v&system=%v&release_id=%v&bundle_id=%v", sys.serverURL, jamfAppId, ascRelease.AppID, url2.QueryEscape(ascRelease.Version), url2.QueryEscape(jamfTargetSystem), ascRelease.ReleaseID, url2.QueryEscape(bundleId))
	_, err := sys.client.UploadFile(url, path, "file", nil, nil, "form")

	if err != nil {
		return errors.Wrap(err, "failed to upload ipa to asc")
	}

	return nil
}
