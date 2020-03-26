package checkmarx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"encoding/xml"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
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

// ProjectCreateResult - ProjectCreateResult Structure
type ProjectCreateResult struct {
	ID   int  `json:"id"`
	Link Link `json:"link"`
}

// Report - Report Structure
type Report struct {
	ReportID int   `json:"reportId"`
	Links    Links `json:"links"`
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
	ID            int    `json:"id"`
	Link          Link   `json:"link"`
	Status        Status `json:"status"`
	ScanType      string `json:"scanType"`
	Comment       string `json:"comment"`
	IsIncremental bool   `json:"isIncremental"`
}

// Status - Status Structure
type Status struct {
	ID      int              `json:"id"`
	Name    string           `json:"name"`
	Details ScanStatusDetail `json:"details"`
}

// ScanStatusDetail - ScanStatusDetail Structure
type ScanStatusDetail struct {
	Stage string `json:"stage"`
	Step  string `json:"step"`
}

// ReportStatusResponse - ReportStatusResponse Structure
type ReportStatusResponse struct {
	Location    string       `json:"location"`
	ContentType string       `json:"contentType"`
	Status      ReportStatus `json:"status"`
}

// ReportStatus - ReportStatus Structure
type ReportStatus struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
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

// Links - Links Structure
type Links struct {
	Report Link `json:"report"`
	Status Link `json:"status"`
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

//DetailedResult - DetailedResult Structure
type DetailedResult struct {
	XMLName                  xml.Name `xml:"CxXMLResults"`
	InitiatorName            string   `xml:"InitiatorName,attr"`
	ScanID                   string   `xml:"ScanId,attr"`
	Owner                    string   `xml:"Owner,attr"`
	ProjectID                string   `xml:"ProjectId,attr"`
	ProjectName              string   `xml:"ProjectName,attr"`
	TeamFullPathOnReportDate string   `xml:"TeamFullPathOnReportDate,attr"`
	DeepLink                 string   `xml:"DeepLink,attr"`
	ScanStart                string   `xml:"ScanStart,attr"`
	Preset                   string   `xml:"Preset,attr"`
	ScanTime                 string   `xml:"ScanTime,attr"`
	LinesOfCodeScanned       string   `xml:"LinesOfCodeScanned,attr"`
	FilesScanned             string   `xml:"FilesScanned,attr"`
	ReportCreationTime       string   `xml:"ReportCreationTime,attr"`
	Team                     string   `xml:"Team,attr"`
	CheckmarxVersion         string   `xml:"CheckmarxVersion,attr"`
	ScanType                 string   `xml:"ScanType,attr"`
	SourceOrigin             string   `xml:"SourceOrigin,attr"`
	Visibility               string   `xml:"Visibility,attr"`
	Queries                  []Query  `xml:"Query"`
}

// Query - Query Structure
type Query struct {
	XMLName xml.Name `xml:"Query"`
	Results []Result `xml:"Result"`
}

// Result - Result Structure
type Result struct {
	XMLName       xml.Name `xml:"Result"`
	State         string   `xml:"state,attr"`
	Severity      string   `xml:"Severity,attr"`
	FalsePositive string   `xml:"FalsePositive,attr"`
}

// SystemInstance is the client communicating with the Checkmarx backend
type SystemInstance struct {
	serverURL string
	username  string
	password  string
	client    piperHttp.Uploader
	logger    *logrus.Entry
}

// System is the interface abstraction of a specific SystemIns
type System interface {
	FilterPresetByName(presets []Preset, presetName string) Preset
	FilterPresetByID(presets []Preset, presetID int) Preset
	FilterProjectByName(projects []Project, projectName string) Project
	FilterTeamByName(teams []Team, teamName string) Team
	FilterTeamByID(teams []Team, teamID string) Team
	DownloadReport(reportID int) (bool, []byte)
	GetReportStatus(reportID int) ReportStatusResponse
	RequestNewReport(scanID int, reportType string) (bool, Report)
	GetResults(scanID int) ResultsStatistics
	GetScanStatusAndDetail(scanID int) (string, ScanStatusDetail)
	GetScans(projectID int) (bool, []ScanStatus)
	ScanProject(projectID int, isIncremental, isPublic, forceScan bool) (bool, Scan)
	UpdateProjectConfiguration(projectID int, presetID int, engineConfigurationID string) bool
	UpdateProjectExcludeSettings(projectID int, excludeFolders string, excludeFiles string) bool
	UploadProjectSourceCode(projectID int, zipFile string) bool
	CreateProject(projectName string, teamID string) (bool, ProjectCreateResult)
	CreateBranch(projectID int, branchName string) int
	GetPresets() []Preset
	GetProjectByID(projectID int) (bool, Project)
	GetProjectsByNameAndTeam(projectName, teamID string) []Project
	GetProjects() []Project
	GetTeams() []Team
}

// NewSystemInstance returns a new Checkmarx client for communicating with the backend
func NewSystemInstance(client piperHttp.Uploader, serverURL, username, password string) (*SystemInstance, error) {
	loggerInstance := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx")
	sys := &SystemInstance{
		serverURL: serverURL,
		username:  username,
		password:  password,
		client:    client,
		logger:    loggerInstance,
	}

	token, err := sys.getOAuth2Token()
	if err != nil {
		return sys, errors.Wrap(err, "Error fetching oAuth token")
	}

	options := piperHttp.ClientOptions{
		Token:              token,
		MaxRequestDuration: 60 * time.Second,
	}
	sys.client.SetOptions(options)

	return sys, nil
}

func sendRequest(sys *SystemInstance, method, url string, body io.Reader, header http.Header) ([]byte, error) {
	return sendRequestInternal(sys, method, url, body, header, "200:399")
}

func sendRequestInternal(sys *SystemInstance, method, url string, body io.Reader, header http.Header, validStatusCodeRange string) ([]byte, error) {
	var requestBody io.Reader
	var requestBodyCopy io.Reader
	if body != nil {
		closer := ioutil.NopCloser(body)
		bodyBytes, _ := ioutil.ReadAll(closer)
		requestBody = bytes.NewBuffer(bodyBytes)
		requestBodyCopy = bytes.NewBuffer(bodyBytes)
		defer closer.Close()
	}
	response, err := sys.client.SendRequest(method, fmt.Sprintf("%v/cxrestapi%v", sys.serverURL, url), requestBody, header, nil)
	if err != nil {
		sys.recordRequestDetailsInErrorCase(requestBodyCopy, response)
		sys.logger.Errorf("HTTP request failed with error: %s", err)
		return nil, err
	}

	var validResponseCodeList []int
	values := strings.Split(validStatusCodeRange, ",")
	for _, value := range values {
		parts := strings.Split(value, ":")
		if len(parts) > 1 {
			lower, _ := strconv.Atoi(parts[0])
			upper, _ := strconv.Atoi(parts[1])
			for i := lower; i <= upper; i++ {
				validResponseCodeList = append(validResponseCodeList, i)
			}
		} else {
			validCode, _ := strconv.Atoi(value)
			validResponseCodeList = append(validResponseCodeList, validCode)
		}
	}

	if piperutils.ContainsInt(validResponseCodeList, response.StatusCode) {
		data, _ := ioutil.ReadAll(response.Body)
		sys.logger.Debugf("Valid response body: %v", string(data))
		defer response.Body.Close()
		return data, nil
	}
	sys.recordRequestDetailsInErrorCase(requestBodyCopy, response)
	sys.logger.Errorf("HTTP request failed with error %s", response.Status)
	return nil, errors.Errorf("Invalid HTTP status %v with with code %v received", response.Status, response.StatusCode)
}

func (sys *SystemInstance) recordRequestDetailsInErrorCase(requestBody io.Reader, response *http.Response) {
	if requestBody != nil {
		data, _ := ioutil.ReadAll(ioutil.NopCloser(requestBody))
		sys.logger.Errorf("Request body: %s", data)
	}
	if response != nil && response.Body != nil {
		data, _ := ioutil.ReadAll(response.Body)
		sys.logger.Errorf("Response body: %s", data)
		response.Body.Close()
	}
}

func (sys *SystemInstance) getOAuth2Token() (string, error) {
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
func (sys *SystemInstance) GetTeams() []Team {
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
func (sys *SystemInstance) GetProjects() []Project {
	return sys.GetProjectsByNameAndTeam("", "")
}

// GetProjectByID returns the project addressed by projectID from the Checkmarx backend which the user has access to
func (sys *SystemInstance) GetProjectByID(projectID int) (bool, Project) {
	sys.logger.Debugf("Getting Project with ID %v...", projectID)
	var project Project

	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/projects/%v", projectID), nil, nil)
	if err != nil {
		sys.logger.Errorf("Fetching projects failed: %s", err)
		return false, project
	}

	json.Unmarshal(data, &project)
	return true, project
}

// GetProjectsByNameAndTeam returns the project addressed by projectID from the Checkmarx backend which the user has access to
func (sys *SystemInstance) GetProjectsByNameAndTeam(projectName, teamID string) []Project {
	sys.logger.Debugf("Getting projects with name %v of team %v...", projectName, teamID)
	var projects []Project
	header := http.Header{}
	header.Set("Accept-Type", "application/json")
	var data []byte
	var err error
	if len(teamID) > 0 && len(projectName) > 0 {
		body := url.Values{
			"projectName": {projectName},
			"teamId":      {teamID},
		}
		data, err = sendRequestInternal(sys, http.MethodGet, fmt.Sprintf("/projects?%v", body.Encode()), nil, header, "200:399,404")
	} else {
		data, err = sendRequestInternal(sys, http.MethodGet, "/projects", nil, header, "200:399,404")
	}
	if err != nil {
		sys.logger.Errorf("Fetching projects failed: %s", err)
		return projects
	}

	json.Unmarshal(data, &projects)
	return projects
}

// CreateProject creates a new project in the Checkmarx backend
func (sys *SystemInstance) CreateProject(projectName string, teamID string) (bool, ProjectCreateResult) {
	var result ProjectCreateResult
	jsonData := map[string]interface{}{
		"name":       projectName,
		"owningTeam": teamID,
		"isPublic":   true,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		sys.logger.Errorf("Error Marshal: %s", err)
		return false, result
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")

	data, err := sendRequest(sys, http.MethodPost, "/projects", bytes.NewBuffer(jsonValue), header)
	if err != nil {
		sys.logger.Errorf("Failed to create project: %s", err)
		return false, result
	}

	json.Unmarshal(data, &result)
	return true, result
}

// CreateBranch creates a branch of an existing project in the Checkmarx backend
func (sys *SystemInstance) CreateBranch(projectID int, branchName string) int {
	jsonData := map[string]interface{}{
		"name": branchName,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		sys.logger.Errorf("Error Marshal: %s", err)
		return 0
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	data, err := sendRequest(sys, http.MethodPost, fmt.Sprintf("/projects/%v/branch", projectID), bytes.NewBuffer(jsonValue), header)
	if err != nil {
		sys.logger.Errorf("Failed to create project: %s", err)
		return 0
	}

	var scan Scan

	json.Unmarshal(data, &scan)
	return scan.ID
}

// UploadProjectSourceCode zips and uploads the project sources for scanning
func (sys *SystemInstance) UploadProjectSourceCode(projectID int, zipFile string) bool {
	sys.logger.Debug("Starting to upload files...")

	header := http.Header{}
	header.Add("Accept-Encoding", "gzip,deflate")
	header.Add("Accept", "text/plain")
	resp, err := sys.client.UploadFile(fmt.Sprintf("%v/cxrestapi/projects/%v/sourceCode/attachments", sys.serverURL, projectID), zipFile, "zippedSource", header, nil)
	if err != nil {
		sys.logger.Errorf("Failed to uploaded zipped sources %s", err)
		return false
	}

	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		sys.logger.Errorf("Error reading the response data %s", err)
		return false
	}

	responseData := make(map[string]string)
	json.Unmarshal(data, &responseData)

	if resp.StatusCode == http.StatusNoContent {
		return true
	}

	sys.logger.Debugf("Body %s", data)
	sys.logger.Errorf("Error writing the request's body: %s", resp.Status)
	return false
}

// UpdateProjectExcludeSettings updates the exclude configuration of the project
func (sys *SystemInstance) UpdateProjectExcludeSettings(projectID int, excludeFolders string, excludeFiles string) bool {
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
	_, err = sendRequest(sys, http.MethodPut, fmt.Sprintf("/projects/%v/sourceCode/excludeSettings", projectID), bytes.NewBuffer(jsonValue), header)
	if err != nil {
		sys.logger.Errorf("HTTP request failed with error: %s", err)
		return false
	}

	return true
}

// GetPresets loads the preset values defined in the Checkmarx backend
func (sys *SystemInstance) GetPresets() []Preset {
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
func (sys *SystemInstance) UpdateProjectConfiguration(projectID int, presetID int, engineConfigurationID string) bool {
	engineConfigID, _ := strconv.Atoi(engineConfigurationID)
	jsonData := map[string]interface{}{
		"projectId":             projectID,
		"presetId":              presetID,
		"engineConfigurationId": engineConfigID,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		sys.logger.Errorf("Error marshal: %s", err)
		return false
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	_, err = sendRequest(sys, http.MethodPost, "/sast/scanSettings", bytes.NewBuffer(jsonValue), header)
	if err != nil {
		sys.logger.Errorf("HTTP request failed with error: %s", err)
		return false
	}

	return true
}

// ScanProject triggers a scan on the project addressed by projectID
func (sys *SystemInstance) ScanProject(projectID int, isIncremental, isPublic, forceScan bool) (bool, Scan) {
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

// GetScans returns all scan status on the project addressed by projectID
func (sys *SystemInstance) GetScans(projectID int) (bool, []ScanStatus) {
	scans := []ScanStatus{}
	body := url.Values{
		"projectId": {fmt.Sprintf("%v", projectID)},
		"last":      {fmt.Sprintf("%v", 20)},
	}

	header := http.Header{}
	header.Set("cxOrigin", "GolangScript")
	header.Set("Accept-Type", "application/json")
	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/sast/scans?%v", body.Encode()), nil, header)
	if err != nil {
		sys.logger.Errorf("Failed to fetch scans of project %v: %s", projectID, err)
		return false, scans
	}

	json.Unmarshal(data, &scans)
	return true, scans
}

// GetScanStatusAndDetail returns the status of the scan addressed by scanID
func (sys *SystemInstance) GetScanStatusAndDetail(scanID int) (string, ScanStatusDetail) {
	var scanStatus ScanStatus

	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/sast/scans/%v", scanID), nil, nil)
	if err != nil {
		sys.logger.Errorf("Failed to get scan status for scanID %v: %s", scanID, err)
		return "Failed", ScanStatusDetail{}
	}

	json.Unmarshal(data, &scanStatus)
	return scanStatus.Status.Name, scanStatus.Status.Details
}

// GetResults returns the results of the scan addressed by scanID
func (sys *SystemInstance) GetResults(scanID int) ResultsStatistics {
	var results ResultsStatistics

	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/sast/scans/%v/resultsStatistics", scanID), nil, nil)
	if err != nil {
		sys.logger.Errorf("Failed to fetch scan results for scanID %v: %s", scanID, err)
		return results
	}

	json.Unmarshal(data, &results)
	return results
}

// RequestNewReport triggers the generation of a  report for a specific scan addressed by scanID
func (sys *SystemInstance) RequestNewReport(scanID int, reportType string) (bool, Report) {
	report := Report{}
	jsonData := map[string]interface{}{
		"scanId":     scanID,
		"reportType": reportType,
		"comment":    "Scan report triggered by Piper",
	}

	jsonValue, _ := json.Marshal(jsonData)

	header := http.Header{}
	header.Set("cxOrigin", "GolangScript")
	header.Set("Content-Type", "application/json")
	data, err := sendRequest(sys, http.MethodPost, "/reports/sastScan", bytes.NewBuffer(jsonValue), header)
	if err != nil {
		sys.logger.Errorf("Failed to trigger report generation for scan %v: %s", scanID, err)
		return false, report
	}

	json.Unmarshal(data, &report)
	return true, report
}

// GetReportStatus returns the status of the report generation process
func (sys *SystemInstance) GetReportStatus(reportID int) ReportStatusResponse {
	var response ReportStatusResponse

	header := http.Header{}
	header.Set("Accept", "application/json")
	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/reports/sastScan/%v/status", reportID), nil, header)
	if err != nil {
		sys.logger.Errorf("Failed to fetch report status for reportID %v: %s", reportID, err)
		return response
	}

	json.Unmarshal(data, &response)
	return response
}

// DownloadReport downloads the report addressed by reportID and returns the XML contents
func (sys *SystemInstance) DownloadReport(reportID int) (bool, []byte) {
	header := http.Header{}
	header.Set("Accept", "application/json")
	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/reports/sastScan/%v", reportID), nil, header)
	if err != nil {
		sys.logger.Errorf("Failed to download report with reportID %v: %s", reportID, err)
		return false, []byte{}
	}
	return true, data
}

// FilterTeamByName filters a team by its name
func (sys *SystemInstance) FilterTeamByName(teams []Team, teamName string) Team {
	for _, team := range teams {
		if team.FullName == teamName {
			return team
		}
	}
	return Team{}
}

// FilterTeamByID filters a team by its ID
func (sys *SystemInstance) FilterTeamByID(teams []Team, teamID string) Team {
	for _, team := range teams {
		if team.ID == teamID {
			return team
		}
	}
	return Team{}
}

// FilterProjectByName filters a project by its name
func (sys *SystemInstance) FilterProjectByName(projects []Project, projectName string) Project {
	for _, project := range projects {
		if project.Name == projectName {
			sys.logger.Debugf("Filtered project with name %v", project.Name)
			return project
		}
	}
	return Project{}
}

// FilterPresetByName filters a preset by its name
func (sys *SystemInstance) FilterPresetByName(presets []Preset, presetName string) Preset {
	for _, preset := range presets {
		if preset.Name == presetName {
			return preset
		}
	}
	return Preset{}
}

// FilterPresetByID filters a preset by its name
func (sys *SystemInstance) FilterPresetByID(presets []Preset, presetID int) Preset {
	for _, preset := range presets {
		if preset.ID == presetID {
			return preset
		}
	}
	return Preset{}
}
