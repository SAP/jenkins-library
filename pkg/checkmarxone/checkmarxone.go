package checkmarxOne

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	//"strconv"
	"strings"
	"time"

	//"encoding/xml"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ReportsDirectory defines the subfolder for the Checkmarx reports which are generated
const ReportsDirectory = "checkmarxOne"
const cxOrigin = "GolangScript"

// AuthToken - Structure to store OAuth2 token
// Updated for Cx1
type AuthToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	//   RefreshExpiresIn        int    `json:"refresh_expires_in"`
	//   NotBeforePolicy         int    `json:"not-before-policy"`
	//   Scope                   string `json:"scope"`
}

type Application struct {
	ApplicationID string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Criticality   uint              `json:"criticality"`
	Rules         []ApplicationRule `json:"rules"`
	Tags          map[string]string `json:"tags"`
	CreatedAt     string            `json:"createdAt"`
	UpdatedAt     string            `json:"updatedAt"`
}

type ApplicationRule struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Preset - Project's Preset
// Updated for Cx1
type Preset struct {
	PresetID int    `json:"id"`
	Name     string `json:"name"`
}

// Project - Project Structure
// Updated for Cx1
type Project struct {
	ProjectID    string            `json:"id"`
	Name         string            `json:"name"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
	Groups       []string          `json:"groups"`
	Applications []string          `json:"applicationIds"`
	Tags         map[string]string `json:"tags"`
	RepoUrl      string            `json:"repoUrl"`
	MainBranch   string            `json:"mainBranch"`
	Origin       string            `json:"origin"`
	Criticality  int               `json:"criticality"`
}

// New for Cx1
// These settings are higher-level settings that define how an engine should run, for example "multi-language" mode or setting a preset.
type ProjectConfigurationSetting struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Category        string `json:"category"`
	OriginLevel     string `json:"originLevel"`
	Value           string `json:"value"`
	ValueType       string `json:"valuetype"`
	ValueTypeParams string `json:"valuetypeparams"`
	AllowOverride   bool   `json:"allowOverride"`
}

type Query struct {
	QueryID            uint64 `json:"queryID,string"`
	Name               string `json:"queryName"`
	Group              string
	Language           string
	Severity           string
	CweID              int64
	QueryDescriptionID int64
	Custom             bool
}

// ReportStatus - ReportStatus Structure
// Updated for Cx1
type ReportStatus struct {
	ReportID  string `json:"reportId"`
	Status    string `json:"status"`
	ReportURL string `json:"url"`
}

type ResultsPredicates struct {
	PredicateID  string `json:"ID"`
	SimilarityID int64  `json:"similarityId,string"`
	ProjectID    string `json:"projectId"`
	State        string `json:"state"`
	Comment      string `json:"comment"`
	Severity     string `json:"severity"`
	CreatedBy    string
	CreatedAt    string
}

// Scan - Scan Structure
// updated for Cx1
type Scan struct {
	ScanID        string              `json:"id"`
	Status        string              `json:"status"`
	StatusDetails []ScanStatusDetails `json:"statusDetails"`
	Branch        string              `json:"branch"`
	CreatedAt     string              `json:"createdAt"`
	UpdatedAt     string              `json:"updatedAt"`
	ProjectID     string              `json:"projectId"`
	ProjectName   string              `json:"projectName"`
	UserAgent     string              `json:"userAgent"`
	Initiator     string              `json:"initiator"`
	Tags          map[string]string   `json:"tags"`
	Metadata      struct {
		Type    string              `json:"type"`
		Configs []ScanConfiguration `json:"configs"`
	} `json:"metadata"`
	Engines      []string `json:"engines"`
	SourceType   string   `json:"sourceType"`
	SourceOrigin string   `json:"sourceOrigin"`
}

// New for Cx1: ScanConfiguration - list of key:value pairs used to configure the scan for each scan engine
// This is specifically for scan-level configurations like "is incremental" and scan tags
type ScanConfiguration struct {
	ScanType string            `json:"type"`
	Values   map[string]string `json:"value"`
}

/*
{"scanId":"bef5d38b-7eb9-4138-b74b-2639fcf49e2e","projectId":"ad34ade3-9bf3-4b5a-91d7-3ad67eca7852","loc":137,"fileCount":12,"isIncremental":false,"isIncrementalCanceled":false,"queryPreset":"ASA Premium"}
*/
type ScanMetadata struct {
	ScanID                string
	ProjectID             string
	LOC                   int
	FileCount             int
	IsIncremental         bool
	IsIncrementalCanceled bool
	PresetName            string `json:"queryPreset"`
}

type ScanResultData struct {
	QueryID      uint64
	QueryName    string
	Group        string
	ResultHash   string
	LanguageName string
	Nodes        []ScanResultNodes
}

type ScanResultNodes struct {
	ID          string
	Line        int
	Name        string
	Column      int
	Length      int
	Method      string
	NodeID      int
	DOMType     string
	FileName    string
	FullName    string
	TypeName    string
	MethodLine  int
	Definitions string
}

type ScanResult struct {
	Type                 string
	ResultID             string `json:"id"`
	SimilarityID         int64  `json:"similarityId,string"`
	Status               string
	State                string
	Severity             string
	CreatedAt            string `json:"created"`
	FirstFoundAt         string
	FoundAt              string
	FirstScanId          string
	Description          string
	Data                 ScanResultData
	VulnerabilityDetails ScanResultDetails
}

type ScanResultDetails struct {
	CweId       int
	Compliances []string
}

// Cx1: StatusDetails - details of each engine type's scan status for a multi-engine scan
type ScanStatusDetails struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Details string `json:"details"`
}

// Very simplified for now
type ScanSummary struct {
	TenantID     string
	ScanID       string
	SASTCounters struct {
		//QueriesCounters           []?
		//SinkFileCounters          []?
		LanguageCounters []struct {
			Language string
			Counter  uint64
		}
		ComplianceCounters []struct {
			Compliance string
			Counter    uint64
		}
		SeverityCounters []struct {
			Severity string
			Counter  uint64
		}
		StatusCounters []struct {
			Status  string
			Counter uint64
		}
		StateCounters []struct {
			State   string
			Counter uint64
		}
		TotalCounter        uint64
		FilesScannedCounter uint64
	}
	// ignoring the other counters
	// KICSCounters
	// SCACounters
	// SCAPackagesCounters
	// SCAContainerCounters
	// APISecCounters
}

// Status - Status Structure
type Status struct {
	ID      int               `json:"id"`
	Name    string            `json:"name"`
	Details ScanStatusDetails `json:"details"`
}

type VersionInfo struct {
	CxOne string `json:"CxOne"`
	KICS  string `json:"KICS"`
	SAST  string `json:"SAST"`
}

type WorkflowLog struct {
	Source    string `json:"Source"`
	Info      string `json:"Info"`
	Timestamp string `json:"Timestamp"`
}

// Cx1 Group/Group - Group Structure
type Group struct {
	GroupID string `json:"id"`
	Name    string `json:"name"`
}

// SystemInstance is the client communicating with the Checkmarx backend
type SystemInstance struct {
	serverURL           string
	iamURL              string // New for Cx1
	tenant              string // New for Cx1
	APIKey              string // New for Cx1
	oauth_client_id     string // separate from APIKey
	oauth_client_secret string //separate from APIKey
	client              piperHttp.Uploader
	logger              *logrus.Entry
}

// System is the interface abstraction of a specific SystemIns
type System interface {
	DownloadReport(reportID string) ([]byte, error)
	GetReportStatus(reportID string) (ReportStatus, error)
	RequestNewReport(scanID, projectID, branch, reportType string) (string, error)

	CreateApplication(appname string) (Application, error)
	GetApplicationByName(appname string) (Application, error)
	GetApplicationByID(appId string) (Application, error)
	UpdateApplication(app *Application) error

	GetScan(scanID string) (Scan, error)
	GetScanMetadata(scanID string) (ScanMetadata, error)
	GetScanResults(scanID string, limit uint64) ([]ScanResult, error)
	GetScanSummary(scanID string) (ScanSummary, error)
	GetResultsPredicates(SimilarityID int64, ProjectID string) ([]ResultsPredicates, error)
	GetScanWorkflow(scanID string) ([]WorkflowLog, error)
	GetLastScans(projectID string, limit int) ([]Scan, error)
	GetLastScansByStatus(projectID string, limit int, status []string) ([]Scan, error)

	ScanProject(projectID, sourceUrl, branch, scanType string, settings []ScanConfiguration) (Scan, error)
	ScanProjectZip(projectID, sourceUrl, branch string, settings []ScanConfiguration) (Scan, error)
	ScanProjectGit(projectID, repoUrl, branch string, settings []ScanConfiguration) (Scan, error)

	UploadProjectSourceCode(projectID string, zipFile string) (string, error)
	CreateProject(projectName string, groupIDs []string) (Project, error)
	CreateProjectInApplication(projectName, applicationID string, groupIDs []string) (Project, error)
	GetPresets() ([]Preset, error)
	GetProjectByID(projectID string) (Project, error)
	GetProjectsByName(projectName string) ([]Project, error)
	GetProjectsByNameAndGroup(projectName, groupID string) ([]Project, error)
	GetProjects() ([]Project, error)
	GetQueries() ([]Query, error)
	//GetShortDescription(scanID int, pathID int) (ShortDescription, error)
	GetGroups() ([]Group, error)
	GetGroupByName(groupName string) (Group, error)
	GetGroupByID(groupID string) (Group, error)
	SetProjectBranch(projectID, branch string, allowOverride bool) error
	SetProjectPreset(projectID, presetName string, allowOverride bool) error
	SetProjectLanguageMode(projectID, languageMode string, allowOverride bool) error
	SetProjectFileFilter(projectID, filter string, allowOverride bool) error

	GetProjectConfiguration(projectID string) ([]ProjectConfigurationSetting, error)
	UpdateProjectConfiguration(projectID string, settings []ProjectConfigurationSetting) error

	GetVersion() (VersionInfo, error)
}

// NewSystemInstance returns a new Checkmarx client for communicating with the backend
// Updated for Cx1
func NewSystemInstance(client piperHttp.Uploader, serverURL, iamURL, tenant, APIKey, client_id, client_secret string) (*SystemInstance, error) {
	loggerInstance := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne")
	sys := &SystemInstance{
		serverURL:           serverURL,
		iamURL:              iamURL,
		tenant:              tenant,
		APIKey:              APIKey,
		oauth_client_id:     client_id,
		oauth_client_secret: client_secret,
		client:              client,
		logger:              loggerInstance,
	}

	var token string
	var err error

	if APIKey != "" {
		token, err = sys.getAPIToken()
		if err != nil {
			return sys, errors.Wrap(err, fmt.Sprintf("Error fetching oAuth token using API Key: %v", shortenGUID(APIKey)))
		}
	} else if client_id != "" && client_secret != "" {
		token, err = sys.getOAuth2Token()
		if err != nil {
			return sys, errors.Wrap(err, fmt.Sprintf("Error fetching oAuth token using OIDC client: %v/%v", shortenGUID(client_id), shortenGUID(client_secret)))
		}
	} else {
		return sys, errors.New("No APIKey or client_id/client_secret provided.")
	}

	log.RegisterSecret(token)

	options := piperHttp.ClientOptions{
		Token:            token,
		TransportTimeout: time.Minute * 15,
	}
	sys.client.SetOptions(options)

	return sys, nil
}

// Updated for Cx1
func sendRequest(sys *SystemInstance, method, url string, body io.Reader, header http.Header, acceptedErrorCodes []int) ([]byte, error) {
	cx1url := fmt.Sprintf("%v/api%v", sys.serverURL, url)
	return sendRequestInternal(sys, method, cx1url, body, header, acceptedErrorCodes)
}

// Updated for Cx1
func sendRequestIAM(sys *SystemInstance, method, base, url string, body io.Reader, header http.Header, acceptedErrorCodes []int) ([]byte, error) {
	iamurl := fmt.Sprintf("%v%v/realms/%v%v", sys.iamURL, base, sys.tenant, url)
	return sendRequestInternal(sys, method, iamurl, body, header, acceptedErrorCodes)
}

// Updated for Cx1
func sendRequestInternal(sys *SystemInstance, method, url string, body io.Reader, header http.Header, acceptedErrorCodes []int) ([]byte, error) {
	var requestBody io.Reader
	var reqBody string
	if body != nil {
		closer := io.NopCloser(body)
		bodyBytes, _ := io.ReadAll(closer)
		reqBody = string(bodyBytes)
		requestBody = bytes.NewBuffer(bodyBytes)
		defer closer.Close()
	}

	if header == nil {
		header = http.Header{}
	}
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:105.0) Gecko/20100101 Firefox/105.0")
	//header.Set("User-Agent", "Project-Piper.io cicd pipeline") // currently this causes some requests to fail due to unknown UA validation in the backend.

	response, err := sys.client.SendRequest(method, url, requestBody, header, nil)
	if err != nil && (response == nil || !piperutils.ContainsInt(acceptedErrorCodes, response.StatusCode)) {

		var resBodyBytes []byte
		if response != nil && response.Body != nil {
			resBodyBytes, _ = io.ReadAll(response.Body)
			defer response.Body.Close()
			resBody := string(resBodyBytes)
			sys.recordRequestDetailsInErrorCase(reqBody, resBody)
		}

		var str string
		if len(resBodyBytes) > 0 {
			var msg map[string]interface{}
			_ = json.Unmarshal(resBodyBytes, &msg)

			if msg["message"] != nil {
				str = msg["message"].(string)
			} else if msg["error_description"] != nil {
				str = msg["error_description"].(string)
			} else if msg["error"] != nil {
				str = msg["error"].(string)
			} else {
				if len(str) > 20 {
					str = string(resBodyBytes[:20])
				} else {
					str = string(resBodyBytes)
				}
			}
		}

		if err != nil {
			if str != "" {
				err = fmt.Errorf("%s: %v", err, str)
			} else {
				err = fmt.Errorf("%s", err)
			}
			sys.logger.Errorf("HTTP request failed with error: %s", err)
			return resBodyBytes, err
		} else {
			if str != "" {
				err = fmt.Errorf("HTTP %v: %v", response.Status, str)
			} else {
				err = fmt.Errorf("HTTP %v", response.Status)
			}

			sys.logger.Errorf("HTTP response indicates error: %s", err)
			return resBodyBytes, err
		}
	}

	data, _ := io.ReadAll(response.Body)
	//sys.logger.Debugf("Valid response body: %v", string(data))
	defer response.Body.Close()
	return data, nil
}

func (sys *SystemInstance) recordRequestDetailsInErrorCase(requestBody string, responseBody string) {
	if requestBody != "" {
		sys.logger.Errorf("Request body: %s", requestBody)
	}
	if responseBody != "" {
		sys.logger.Errorf("Response body: %s", responseBody)
	}
}

/*
   CxOne authentication options are:

       1. APIKey: post to /protocol/openid-connect/token with client_id ("ast-app"), refresh_token (APIKey generated in the UI), & grant_type ("refresh_token")
       2. OAuth Client (service account): post to /protocol/openid-connect/token with client_id (set in the OIDC client in the Cx1 UI), client_secret (set in the UI), grant_type ("client_credentials")

   For regular users, the API Key is likely to be used - it is a token tied to the user account.
   For service accounts, Administrators will need to create OAuth clients.
*/
// Updated for Cx1
func (sys *SystemInstance) getOAuth2Token() (string, error) {
	body := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {sys.oauth_client_id},
		"client_secret": {sys.oauth_client_secret},
	}
	header := http.Header{}
	header.Add("Content-type", "application/x-www-form-urlencoded")
	data, err := sendRequestIAM(sys, http.MethodPost, "/auth", "/protocol/openid-connect/token", strings.NewReader(body.Encode()), header, []int{})
	if err != nil {
		return "", err
	}

	var token AuthToken
	json.Unmarshal(data, &token)
	return token.TokenType + " " + token.AccessToken, nil
}

// Updated for Cx1
func (sys *SystemInstance) getAPIToken() (string, error) {
	body := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {"ast-app"},
		"refresh_token": {sys.APIKey},
	}
	header := http.Header{}
	header.Add("Content-type", "application/x-www-form-urlencoded")
	data, err := sendRequestIAM(sys, http.MethodPost, "/auth", "/protocol/openid-connect/token", strings.NewReader(body.Encode()), header, []int{})
	if err != nil {
		return "", err
	}

	var token AuthToken
	json.Unmarshal(data, &token)
	return token.TokenType + " " + token.AccessToken, nil
}

func (sys *SystemInstance) GetApplicationsByName(name string, limit uint64) ([]Application, error) {
	sys.logger.Debugf("Get Cx1 Applications by name: %v", name)

	var ApplicationResponse struct {
		TotalCount    uint64
		FilteredCount uint64
		Applications  []Application
	}

	body := url.Values{
		//"offset":     {fmt.Sprintf("%d", 0)},
		"limit": {fmt.Sprintf("%d", limit)},
		"name":  {name},
	}

	response, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/applications?%v", body.Encode()), nil, nil, []int{})

	if err != nil {
		return ApplicationResponse.Applications, err
	}

	err = json.Unmarshal(response, &ApplicationResponse)
	sys.logger.Tracef("Retrieved %d applications", len(ApplicationResponse.Applications))
	return ApplicationResponse.Applications, err
}

func (sys *SystemInstance) GetApplicationByID(appId string) (Application, error) {
	sys.logger.Debugf("Get Cx1 Application by ID: %v", appId)

	var ret Application

	response, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/applications/%v", appId), nil, nil, []int{})

	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	return ret, err
}

func (sys *SystemInstance) GetApplicationByName(name string) (Application, error) {
	apps, err := sys.GetApplicationsByName(name, 0)
	if err != nil {
		return Application{}, err
	}

	for _, a := range apps {
		if a.Name == name {
			return a, nil
		}
	}

	return Application{}, fmt.Errorf("no application found named %v", name)
}

func (sys *SystemInstance) CreateApplication(appname string) (Application, error) {
	sys.logger.Debugf("Create Application: %v", appname)
	data := map[string]interface{}{
		"name":        appname,
		"description": "",
		"criticality": 3,
		"rules":       []ApplicationRule{},
		"tags":        map[string]string{},
	}

	var app Application

	jsonBody, err := json.Marshal(data)
	if err != nil {
		return app, err
	}

	response, err := sendRequest(sys, http.MethodPost, "/applications", bytes.NewReader(jsonBody), nil, []int{})
	if err != nil {
		sys.logger.Tracef("Error while creating application: %s", err)
		return app, err
	}

	err = json.Unmarshal(response, &app)

	return app, err
}

func (a *Application) GetRuleByType(ruletype string) *ApplicationRule {
	for id := range a.Rules {
		if a.Rules[id].Type == ruletype {
			return &(a.Rules[id])
		}
	}
	return nil
}

func (a *Application) AddRule(ruletype, value string) {
	rule := a.GetRuleByType(ruletype)
	if rule == nil {
		var newrule ApplicationRule
		newrule.Type = ruletype
		newrule.Value = value
		a.Rules = append(a.Rules, newrule)
	} else {
		if rule.Value == value || strings.Contains(rule.Value, fmt.Sprintf(";%v;", value)) || rule.Value[len(rule.Value)-len(value)-1:] == fmt.Sprintf(";%v", value) || rule.Value[:len(value)+1] == fmt.Sprintf("%v;", value) {
			return // rule value already contains this value
		}
		rule.Value = fmt.Sprintf("%v;%v", rule.Value, value)
	}
}

func (a *Application) AssignProject(project *Project) {
	a.AddRule("project.name.in", project.Name)
}

func (sys *SystemInstance) UpdateApplication(app *Application) error {
	sys.logger.Debugf("Updating application: %v", app.Name)
	jsonBody, err := json.Marshal(*app)
	if err != nil {
		return err
	}

	_, err = sendRequest(sys, http.MethodPut, fmt.Sprintf("/applications/%v", app.ApplicationID), bytes.NewReader(jsonBody), nil, []int{})
	if err != nil {
		sys.logger.Tracef("Error while updating application: %s", err)
		return err
	}

	return nil
}

// Updated for Cx1
func (sys *SystemInstance) GetGroups() ([]Group, error) {
	sys.logger.Debug("Getting Groups...")
	var groups []Group

	data, err := sendRequestIAM(sys, http.MethodGet, "/auth", "/pip/groups", nil, http.Header{}, []int{})
	if err != nil {
		sys.logger.Errorf("Fetching groups failed: %s", err)
		return groups, err
	}

	err = json.Unmarshal(data, &groups)
	if err != nil {
		sys.logger.Errorf("Fetching groups failed: %s", err)
		return groups, err
	}

	return groups, nil
}

// New for Cx1
func (sys *SystemInstance) GetGroupByName(groupName string) (Group, error) {
	sys.logger.Debugf("Getting Group named %v...", groupName)
	groups, err := sys.GetGroups()
	var group Group
	if err != nil {
		return group, err
	}

	for _, g := range groups {
		if g.Name == groupName {
			return g, nil
		}
	}

	return group, errors.New(fmt.Sprintf("No group matching %v", groupName))
}

// New for Cx1
func (sys *SystemInstance) GetGroupByID(groupID string) (Group, error) {
	sys.logger.Debugf("Getting Group with ID %v...", groupID)
	groups, err := sys.GetGroups()
	var group Group
	if err != nil {
		return group, err
	}

	for _, g := range groups {
		if g.GroupID == groupID {
			return g, nil
		}
	}

	return group, errors.New(fmt.Sprintf("No group with ID %v", groupID))
}

// GetProjects returns the projects defined in the Checkmarx backend which the user has access to
func (sys *SystemInstance) GetProjects() ([]Project, error) {
	return sys.GetProjectsByNameAndGroup("", "")
}

// GetProjectByID returns the project addressed by projectID from the Checkmarx backend which the user has access to
// Updated for Cx1
func (sys *SystemInstance) GetProjectByID(projectID string) (Project, error) {
	sys.logger.Debugf("Getting Project with ID %v...", projectID)
	var project Project

	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/projects/%v", projectID), nil, http.Header{}, []int{})
	if err != nil {
		return project, errors.Wrapf(err, "fetching project %v failed", projectID)
	}

	err = json.Unmarshal(data, &project)
	return project, err
}

// GetProjectsByNameAndGroup returns the project addressed by project name from the Checkmarx backend which the user has access to
// Updated for Cx1
func (sys *SystemInstance) GetProjectsByName(projectName string) ([]Project, error) {
	sys.logger.Debugf("Getting projects with name %v", projectName)

	var projectResponse struct {
		TotalCount    int       `json:"totalCount"`
		FilteredCount int       `json:"filteredCount"`
		Projects      []Project `json:"projects"`
	}

	header := http.Header{}
	header.Set("Accept-Type", "application/json")
	var data []byte
	var err error

	body := url.Values{}
	body.Add("name", projectName)

	data, err = sendRequest(sys, http.MethodGet, fmt.Sprintf("/projects/?%v", body.Encode()), nil, header, []int{404})

	if err != nil {
		return []Project{}, errors.Wrapf(err, "fetching project %v failed", projectName)
	}

	err = json.Unmarshal(data, &projectResponse)
	return projectResponse.Projects, err
}

// GetProjectsByNameAndGroup returns the project addressed by project name from the Checkmarx backend which the user has access to
// Updated for Cx1
func (sys *SystemInstance) GetProjectsByNameAndGroup(projectName, groupID string) ([]Project, error) {
	sys.logger.Debugf("Getting projects with name %v of group %v...", projectName, groupID)

	var projectResponse struct {
		TotalCount    int       `json:"totalCount"`
		FilteredCount int       `json:"filteredCount"`
		Projects      []Project `json:"projects"`
	}

	header := http.Header{}
	header.Set("Accept-Type", "application/json")
	var data []byte
	var err error

	body := url.Values{}
	if len(groupID) > 0 {
		body.Add("groups", groupID)
	}
	if len(projectName) > 0 {
		body.Add("name", projectName)
	}

	if len(body) > 0 {
		data, err = sendRequest(sys, http.MethodGet, fmt.Sprintf("/projects/?%v", body.Encode()), nil, header, []int{404})
	} else {
		data, err = sendRequest(sys, http.MethodGet, "/projects/", nil, header, []int{404})
	}
	if err != nil {
		return projectResponse.Projects, errors.Wrapf(err, "fetching project %v failed", projectName)
	}

	err = json.Unmarshal(data, &projectResponse)
	return projectResponse.Projects, err
}

// CreateProject creates a new project in the Checkmarx backend
// Updated for Cx1
func (sys *SystemInstance) CreateProject(projectName string, groupIDs []string) (Project, error) {
	var project Project
	jsonData := map[string]interface{}{
		"name":        projectName,
		"groups":      groupIDs,
		"origin":      cxOrigin,
		"criticality": 3, // default
		// multiple additional parameters exist as options
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return project, errors.Wrapf(err, "failed to marshal project data")
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")

	data, err := sendRequest(sys, http.MethodPost, "/projects", bytes.NewBuffer(jsonValue), header, []int{})
	if err != nil {
		return project, errors.Wrapf(err, "failed to create project %v", projectName)
	}

	err = json.Unmarshal(data, &project)
	return project, err
}

func (sys *SystemInstance) CreateProjectInApplication(projectName, applicationID string, groupIDs []string) (Project, error) {
	var project Project
	jsonData := map[string]interface{}{
		"name":           projectName,
		"groups":         groupIDs,
		"origin":         cxOrigin,
		"criticality":    3, // default
		"applicationIds": []string{applicationID},
		// multiple additional parameters exist as options
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return project, errors.Wrapf(err, "failed to marshal project data")
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	data, err := sendRequest(sys, http.MethodPost, "/projects", bytes.NewReader(jsonValue), header, []int{})

	if err != nil {
		return project, errors.Wrapf(err, "failed to create project %v under %v", projectName, applicationID)
	}

	err = json.Unmarshal(data, &project)
	if err != nil {
		return project, errors.Wrapf(err, "failed to unmarshal project data")
	}

	// since there is a delay to assign a project to an application, adding a check to ensure project is ready after creation
	// (if project is not ready, 403 will be returned)
	projectID := project.ProjectID
	project, err = sys.GetProjectByID(projectID)
	if err != nil {
		const max_retry = 12 // 3 minutes
		const delay = 15
		retry_counter := 1
		for retry_counter <= max_retry && err != nil {
			sys.logger.Debug("Waiting for project assignment to application, retry #", retry_counter)
			time.Sleep(delay * time.Second)
			retry_counter++
			project, err = sys.GetProjectByID(projectID)
		}
	}

	return project, err
}

// New for Cx1
func (sys *SystemInstance) GetUploadURI() (string, error) {
	sys.logger.Debug("Retrieving upload URI")
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	resp, err := sendRequest(sys, http.MethodPost, "/uploads", nil, header, []int{})

	if err != nil {
		return "", errors.Wrap(err, "failed to get an upload uri")
	}

	responseData := make(map[string]string)
	json.Unmarshal(resp, &responseData)
	sys.logger.Debugf("Upload URI %s", responseData["url"])

	return responseData["url"], nil
}

func (sys *SystemInstance) UploadProjectSourceCode(projectID string, zipFile string) (string, error) {
	sys.logger.Debugf("Preparing to upload file %v...", zipFile)

	// get URI
	uploadUri, err := sys.GetUploadURI()
	if err != nil {
		return "", err
	}

	header := http.Header{}
	header.Add("Accept-Encoding", "gzip,deflate")
	header.Add("Content-Type", "application/zip")
	header.Add("Accept", "application/json")

	zipContents, err := os.ReadFile(zipFile)
	if err != nil {
		sys.logger.Error("Failed to Read the File " + zipFile + ": " + err.Error())
		return "", err
	}

	response, err := sendRequestInternal(sys, http.MethodPut, uploadUri, bytes.NewReader(zipContents), header, []int{})
	if err != nil {
		sys.logger.Errorf("Failed to upload file %v: %s", zipFile, err)
		return uploadUri, err
	}

	sys.logger.Debugf("Upload request response: %v", string(response))

	return uploadUri, nil
}

func (sys *SystemInstance) scanProject(scanConfig map[string]interface{}) (Scan, error) {
	scan := Scan{}

	jsonValue, err := json.Marshal(scanConfig)
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	sys.logger.Tracef("Starting scan with settings: " + string(jsonValue))

	data, err := sendRequest(sys, http.MethodPost, "/scans/", bytes.NewBuffer(jsonValue), header, []int{})
	if err != nil {
		return scan, err
	}

	err = json.Unmarshal(data, &scan)
	return scan, err
}

func (sys *SystemInstance) ScanProjectZip(projectID, sourceUrl, branch string, settings []ScanConfiguration) (Scan, error) {
	jsonBody := map[string]interface{}{
		"project": map[string]interface{}{"id": projectID},
		"type":    "upload",
		"handler": map[string]interface{}{
			"uploadurl": sourceUrl,
			"branch":    branch,
		},
		"config": settings,
	}

	scan, err := sys.scanProject(jsonBody)
	if err != nil {
		return scan, errors.Wrapf(err, "Failed to start a zip scan for project %v", projectID)
	}
	return scan, err
}

func (sys *SystemInstance) ScanProjectGit(projectID, repoUrl, branch string, settings []ScanConfiguration) (Scan, error) {
	jsonBody := map[string]interface{}{
		"project": map[string]interface{}{"id": projectID},
		"type":    "git",
		"handler": map[string]interface{}{
			"repoUrl": repoUrl,
			"branch":  branch,
		},
		"config": settings,
	}

	scan, err := sys.scanProject(jsonBody)
	if err != nil {
		return scan, errors.Wrapf(err, "Failed to start a git scan for project %v", projectID)
	}
	return scan, err
}

func (sys *SystemInstance) ScanProject(projectID, sourceUrl, branch, scanType string, settings []ScanConfiguration) (Scan, error) {
	if scanType == "upload" {
		return sys.ScanProjectZip(projectID, sourceUrl, branch, settings)
	} else if scanType == "git" {
		return sys.ScanProjectGit(projectID, sourceUrl, branch, settings)
	}

	return Scan{}, errors.New("Invalid scanType provided, must be 'upload' or 'git'")
}

func (sys *SystemInstance) GetPresets() ([]Preset, error) {
	sys.logger.Debug("Getting Presets...")
	var presets []Preset

	data, err := sendRequest(sys, http.MethodGet, "/queries/presets", nil, http.Header{}, []int{})
	if err != nil {
		sys.logger.Errorf("Fetching presets failed: %s", err)
		return presets, err
	}

	err = json.Unmarshal(data, &presets)
	return presets, err
}

func (sys *SystemInstance) GetProjectConfiguration(projectID string) ([]ProjectConfigurationSetting, error) {
	sys.logger.Debug("Getting project configuration")
	var projectConfigurations []ProjectConfigurationSetting
	params := url.Values{
		"project-id": {projectID},
	}
	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/configuration/project?%v", params.Encode()), nil, http.Header{}, []int{})

	if err != nil {
		sys.logger.Errorf("Failed to get project configuration for project ID %v: %s", projectID, err)
		return projectConfigurations, err
	}

	err = json.Unmarshal(data, &projectConfigurations)
	return projectConfigurations, err
}

func (sys *SystemInstance) UpdateProjectConfiguration(projectID string, settings []ProjectConfigurationSetting) error {
	if len(settings) == 0 {
		return errors.New("Empty list of settings provided.")
	}

	params := url.Values{
		"project-id": {projectID},
	}

	jsonValue, err := json.Marshal(settings)

	if err != nil {
		sys.logger.Errorf("Failed to marshal settings.")
		return err
	}

	_, err = sendRequest(sys, http.MethodPatch, fmt.Sprintf("/configuration/project?%v", params.Encode()), bytes.NewReader(jsonValue), http.Header{}, []int{})
	if err != nil {
		sys.logger.Errorf("Failed to update project configuration: %s", err)
		return err
	}

	return nil
}

func (sys *SystemInstance) SetProjectBranch(projectID, branch string, allowOverride bool) error {
	var setting ProjectConfigurationSetting
	setting.Key = "scan.handler.git.branch"
	setting.Value = branch
	setting.AllowOverride = allowOverride

	return sys.UpdateProjectConfiguration(projectID, []ProjectConfigurationSetting{setting})
}

func (sys *SystemInstance) SetProjectPreset(projectID, presetName string, allowOverride bool) error {
	var setting ProjectConfigurationSetting
	setting.Key = "scan.config.sast.presetName"
	setting.Value = presetName
	setting.AllowOverride = allowOverride

	return sys.UpdateProjectConfiguration(projectID, []ProjectConfigurationSetting{setting})
}

func (sys *SystemInstance) SetProjectLanguageMode(projectID, languageMode string, allowOverride bool) error {
	var setting ProjectConfigurationSetting
	setting.Key = "scan.config.sast.languageMode"
	setting.Value = languageMode
	setting.AllowOverride = allowOverride

	return sys.UpdateProjectConfiguration(projectID, []ProjectConfigurationSetting{setting})
}

func (sys *SystemInstance) SetProjectFileFilter(projectID, filter string, allowOverride bool) error {
	var setting ProjectConfigurationSetting
	setting.Key = "scan.config.sast.filter"
	setting.Value = filter
	setting.AllowOverride = allowOverride

	// TODO - apply the filter across all languages? set up separate calls per engine? engine as param?

	return sys.UpdateProjectConfiguration(projectID, []ProjectConfigurationSetting{setting})
}

// GetScans returns all scan status on the project addressed by projectID
func (sys *SystemInstance) GetScan(scanID string) (Scan, error) {
	var scan Scan

	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/scans/%v", scanID), nil, http.Header{}, []int{})
	if err != nil {
		sys.logger.Errorf("Failed to fetch scan with ID %v: %s", scanID, err)
		return scan, errors.Wrapf(err, "failed to fetch scan with ID %v", scanID)
	}

	json.Unmarshal(data, &scan)
	return scan, nil
}

func (sys *SystemInstance) GetScanMetadata(scanID string) (ScanMetadata, error) {
	var scanmeta ScanMetadata

	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/sast-metadata/%v", scanID), nil, http.Header{}, []int{})
	if err != nil {
		sys.logger.Errorf("Failed to fetch metadata for scan with ID %v: %s", scanID, err)
		return scanmeta, errors.Wrapf(err, "failed to fetch metadata for scan with ID %v", scanID)
	}

	json.Unmarshal(data, &scanmeta)
	return scanmeta, nil
}

func (sys *SystemInstance) GetScanWorkflow(scanID string) ([]WorkflowLog, error) {
	var workflow []WorkflowLog

	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/scans/%v/workflow", scanID), nil, http.Header{}, []int{})
	if err != nil {
		sys.logger.Errorf("Failed to fetch scan with ID %v: %s", scanID, err)
		return []WorkflowLog{}, errors.Wrapf(err, "failed to fetch scan with ID %v", scanID)
	}

	json.Unmarshal(data, &workflow)
	return workflow, nil
}

func (sys *SystemInstance) GetLastScans(projectID string, limit int) ([]Scan, error) {
	var scanResponse struct {
		TotalCount         uint64
		FilteredTotalCount uint64
		Scans              []Scan
	}

	body := url.Values{
		"project-id": {projectID},
		"offset":     {fmt.Sprintf("%v", 0)},
		"limit":      {fmt.Sprintf("%v", limit)},
		"sort":       {"+created_at"},
	}

	header := http.Header{}
	header.Set("Accept-Type", "application/json")
	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/scans/?%v", body.Encode()), nil, header, []int{})
	if err != nil {
		sys.logger.Errorf("Failed to fetch scans of project %v: %s", projectID, err)
		return []Scan{}, errors.Wrapf(err, "failed to fetch scans of project %v", projectID)
	}

	err = json.Unmarshal(data, &scanResponse)
	return scanResponse.Scans, err
}

func (sys *SystemInstance) GetLastScansByStatus(projectID string, limit int, status []string) ([]Scan, error) {
	var scanResponse struct {
		TotalCount         uint64
		FilteredTotalCount uint64
		Scans              []Scan
	}

	body := url.Values{
		"project-id": {projectID},
		"offset":     {fmt.Sprintf("%d", 0)},
		"limit":      {fmt.Sprintf("%d", limit)},
		"sort":       {"+created_at"},
		"statuses":   status,
	}

	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/scans/?%v", body.Encode()), nil, nil, []int{})
	if err != nil {
		sys.logger.Errorf("Failed to fetch scans of project %v: %s", projectID, err)
		return []Scan{}, errors.Wrapf(err, "failed to fetch scans of project %v", projectID)
	}

	err = json.Unmarshal(data, &scanResponse)

	return scanResponse.Scans, err
}

func (s *Scan) IsIncremental() (bool, error) {
	for _, scanconfig := range s.Metadata.Configs {
		if scanconfig.ScanType == "sast" {
			if val, ok := scanconfig.Values["incremental"]; ok {
				return val == "true", nil
			}
		}
	}
	return false, errors.New(fmt.Sprintf("Scan %v did not have a sast-engine incremental flag set", s.ScanID))
}

func (sys *SystemInstance) GetScanResults(scanID string, limit uint64) ([]ScanResult, error) {
	sys.logger.Debug("Get Cx1 Scan Results")
	var resultResponse struct {
		Results    []ScanResult
		TotalCount int
	}

	params := url.Values{
		"scan-id":  {scanID},
		"limit":    {fmt.Sprintf("%d", limit)},
		"state":    []string{},
		"severity": []string{},
		"status":   []string{},
	}

	response, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/results/?%v", params.Encode()), nil, nil, []int{})
	if err != nil && len(response) == 0 {
		sys.logger.Errorf("Failed to retrieve scan results for scan ID %v", scanID)
		return []ScanResult{}, err
	}

	err = json.Unmarshal(response, &resultResponse)
	if err != nil {
		sys.logger.Errorf("Failed while parsing response: %s", err)
		sys.logger.Tracef("Response contents: %s", string(response))
		return []ScanResult{}, err
	}
	sys.logger.Debugf("Retrieved %d results", resultResponse.TotalCount)

	if len(resultResponse.Results) != resultResponse.TotalCount {
		sys.logger.Warnf("Expected results total count %d but parsed only %d", resultResponse.TotalCount, len(resultResponse.Results))
		sys.logger.Warnf("Response was: %v", string(response))
	}

	return resultResponse.Results, nil
}

func (s *ScanSummary) TotalCount() uint64 {
	var count uint64
	count = 0

	for _, c := range s.SASTCounters.StateCounters {
		count += c.Counter
	}

	return count
}

func (sys *SystemInstance) GetScanSummary(scanID string) (ScanSummary, error) {
	var ScansSummaries struct {
		ScanSum    []ScanSummary `json:"scansSummaries"`
		TotalCount uint64
	}

	params := url.Values{
		"scan-ids":                {scanID},
		"include-queries":         {"false"},
		"include-status-counters": {"true"},
		"include-files":           {"false"},
	}

	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/scan-summary/?%v", params.Encode()), nil, nil, []int{})
	if err != nil {
		sys.logger.Errorf("Failed to fetch metadata for scan with ID %v: %s", scanID, err)
		return ScanSummary{}, errors.Wrapf(err, "failed to fetch metadata for scan with ID %v", scanID)
	}

	err = json.Unmarshal(data, &ScansSummaries)

	if err != nil {
		return ScanSummary{}, err
	}
	if ScansSummaries.TotalCount == 0 {
		return ScanSummary{}, errors.New(fmt.Sprintf("Failed to retrieve scan summary for scan ID %v", scanID))
	}

	if len(ScansSummaries.ScanSum) == 0 {
		sys.logger.Errorf("Failed to parse data, 0-len ScanSum.\n%v", string(data))
		return ScanSummary{}, errors.New("Fail")
	}

	return ScansSummaries.ScanSum[0], nil
}

func (sys *SystemInstance) GetResultsPredicates(SimilarityID int64, ProjectID string) ([]ResultsPredicates, error) {
	sys.logger.Debugf("Fetching results predicates for project %v similarityId %d", ProjectID, SimilarityID)

	var Predicates struct {
		PredicateHistoryPerProject []struct {
			ProjectID    string
			SimilarityID int64 `json:"similarityId,string"`
			Predicates   []ResultsPredicates
			TotalCount   uint
		}

		TotalCount uint
	}
	response, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/sast-results-predicates/%d?project-ids=%v", SimilarityID, ProjectID), nil, nil, []int{})
	if err != nil {
		return []ResultsPredicates{}, err
	}

	err = json.Unmarshal(response, &Predicates)
	if err != nil {
		return []ResultsPredicates{}, err
	}

	if Predicates.TotalCount == 0 {
		return []ResultsPredicates{}, nil
	}

	return Predicates.PredicateHistoryPerProject[0].Predicates, err
}

// RequestNewReport triggers the generation of a  report for a specific scan addressed by scanID
func (sys *SystemInstance) RequestNewReport(scanID, projectID, branch, reportType string) (string, error) {
	jsonData := map[string]interface{}{
		"fileFormat": reportType,
		"reportType": "ui",
		"reportName": "scan-report",
		"data": map[string]interface{}{
			"scanId":     scanID,
			"projectId":  projectID,
			"branchName": branch,
			"sections": []string{
				"ScanSummary",
				"ExecutiveSummary",
				"ScanResults",
			},
			"scanners": []string{"SAST"},
			"host":     "",
		},
	}

	jsonValue, _ := json.Marshal(jsonData)

	header := http.Header{}
	header.Set("cxOrigin", cxOrigin)
	header.Set("Content-Type", "application/json")
	data, err := sendRequest(sys, http.MethodPost, "/reports", bytes.NewBuffer(jsonValue), header, []int{})
	if err != nil {
		return "", errors.Wrapf(err, "Failed to trigger report generation for scan %v", scanID)
	} else {
		sys.logger.Infof("Generating report %v", string(data))
	}

	var reportResponse struct {
		ReportId string
	}
	err = json.Unmarshal(data, &reportResponse)

	return reportResponse.ReportId, err
}

// GetReportStatus returns the status of the report generation process
func (sys *SystemInstance) GetReportStatus(reportID string) (ReportStatus, error) {
	var response ReportStatus

	header := http.Header{}
	header.Set("Accept", "application/json")
	data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/reports/%v", reportID), nil, header, []int{})
	if err != nil {
		sys.logger.Errorf("Failed to fetch report status for reportID %v: %s", reportID, err)
		return response, errors.Wrapf(err, "failed to fetch report status for reportID %v", reportID)
	}

	json.Unmarshal(data, &response)
	return response, nil
}

func (sys *SystemInstance) GetQueries() ([]Query, error) {
	sys.logger.Debug("Get Cx1 Queries")
	var queries []Query

	response, err := sendRequest(sys, http.MethodGet, "/presets/queries", nil, nil, []int{})
	if err != nil {
		return queries, err
	}

	err = json.Unmarshal(response, &queries)
	if err != nil {
		sys.logger.Errorf("Failed to parse %v", string(response))
	}
	return queries, err
}

func shortenGUID(guid string) string {
	return fmt.Sprintf("%v..%v", guid[:2], guid[len(guid)-2:])
}

// DownloadReport downloads the report addressed by reportID and returns the XML contents
func (sys *SystemInstance) DownloadReport(reportUrl string) ([]byte, error) {
	header := http.Header{}
	header.Set("Accept", "application/json")
	data, err := sendRequestInternal(sys, http.MethodGet, reportUrl, nil, header, []int{})
	if err != nil {
		return []byte{}, errors.Wrapf(err, "failed to download report from url: %v", reportUrl)
	}
	return data, nil
}

func (sys *SystemInstance) GetVersion() (VersionInfo, error) {
	sys.logger.Debug("Getting Version information...")
	var version VersionInfo

	data, err := sendRequest(sys, http.MethodGet, "/versions", nil, http.Header{}, []int{})
	if err != nil {
		sys.logger.Errorf("Fetching versions failed: %s", err)
		return version, err
	}

	err = json.Unmarshal(data, &version)
	return version, err
}
