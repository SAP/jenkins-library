package checkmarx

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/SAP/jenkins-library/pkg/log"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
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

// Checkmarx is the client communicating with the Checkmarx backend
type Checkmarx struct {
	serverURL string
	client    piperHttp.Client
	logger    *logrus.Entry
}

// NewCheckmarx returns a new Checkmarx client for communicating with the backend
func NewCheckmarx(serverURL, username, password string ) (*Checkmarx, error) {
	cmx := Checkmarx {
		serverURL: serverURL,
		logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx"),
		client: piperHttp.Client {},
	}

	token, err := getOAuth2Token(serverURL, username, password)
	if err != nil {
		return &cmx, errors.Wrap(err, "error fetching oAuth token")
	}

	options := piperHttp.ClientOptions {
		Token: token,
	}
	
	cmx.client.SetOptions(options)
	
	return &cmx, nil
}

func getOAuth2Token(serverURL, username, password string) (string, error) {
	resp, err := http.PostForm(serverURL+"auth/identity/connect/token",
		url.Values{
			"username":      {username},
			"password":      {password},
			"grant_type":    {"password"},
			"scope":         {"sast_rest_api"},
			"client_id":     {"resource_owner_client"},
			"client_secret": {"014DF517-39D1-4453-B7B3-9930C563627C"},
		})
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		var token AuthToken
		json.Unmarshal(body, &token)
		return token.TokenType + " " + token.AccessToken, nil
	}
	data, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return "", errors.Errorf("invalid response status %v (%v) received when fetching oAuth token: %v", resp.Status, resp.StatusCode, data)
}

// GetTeams returns the teams the user is assigned to
func (cmx *Checkmarx) GetTeams() []Team {
	cmx.logger.Debug("Getting Teams...")
	var teams []Team

	resp, err := cmx.client.SendRequest(http.MethodGet, cmx.serverURL+"auth/teams", nil, nil, nil)
	if err != nil {
		cmx.logger.Errorf("HTTP request failed with error: %s", err)
		return teams
	}

	if resp.StatusCode == http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(data, &teams)
		return teams
	}

	data, _ := ioutil.ReadAll(resp.Body)

	cmx.logger.Debugf("Body %s", data)
	resp.Body.Close()
	cmx.logger.Errorf("HTTP request failed with error %s", resp.Status)
	return teams
}

// GetProjects returns the projects defined in the Checkmarx backend which the user has access to
func (cmx *Checkmarx) GetProjects() []Project {
	cmx.logger.Debug("Getting Projects...")
	var projects []Project

	resp, err := cmx.client.SendRequest(http.MethodGet, cmx.serverURL+"projects", nil, nil, nil)
	if err != nil {
		cmx.logger.Errorf("HTTP request failed with error: %s", err)
		return projects
	}

	if resp.StatusCode == http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(data, &projects)
		return projects
	}

	data, _ := ioutil.ReadAll(resp.Body)

	cmx.logger.Debugf("Body %s", data)
	resp.Body.Close()
	cmx.logger.Errorf("HTTP request failed with error %s", resp.Status)
	return projects
}

// CreateProject creates a new project in the Checkmarx backend
func (cmx *Checkmarx) CreateProject(projectName string, teamID string) bool {

	jsonData := map[string]interface{}{
		"name":       projectName,
		"owningTeam": teamID,
		"isPublic":   true,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		cmx.logger.Errorf("Error Marshal: %s", err)
		return false
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	resp, err := cmx.client.SendRequest(http.MethodPost, cmx.serverURL+"projects", bytes.NewBuffer(jsonValue), header, nil)
	if err != nil {
		cmx.logger.Errorf("HTTP request failed with error: %s", err)
		return false
	}

	if resp.StatusCode == http.StatusCreated {
		return true
	}

	data, _ := ioutil.ReadAll(resp.Body)

	cmx.logger.Debugf("Body %s", data)
	resp.Body.Close()
	cmx.logger.Errorf("The HTTP request failed with error %s", resp.Status)
	return false
}


// UploadProjectSourceCode zips and uploads the project sources for scanning
func (cmx *Checkmarx) UploadProjectSourceCode(projectID int, zipFile string) bool {

	cmx.logger.Debug("Starting to upload files...")
	
	var header http.Header
	header.Add("Accept-Encoding", "gzip,deflate")
	header.Add("Accept", "text/plain")
	resp, err := cmx.client.UploadFile(cmx.serverURL+"projects/"+strconv.Itoa(projectID)+"/sourceCode/attachments", zipFile, "zippedSource", header, nil)
	if err != nil {
		cmx.logger.Errorf("The HTTP request failed with error %s", err)
		return false
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		cmx.logger.Errorf("Error reading the response data %s", err)
		return false
	}

	resp.Body.Close()
	responseData := make(map[string]string)
	json.Unmarshal(data, &responseData)

	if resp.StatusCode == http.StatusNoContent {
		return true
	}

	cmx.logger.Debugf("Body %s", data)
	resp.Body.Close()
	cmx.logger.Errorf("Error writting the request's body: ", resp.Status)
	return false
}

// UpdateProjectExcludeSettings updates the exclude configuration of the project
func (cmx *Checkmarx) UpdateProjectExcludeSettings(projectID int, excludeFolders string, excludeFiles string) bool {
	jsonData := map[string]string{
		"excludeFoldersPattern": excludeFolders,
		"excludeFilesPattern":   excludeFiles,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		cmx.logger.Errorf("Error Marshal: %s", err)
		return false
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	resp, err := cmx.client.SendRequest(http.MethodPut, cmx.serverURL+"projects/"+strconv.Itoa(projectID)+"/sourceCode/excludeSettings", bytes.NewBuffer(jsonValue), header, nil)
	if err != nil {
		cmx.logger.Errorf("HTTP request failed with error: %s", err)
		return false
	}

	if resp.StatusCode == http.StatusOK {
		return true
	}

	data, _ := ioutil.ReadAll(resp.Body)

	cmx.logger.Debugf("Body %s", data)
	resp.Body.Close()
	cmx.logger.Errorf("The HTTP request failed with error %s", resp.Status)
	return false
}

// GetPresets loads the preset values defined in the Checkmarx backend
func (cmx *Checkmarx) GetPresets() []Preset {
	cmx.logger.Debug("Getting Presets...")
	var presets []Preset

	resp, err := cmx.client.SendRequest(http.MethodGet, cmx.serverURL+"sast/presets", nil, nil, nil)
	if err != nil {
		cmx.logger.Errorf("HTTP request failed with error: %s", err)
		return presets
	}

	if resp.StatusCode == http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(data, &presets)
		return presets
	}

	data, _ := ioutil.ReadAll(resp.Body)

	cmx.logger.Debugf("Body %s", data)
	resp.Body.Close()
	cmx.logger.Errorf("The HTTP request failed with error %s", resp.Status)
	return presets
}

// UpdateProjectConfiguration updates the configuration of the project addressed by projectID
func (cmx *Checkmarx) UpdateProjectConfiguration(projectID int, presetID int, engineConfigurationID string) bool {
	engineConfigID, _ := strconv.Atoi(engineConfigurationID)
	jsonData := map[string]interface{}{
		"projectId":             projectID,
		"presetId":              presetID,
		"engineConfigurationId": engineConfigID,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		cmx.logger.Errorf("Error Marshal: %s", err)
		return false
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	resp, err := cmx.client.SendRequest(http.MethodPost, cmx.serverURL+"sast/scanSettings", bytes.NewBuffer(jsonValue), nil, nil)
	if err != nil {
		cmx.logger.Errorf("HTTP request failed with error: %s", err)
		return false
	}

	if resp.StatusCode == http.StatusOK {
		return true
	}

	data, _ := ioutil.ReadAll(resp.Body)

	cmx.logger.Debugf("Body %s", data)
	resp.Body.Close()
	cmx.logger.Errorf("The HTTP request failed with error %s", resp.Status)
	return false
}

// ScanProject triggers a scan on the project addressed by projectID
func (cmx *Checkmarx) ScanProject(projectID int) (bool, Scan) {
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
	resp, err := cmx.client.SendRequest(http.MethodPost, cmx.serverURL+"sast/scans", bytes.NewBuffer(jsonValue), header, nil)
	if err != nil {
		cmx.logger.Errorf("HTTP request failed with error: %s", err)
		return false, scan
	}

	if resp.StatusCode == http.StatusCreated {
		data, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(data, &scan)
		return true, scan
	}

	cmx.logger.Debug(resp.Status)
	return false, scan
}

// GetScanStatus returns the status of the scan addressed by scanID
func (cmx *Checkmarx) GetScanStatus(scanID int) string {
	var scanStatus ScanStatus

	resp, err := cmx.client.SendRequest(http.MethodGet, cmx.serverURL+"sast/scans/"+strconv.Itoa(scanID), nil, nil, nil)
	if err != nil {
		cmx.logger.Errorf("The HTTP request failed with error %s", err)
		return ""
	}

	if resp.StatusCode == http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(data, &scanStatus)

		return scanStatus.Status.Name
	}

	data, _ := ioutil.ReadAll(resp.Body)

	cmx.logger.Debugf("Body %s", data)
	resp.Body.Close()
	cmx.logger.Errorf("The HTTP request failed with error %s", resp.Status)
	return ""
}

// GetResults returns the results of the scan addressed by scanID
func (cmx *Checkmarx) GetResults(scanID int) ResultsStatistics {
	var results ResultsStatistics

	resp, err := cmx.client.SendRequest(http.MethodGet, cmx.serverURL+"sast/scans/"+strconv.Itoa(scanID)+"/resultsStatistics", nil, nil, nil)
	if err != nil {
		cmx.logger.Errorf("The HTTP request failed with error %s", err)
		return results
	}

	if resp.StatusCode == http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(data, &results)

		return results
	}

	data, _ := ioutil.ReadAll(resp.Body)

	cmx.logger.Debugf("Body %s", data)
	resp.Body.Close()
	cmx.logger.Errorf("The HTTP request failed with error %s", resp.Status)
	return results
}

// GetTeamByName filters a team by its name
func (cmx *Checkmarx) GetTeamByName(teams []Team, teamName string) Team {
	for _, team := range teams {
		if team.FullName == teamName {
			return team
		}
	}
	return Team{}
}

// GetProjectByName filters a project by its name
func (cmx *Checkmarx) GetProjectByName(projects []Project, projectName string) Project {
	for _, project := range projects {
		if project.Name == projectName {
			return project
		}
	}
	return Project{}
}

// GetPresetByName filters a preset by its name
func (cmx *Checkmarx) GetPresetByName(presets []Preset, presetName string) Preset {
	for _, preset := range presets {
		if preset.Name == presetName {
			return preset
		}
	}
	return Preset{}
}