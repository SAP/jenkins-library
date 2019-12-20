package checkmarx

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// AuthToken - Structure to store OAuth2 token
type AuthToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// Preset - Project's Preset
type Preset struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	OwnerName string `json:"ownerName"`
	Link      Link   `json:"link"`
}

// Scan - Scan Structure
type Scan struct {
	ID   int  `json:"id"`
	Link Link `json:"link"`
}

// ResultsStatistics - ResultsStatistics Structure
type ResultsStatistics struct {
	High   int `json:"highSeverity"`
	Medium int `json:"mediumSeverity"`
	Low    int `json:"lowSeverity"`
	Info   int `json:"infoSeverity"`
}

// ScanStatus - ScanStatus Structure
type ScanStatus struct {
	ID       int    `json:"id"`
	Link     Link   `json:"link"`
	Status   Status `json:"status"`
	ScanType string `json:"scanType"`
	Comment  string `json:"comment"`
}

// Status - Status Structure
type Status struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Project - Project Structure
type Project struct {
	ID                 int                `json:"id"`
	TeamID             string             `json:"teamId"`
	Name               string             `json:"name"`
	IsPublic           bool               `json:"isPublic"`
	SourceSettingsLink SourceSettingsLink `json:"sourceSettingsLink"`
	Link               Link               `json:"link"`
}

// Team - Team Structure
type Team struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
}

// Link - Link Structure
type Link struct {
	Rel string `json:"rel"`
	URI string `json:"uri"`
}

// SourceSettingsLink - SourceSettingsLink Structure
type SourceSettingsLink struct {
	Type string `json:"type"`
	Rel  string `json:"rel"`
	URI  string `json:"uri"`
}

// System is the client communicating with the Checkmarx backend
type System struct {
	serverURL string
	username  string
	password  string
	client    piperHttp.Uploader
	logger    *logrus.Entry
}

// NewSystem returns a new Checkmarx client for communicating with the backend
func NewSystem(serverURL, username, password string) (*System, error) {
	sys := &System{
		serverURL: serverURL,
		username:  username,
		password:  password,
		client:    &piperHttp.Client{},
		logger:    log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx"),
	}

	token, err := sys.getOAuth2Token()
	if err != nil {
		return sys, errors.Wrap(err, "error fetching oAuth token")
	}

	options := piperHttp.ClientOptions{
		Token: token,
	}
	sys.client.SetOptions(options)

	return sys, nil
}

func sendRequest(sys *System, method, url string, body io.Reader, header http.Header) ([]byte, error) {
	response, err := sys.client.SendRequest(method, sys.serverURL+"/CxRestAPI"+url, body, header, nil)
	if err != nil {
		sys.logger.Errorf("HTTP request failed with error: %s", err)
		return nil, err
	}

	if response.StatusCode >= 200 && response.StatusCode < 400 {
		data, _ := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		return data, nil
	}

	data, _ := ioutil.ReadAll(response.Body)
	sys.logger.Debugf("Body %s", data)
	response.Body.Close()
	sys.logger.Errorf("HTTP request failed with error %s", response.Status)
	return nil, errors.Errorf("Invalid HTTP status %v with with code %v received", response.Status, response.StatusCode)
}

func (sys *System) getOAuth2Token() (string, error) {
	body := url.Values{
		"username":      {sys.username},
		"password":      {sys.password},
		"grant_type":    {"password"},
		"scope":         {"sast_rest_api"},
		"client_id":     {"resource_owner_client"},
		"client_secret": {"014DF517-39D1-4453-B7B3-9930C563627C"},
	}
	header := http.Header{}
	header.Add("Content-type", "application/x-www-form-urlencoded")
	data, err := sendRequest(sys, http.MethodPost, "/auth/identity/connect/token", strings.NewReader(body.Encode()), header)
	if err != nil {
		return "", err
	}

	var token AuthToken
	json.Unmarshal(data, &token)
	return token.TokenType + " " + token.AccessToken, nil
}

// GetTeams returns the teams the user is assigned to
func (sys *System) GetTeams() []Team {
	sys.logger.Debug("Getting Teams...")
	var teams []Team

	data, err := sendRequest(sys, http.MethodGet, "/auth/teams", nil, nil)
	if err != nil {
		sys.logger.Errorf("Fetching teams failed: %s", err)
		return teams
	}

	json.Unmarshal(data, &teams)
	return teams
}

// GetProjects returns the projects defined in the Checkmarx backend which the user has access to
func (sys *System) GetProjects() []Project {
	sys.logger.Debug("Getting Projects...")
	var projects []Project

	data, err := sendRequest(sys, http.MethodGet, "/projects", nil, nil)
	if err != nil {
		sys.logger.Errorf("Fetching projects failed: %s", err)
		return projects
	}

	json.Unmarshal(data, &projects)
	return projects
}

// CreateProject creates a new project in the Checkmarx backend
func (sys *System) CreateProject(projectName string, teamID string) bool {

	jsonData := map[string]interface{}{
		"name":       projectName,
		"owningTeam": teamID,
		"isPublic":   true,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		sys.logger.Errorf("Error Marshal: %s", err)
		return false
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	_, err = sendRequest(sys, http.MethodPost, "/projects", bytes.NewBuffer(jsonValue), header)
	if err != nil {
		sys.logger.Errorf("Failed to create project: %s", err)
		return false
	}

	return true
}

// UploadProjectSourceCode zips and uploads the project sources for scanning
func (sys *System) UploadProjectSourceCode(projectID int, zipFile string) bool {
	sys.logger.Debug("Starting to upload files...")

	header := http.Header{}
	header.Add("Accept-Encoding", "gzip,deflate")
	header.Add("Accept", "text/plain")
	resp, err := sys.client.UploadFile(sys.serverURL+"/CxRestAPI/projects/"+strconv.Itoa(projectID)+"/sourceCode/attachments", zipFile, "zippedSource", header, nil)
	if err != nil {
		sys.logger.Errorf("Failed to uploaded zipped sources %s", err)
		return false
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sys.logger.Errorf("Error reading the response data %s", err)
		return false
	}

	resp.Body.Close()
	responseData := make(map[string]string)
	json.Unmarshal(data, &responseData)

	if resp.StatusCode == http.StatusNoContent {
		return true
	}

	sys.logger.Debugf("Body %s", data)
	resp.Body.Close()
	sys.logger.Errorf("Error writing the request's body: %s", resp.Status)
	return false
}

// UpdateProjectExcludeSettings updates the exclude configuration of the project
func (sys *System) UpdateProjectExcludeSettings(projectID int, excludeFolders string, excludeFiles string) bool {
	jsonData := map[string]string{
		"excludeFoldersPattern": excludeFolders,
		"excludeFilesPattern":   excludeFiles,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		sys.logger.Errorf("Error Marshal: %s", err)
		return false
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	_, err = sendRequest(sys, http.MethodPut, "/projects/"+strconv.Itoa(projectID)+"/sourceCode/excludeSettings", bytes.NewBuffer(jsonValue), header)
	if err != nil {
		sys.logger.Errorf("HTTP request failed with error: %s", err)
		return false
	}

	return true
}

// GetPresets loads the preset values defined in the Checkmarx backend
func (sys *System) GetPresets() []Preset {
	sys.logger.Debug("Getting Presets...")
	var presets []Preset

	data, err := sendRequest(sys, http.MethodGet, "/sast/presets", nil, nil)
	if err != nil {
		sys.logger.Errorf("Fetching presets failed: %s", err)
		return presets
	}

	json.Unmarshal(data, &presets)
	return presets
}

// UpdateProjectConfiguration updates the configuration of the project addressed by projectID
func (sys *System) UpdateProjectConfiguration(projectID int, presetID int, engineConfigurationID string) bool {
	engineConfigID, _ := strconv.Atoi(engineConfigurationID)
	jsonData := map[string]interface{}{
		"projectId":             projectID,
		"presetId":              presetID,
		"engineConfigurationId": engineConfigID,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		sys.logger.Errorf("Error Marshal: %s", err)
		return false
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	_, err = sendRequest(sys, http.MethodPost, "/sast/scanSettings", bytes.NewBuffer(jsonValue), nil)
	if err != nil {
		sys.logger.Errorf("HTTP request failed with error: %s", err)
		return false
	}

	return true
}

// ScanProject triggers a scan on the project addressed by projectID
func (sys *System) ScanProject(projectID int) (bool, Scan) {
	scan := Scan{}
	jsonData := map[string]interface{}{
		"projectId":     projectID,
		"isIncremental": false,
		"isPublic":      true,
		"forceScan":     true,
		"comment":       "Scan From Golang Script",
	}

	jsonValue, _ := json.Marshal(jsonData)

	header := http.Header{}
	header.Set("cxOrigin", "GolangScript")
	header.Set("Content-Type", "application/json")
	data, err := sendRequest(sys, http.MethodPost, "/sast/scans", bytes.NewBuffer(jsonValue), header)
	if err != nil {
		sys.logger.Errorf("Failed to trigger scan of project %v: %s", projectID, err)
		return false, scan
	}

	json.Unmarshal(data, &scan)
	return true, scan
}

// GetScanStatus returns the status of the scan addressed by scanID
func (sys *System) GetScanStatus(scanID int) string {
	var scanStatus ScanStatus

	data, err := sendRequest(sys, http.MethodGet, "/sast/scans/"+strconv.Itoa(scanID), nil, nil)
	if err != nil {
		sys.logger.Errorf("Failed to get scan status for scanID %v: %s", scanID, err)
		return ""
	}

	json.Unmarshal(data, &scanStatus)
	return scanStatus.Status.Name
}

// GetResults returns the results of the scan addressed by scanID
func (sys *System) GetResults(scanID int) ResultsStatistics {
	var results ResultsStatistics

	data, err := sendRequest(sys, http.MethodGet, "/sast/scans/"+strconv.Itoa(scanID)+"/resultsStatistics", nil, nil)
	if err != nil {
		sys.logger.Errorf("Failed to fetch scan results for scanID %v: %s", scanID, err)
		return results
	}

	json.Unmarshal(data, &results)
	return results
}

// GetTeamByName filters a team by its name
func (sys *System) GetTeamByName(teams []Team, teamName string) Team {
	for _, team := range teams {
		if team.FullName == teamName {
			return team
		}
	}
	return Team{}
}

// GetProjectByName filters a project by its name
func (sys *System) GetProjectByName(projects []Project, projectName string) Project {
	for _, project := range projects {
		if project.Name == projectName {
			return project
		}
	}
	return Project{}
}

// GetPresetByName filters a preset by its name
func (sys *System) GetPresetByName(presets []Preset, presetName string) Preset {
	for _, preset := range presets {
		if preset.Name == presetName {
			return preset
		}
	}
	return Preset{}
}
