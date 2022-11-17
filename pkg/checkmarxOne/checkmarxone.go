package checkmarxone

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    //"strconv"
    "strings"
    "time"
    "encoding/xml"
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
    AccessToken             string `json:"access_token"`    
    ExpiresIn               int    `json:"expires_in"`
    RefreshExpiresIn        int    `json:"refresh_expires_in"`
    TokenType               string `json:"token_type"`
    NotBeforePolicy         int    `json:"not-before-policy"`
    Scope                   string `json:"scope"`
}

// Preset - Project's Preset
// Updated for Cx1
type Preset struct {
    PresetID        int    `json:"id"`
    Name      string `json:"name"`
}

// Project - Project Structure
// Updated for Cx1
type Project struct {
    ProjectID           string              `json:"id"`
    Name                string              `json:"name"`
    CreatedAt           string              `json:"createdAt"`
    UpdatedAt           string              `json:"updatedAt"`
    Groups              []string            `json:"groups"`
    Tags                map[string]string   `json:"tags"`
    RepoUrl             string              `json:"repoUrl"`
    MainBranch          string              `json:"mainBranch"`
    Origin              string              `json:"origin"`
    Criticality         int                 `json:"criticality"`
}

// New for Cx1
// These settings are higher-level settings that define how an engine should run, for example "multi-language" mode or setting a preset.
type ProjectConfigurationSetting struct {
    Key                 string              `json:"key"`
    Name                string              `json:"name"`
    Category            string              `json:"category"`
    OriginLevel         string              `json:"originLevel"`
    Value               string              `json:"value"`
    ValueType           string              `json:"valuetype"`
    ValueTypeParams     string              `json:"valuetypeparams"`
    AllowOverride       bool                `json:"allowOverride"`
}


// ReportStatus - ReportStatus Structure
// Updated for Cx1
type ReportStatus struct {
    ReportID            string              `json:"reportId"`
    Status              string              `json:"status"`
    ReportURL           string              `json:"url"`
}


// Scan - Scan Structure
// updated for Cx1
type Scan struct {
    ScanID   string  `json:"id"`
    Status string `json:"status"`
    StatusDetails []ScanStatusDetails  `json:"statusDetails"`
    Branch string `json:"branch"`
    CreatedAt string `json:"createdAt"`
    UpdatedAt string `json:"updatedAt"`
    ProjectID string `json:"projectId"`
    ProjectName string `json:"projectName"`
    UserAgent string `json:"userAgent"`
    Initiator string `json:"initiator"`
    Tags map[string]string `json:"tags"`
    Metadata struct {
        Type string `json:"type"`
        Configs []ScanConfiguration `json:"configs"`
    } `json:"metadata"`
    Engines []string `json:"engines"`
    SourceType string `json:"sourceType"`
    SourceOrigin string `json:"sourceOrigin"`
}


// New for Cx1: ScanConfiguration - list of key:value pairs used to configure the scan for each scan engine
// This is specifically for scan-level configurations like "is incremental" and scan tags
type ScanConfiguration struct {
    ScanType string `json:"type"`
    Values map[string]string `json:"value"`
}

/*
// ScanStatus - ScanStatus Structure
type ScanStatus struct {
    ID              int    `json:"id"`
    Link            Link   `json:"link"`
    Status          Status `json:"status"`
    ScanType        string `json:"scanType"`
    Comment         string `json:"comment"`
    IsIncremental   bool   `json:"isIncremental"`
}
*/

// Cx1: StatusDetails - details of each engine type's scan status for a multi-engine scan
type ScanStatusDetails struct {
    Name            string `json:"name"`
    Status          string `json:"status"`
    Details         string `json:"details"`
}

// Status - Status Structure
type Status struct {
    ID      int                 `json:"id"`
    Name    string              `json:"name"`
    Details ScanStatusDetails   `json:"details"`
}


// Cx1 Group/Group - Group Structure
type Group struct {
    GroupID          string              `json:"id"`
    Name        string              `json:"name"`
}


// Query - Query Structure
type Query struct {
    XMLName xml.Name `xml:"Query"`
    Name    string   `xml:"name,attr"`
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
    iamURL       string // New for Cx1
    tenant       string // New for Cx1
    APIKey      string // New for Cx1
    oauth_client_id string // separate from APIKey
    oauth_client_secret string //separate from APIKey
    client    piperHttp.Uploader
    logger    *logrus.Entry
}

// System is the interface abstraction of a specific SystemIns
type System interface {
    //FilterPresetByName(presets []Preset, presetName string) Preset
    //FilterPresetByID(presets []Preset, presetID int) Preset
    //FilterProjectByName(projects []Project, projectName string) Project
    //FilterGroupByName(groups []Group, groupName string) (Group, error)
    //FilterGroupByID(groups []Group, groupID json.RawMessage) Group
    //DownloadReport(reportID int) ([]byte, error)
    //GetReportStatus(reportID int) (ReportStatusResponse, error)
    //RequestNewReport(scanID int, reportType string) (Report, error)
    //GetResults(scanID int) ResultsStatistics
    //GetScanStatusAndDetail(scanID int) (string, ScanStatusDetail)
    GetScans(projectID string) ([]Scan, error)
    //ScanProject(projectID string, isIncremental, isPublic, forceScan bool) (Scan, error)
    //UpdateProjectConfiguration(projectID string, presetID int, engineConfigurationID string) error
    //UpdateProjectExcludeSettings(projectID string, excludeFolders string, excludeFiles string) error
    UploadAndScanProjectSourceCode(projectID string, zipFile string) (Scan, error)
    UploadProjectSourceCode(projectID string, zipFile string) (string, error)
    CreateProject(projectName string, groupIDs []string) (Project, error)
    //CreateBranch(projectID string, branchName string) int
    GetPresets() ([]Preset, error)
    GetProjectByID(projectID string) (Project, error)
    GetProjectsByNameAndGroup(projectName, groupID string) ([]Project, error)
    GetProjects() ([]Project, error)
    //GetShortDescription(scanID int, pathID int) (ShortDescription, error)
    GetGroups() ([]Group, error)
    GetGroupByName( groupName string ) (Group, error)
}

// NewSystemInstance returns a new Checkmarx client for communicating with the backend
// Updated for Cx1
func NewSystemInstance(client piperHttp.Uploader, serverURL, iamURL, tenant, APIKey, client_id, client_secret string) (*SystemInstance, error) {
    loggerInstance := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarxOne")
    sys := &SystemInstance{
        serverURL: serverURL,
        iamURL: iamURL,
        tenant: tenant,
        APIKey: APIKey,
        oauth_client_id: client_id,
        oauth_client_secret: client_secret,
        client:    client,
        logger:    loggerInstance,
    }

    var token string
    var err error
    
    if APIKey != "" {
        token, err = sys.getAPIToken()
        if err != nil {
            return sys, errors.Wrap(err, "Error fetching oAuth token using API Key")
        }
    } else {
        token, err = sys.getOAuth2Token()
        if err != nil {
            return sys, errors.Wrap(err, "Error fetching oAuth token using OIDC client")
        }
    }


    log.RegisterSecret(token)

    options := piperHttp.ClientOptions{
        Token:            token,
        TransportTimeout: time.Minute * 15,
    }
    sys.client.SetOptions(options)

    return sys, nil
}


/*
    Different API calls:

    {{Cx1_URL}}/api/projects
    {{Cx1_URL}}/api/applications
    {{Cx1_URL}}/api/presets?limit=100
*/
// Updated for Cx1
func sendRequest(sys *SystemInstance, method, url string, body io.Reader, header http.Header, acceptedErrorCodes []int ) ([]byte, error) {
    cx1url := fmt.Sprintf("%v/api%v", sys.serverURL, url)
    return sendRequestInternal(sys, method, cx1url, body, header, acceptedErrorCodes )
}

/*
    Different IAM calls:

    {{Cx1_IAM}}/auth/admin/realms/{{Cx1_Tenant}}/users?first=0&max=20&briefRepresentation=true
    {{Cx1_IAM}}/auth/realms/{{Cx1_Tenant}}/users

    {{Cx1_IAM}}/auth/admin/realms/{{Cx1_Tenant}}/groups?briefRepresentation=true
    {{Cx1_IAM}}/auth/realms/{{Cx1_Tenant}}/pip/groups

    Note: some have /auth/admin, others just /auth
*/
// Updated for Cx1
func sendRequestIAM(sys *SystemInstance, method, base, url string, body io.Reader, header http.Header, acceptedErrorCodes []int ) ([]byte, error) {
    iamurl := fmt.Sprintf("%v%v/realms/%v/api%v", sys.iamURL, base, sys.tenant, url)
    return sendRequestInternal(sys, method, iamurl, body, header, acceptedErrorCodes)
}

// Updated for Cx1
func sendRequestInternal(sys *SystemInstance, method, url string, body io.Reader, header http.Header, acceptedErrorCodes []int) ([]byte, error) {
    var requestBody io.Reader
    var requestBodyCopy io.Reader
    if body != nil {
        closer := ioutil.NopCloser(body)
        bodyBytes, _ := ioutil.ReadAll(closer)
        requestBody = bytes.NewBuffer(bodyBytes)
        requestBodyCopy = bytes.NewBuffer(bodyBytes)
        defer closer.Close()
    }
    response, err := sys.client.SendRequest(method, url, requestBody, header, nil)
    if err != nil && (response == nil || !piperutils.ContainsInt(acceptedErrorCodes, response.StatusCode)) {
        sys.recordRequestDetailsInErrorCase(requestBodyCopy, response)
        sys.logger.Errorf("HTTP request failed with error: %s", err)
        return nil, err
    }

    data, _ := ioutil.ReadAll(response.Body)
    sys.logger.Debugf("Valid response body: %v", string(data))
    defer response.Body.Close()
    return data, nil
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
    data, err := sendRequest(sys, http.MethodPost, "/auth/identity/connect/token", strings.NewReader(body.Encode()), header, []int{})
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
        "client_secret": {sys.APIKey},
    }
    header := http.Header{}
    header.Add("Content-type", "application/x-www-form-urlencoded")
    data, err := sendRequest(sys, http.MethodPost, "/auth/identity/connect/token", strings.NewReader(body.Encode()), header, []int{})
    if err != nil {
        return "", err
    }

    var token AuthToken
    json.Unmarshal(data, &token)
    return token.TokenType + " " + token.AccessToken, nil
}

// GetGroups returns the groups the user is assigned to
// TODO: This functionality doesn't seem to work correctly in Cx1 yet, returns all groups
/*
    {{Cx1_IAM}}/auth/realms/{{Cx1_Tenant}}/pip/groups  - regular users/oidc clients/api keys
    {{Cx1_IAM}}/auth/admin/realms/{{Cx1_Tenant}}/groups - regular users/oidc clients/api keys?
    {{Cx1_IAM}}/auth/admin/realms/{{Cx1_Tenant}}/groups/:groupId/members - admin API Key only. This actually returns the members of a group
    {{Cx1_IAM}}/auth/admin/realms/{{Cx1_Tenant}}/users/:userid/groups - admin API key only, groups assigned to a user
*/
// Updated for Cx1
func (sys *SystemInstance) GetGroups() ([]Group, error) {
    sys.logger.Debug("Getting Groups...")
    var groups []Group


    data, err := sendRequestIAM(sys, http.MethodGet, "/auth", "/pip/groups", nil, nil, []int{})
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
func (sys *SystemInstance) GetGroupByName( groupName string ) (Group, error) {
    sys.logger.Debug("Getting Group named %v...", groupName)
    var groups []Group

    body := url.Values {
        "briefRepresentation" : {"true"},
        "search": {groupName},
    }

    data, err := sendRequestIAM(sys, http.MethodGet, "/auth/admin", "/groups?%v", strings.NewReader(body.Encode()), nil, nil)
    if err != nil {
        sys.logger.Errorf("Fetching groups failed: %s", err)
        return Group{}, err
    }

    err = json.Unmarshal(data, &groups)
    if err != nil {
        sys.logger.Errorf("Fetching groups failed: %s", err)
        return Group{}, err
    }

    if len(groups) == 0 {
        err := errors.New("No matching group found")
        sys.logger.Errorf( "Error fetching group: %s", err )
        return Group{}, err
    }

    if len(groups) > 1 {
        err := errors.New( "Multiple matching groups found, returning the first match" )
        sys.logger.Errorf("Error fetching group: %s", err )
        return groups[0], err
    }

    return groups[0], nil
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

    data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/projects/%v", projectID), nil, nil, []int{})
    if err != nil {
        return project, errors.Wrapf(err, "fetching project %v failed", projectID)
    }

    err = json.Unmarshal(data, &project)
    return project, err
}

// GetProjectsByNameAndGroup returns the project addressed by projectID from the Checkmarx backend which the user has access to
// Updated for Cx1
func (sys *SystemInstance) GetProjectsByNameAndGroup(projectName, groupID string) ([]Project, error) {
    sys.logger.Debugf("Getting projects with name %v of group %v...", projectName, groupID)
    var projects []Project
    header := http.Header{}
    header.Set("Accept-Type", "application/json")
    var data []byte
    var err error

    body := url.Values{}
    if len(groupID) > 0 {
        body.Add( "groups", groupID )
    }
    if len(projectName) > 0 {
        body.Add( "name", projectName )
    }

    if len(body) > 0 {
        data, err = sendRequest(sys, http.MethodGet, fmt.Sprintf("/projects/?%v", body.Encode()), nil, header, []int{404})
    } else {
        data, err = sendRequest(sys, http.MethodGet, "/projects/", nil, header, []int{404})
    }
    if err != nil {
        return projects, errors.Wrapf(err, "fetching project %v failed", projectName)
    }

    err = json.Unmarshal(data, &projects)
    return projects, err
}

// CreateProject creates a new project in the Checkmarx backend
// Updated for Cx1
func (sys *SystemInstance) CreateProject(projectName string, groupIDs []string) (Project, error) {
    var project Project
    jsonData := map[string]interface{}{
        "name":   projectName,
        "groups": groupIDs,
        "origin": cxOrigin,
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

// New for Cx1
func (sys *SystemInstance) GetUploadURI() (string,error) {
    sys.logger.Debug("Retrieving upload URI")
    header := http.Header{}
    header.Set("Content-Type", "application/json")
    resp, err := sendRequest(sys, http.MethodPost, "/uploads", nil, header, []int{})
    
    if err != nil {
        return "", errors.Wrap(err, "failed to get an upload uri")
    }

    responseData := make(map[string]string)
    json.Unmarshal( resp, &responseData )
    sys.logger.Debugf("Upload URI %s", responseData["url"] )

    return responseData["url"], nil
}

func (sys *SystemInstance) UploadProjectSourceCode(projectID string, zipFile string) (string, error) {
    sys.logger.Debug("Preparing to upload file...")

    // get URI
    uploadUri, err := sys.GetUploadURI()
    if err != nil {
        return "", err
    }

    // PUT request to uri
    // TODO - does this work?
    _, err = sendRequest(sys, http.MethodPut, uploadUri, strings.NewReader(zipFile), nil, []int{})
    if err != nil {
        return uploadUri, err
    }

    return uploadUri, nil
}


// Originally: func (sys *SystemInstance) UploadProjectSourceCode(projectID string, zipFile string) (string, error) 
// For Cx1: updated as there is no "per-project upload" anymore, high level steps are:
//    1. Get an upload URL
//  2. PUT a file there
//  3. Tell Cx1 to start a scan for a project using this uploaded file
// New for Cx1
func (sys *SystemInstance) UploadAndScanProjectSourceCode(projectID string, zipFile string) (Scan, error) {
    scan := Scan{}
    uploadUri, err := sys.UploadProjectSourceCode(projectID, zipFile)
    if err != nil {
        return scan, err
    }

    // Run a scan
    // ToDo: Full vs incremental?
    // ToDo: Preset?
    // ToDo: scan tags
    jsonBody := map[string]interface{}{
        "project" : map[string]interface{}{    "id" : projectID },
        "type": "upload",
        "handler" : map[string]interface{}{ "uploadurl" : uploadUri },
        "config" : []map[string]interface{}{
            map[string]interface{}{
                "type" : "sast",
                "value" : map[string]interface{}{
                    "incremental" : "false",
                    "presetName": "Checkmarx Default",
                },
            },
        },
    }
    jsonValue, err := json.Marshal( jsonBody )
    
    header := http.Header{}
    header.Set("Content-Type", "application/json")
    data, err := sendRequest(sys, http.MethodPost, "/scans", bytes.NewBuffer(jsonValue), header, []int{})
    if err != nil {
        return scan, errors.Wrapf(err, "failed to start a scan with project %v and url %v", projectID, uploadUri )
    }

    err = json.Unmarshal(data, &scan)
    return scan, err

}

// UpdateProjectExcludeSettings updates the exclude configuration of the project
// TODO
func (sys *SystemInstance) UpdateProjectExcludeSettings(projectID string, excludeFolders string, excludeFiles string) error {
    /*jsonData := map[string]string{
        "excludeFoldersPattern": excludeFolders,
        "excludeFilesPattern":   excludeFiles,
    }

    jsonValue, err := json.Marshal(jsonData)
    if err != nil {
        return errors.Wrap(err, "error marhalling project exclude settings")
    }

    header := http.Header{}
    header.Set("Content-Type", "application/json")
    _, err = sendRequest(sys, http.MethodPut, fmt.Sprintf("/projects/%v/sourceCode/excludeSettings", projectID), bytes.NewBuffer(jsonValue), header, []int{})
    if err != nil {
        return errors.Wrap(err, "request to checkmarx system failed")
    }*/

    return nil
}

// Updated for Cx1: GetPresets loads the preset values defined in the Checkmarx backend
func (sys *SystemInstance) GetPresets() ([]Preset,error) {
    sys.logger.Debug("Getting Presets...")
    var presets []Preset

    data, err := sendRequest(sys, http.MethodGet, "/queries/presets", nil, nil, []int{})
    if err != nil {
        sys.logger.Errorf("Fetching presets failed: %s", err)
        return presets, err
    }

    err = json.Unmarshal(data, &presets)
    return presets, err
}

// UpdateProjectConfiguration updates the configuration of the project addressed by projectID
// TODO
// unclear if this is still relevant? need to investigate usage.
func (sys *SystemInstance) UpdateProjectConfiguration(projectID string, presetID int, engineConfigurationID string) error {
    /*
        This behaves differently in Cx1. Configurations like the "Multi-Language Scan Mode" and the Preset are one of the 
            configurations returned by the /api/configurations/projects/?project-id=
            This returns an array of object 


    */


    /*engineConfigID, _ := strconv.Atoi(engineConfigurationID)

    var projectScanSettings ScanSettings
    header := http.Header{}
    header.Set("Content-Type", "application/json")
    data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/sast/scanSettings/%v", projectID), nil, header, []int{})
    if err != nil {
        // if an error happens, try to update the config anyway
        sys.logger.Warnf("Failed to fetch scan settings of project %v: %s", projectID, err)
    } else {
        // Check if the current project config needs to be updated
        json.Unmarshal(data, &projectScanSettings)
        if projectScanSettings.Preset.PresetID == presetID && projectScanSettings.EngineConfiguration.EngineConfigurationID == engineConfigID {
            sys.logger.Debugf("Project configuration does not need to be updated")
            return nil
        }
    }

    jsonData := map[string]interface{}{
        "projectId":             projectID,
        "presetId":              presetID,
        "engineConfigurationId": engineConfigID,
    }

    jsonValue, err := json.Marshal(jsonData)
    if err != nil {
        return errors.Wrapf(err, "error marshalling project data")
    }

    _, err = sendRequest(sys, http.MethodPost, "/sast/scanSettings", bytes.NewBuffer(jsonValue), header, []int{})
    if err != nil {
        return errors.Wrapf(err, "request to checkmarx system failed")
    }
    sys.logger.Debugf("Project configuration updated")*/

    return nil
}

// ScanProject triggers a scan on the project addressed by projectID
// TODO
// In Cx1, the request to scan a project is similar to the Zip-Scan above. Example:
/*
{
    "project": {
        "id": "{{Cx1_ProjectId}}"
    },
    "type": "git",
    "handler": {
        "branch": "master",
        "repoUrl": "https://github.com/michaelkubiaczyk/private_test"
    },
    "config": [
        {
            "type": "sast",
            "value": {
                "incremental": "false",
                "presetName": "Checkmarx Default"
            }
        }
    ]
}
*/
func (sys *SystemInstance) ScanProject(projectID string, isIncremental, isPublic, forceScan bool) (Scan, error) {
    scan := Scan{}

    /*
    jsonData := map[string]interface{}{
        "projectId":     projectID,
        "isIncremental": isIncremental,
        "isPublic":      isPublic,
        "forceScan":     forceScan,
        "comment":       "Scan From Golang Script",
    }

    jsonValue, _ := json.Marshal(jsonData)

    header := http.Header{}
    header.Set("cxOrigin", cxOrigin)
    header.Set("Content-Type", "application/json")
    data, err := sendRequest(sys, http.MethodPost, "/sast/scans", bytes.NewBuffer(jsonValue), header, []int{})
    if err != nil {
        sys.logger.Errorf("Failed to trigger scan of project %v: %s", projectID, err)
        return scan, errors.Wrapf(err, "Failed to trigger scan of project %v", projectID)
    }

    json.Unmarshal(data, &scan)*/
    return scan, nil
}

// GetScans returns all scan status on the project addressed by projectID
// Partially updated for Cx1 but the data structure to store the response is not yet fully defined
func (sys *SystemInstance) GetScans(projectID string) ([]Scan, error) {
    scans := []Scan{}
    body := url.Values{
        "projectId": {projectID},
        "offset":     {fmt.Sprintf("%v",0)},
        "limit":      {fmt.Sprintf("%v", 20)},
        "sort":        {"+created_at"},
    }

    header := http.Header{}
    header.Set("Accept-Type", "application/json")
    data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/scans?%v", body.Encode()), nil, header, []int{})
    if err != nil {
        sys.logger.Errorf("Failed to fetch scans of project %v: %s", projectID, err)
        return scans, errors.Wrapf(err, "failed to fetch scans of project %v", projectID)
    }

    json.Unmarshal(data, &scans)
    return scans, nil
}

// GetScanStatusAndDetail returns the status of the scan addressed by scanID
// Partially updated for Cx1 but the data structure to store the response is not yet fully defined
/*
func (sys *SystemInstance) GetScanStatusAndDetail(scanID string) (string, ScanStatusDetail) {
    var scanStatus ScanStatus
    header := http.Header{}
    header.Set("Accept-Type", "application/json")
    data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/scans/%v", scanID), nil, header, []int{})
    if err != nil {
        sys.logger.Errorf("Failed to get scan status for scanID %v: %s", scanID, err)
        return "Failed", ScanStatusDetail{}
    }

    json.Unmarshal(data, &scanStatus)
    return scanStatus.Status.Name, scanStatus.Status.Details //TODO after creating ScanStatus type
}/

// GetResults returns the results of the scan addressed by scanID
// Two options:
//   1. /results/?scan-id= &offset=0&limit=20&sort=%2Bstatus&sort=%2Bseverity
//   2. /sast-results/?scan-id=
// TODO - results are different in Cx1 and it is not a "ResultStatistics" object
func (sys *SystemInstance) GetResults(scanID string) ResultsStatistics {
    var results ResultsStatistics
    data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/results/?%v", scanID), nil, nil)
    if err != nil {
        sys.logger.Errorf("Failed to fetch scan results for scanID %v: %s", scanID, err)
        return results
    }

    json.Unmarshal(data, &results)
    return results
}

// RequestNewReport triggers the generation of a  report for a specific scan addressed by scanID
// TODO
func (sys *SystemInstance) RequestNewReport(scanID string, reportType string) (Report, error) {
    report := Report{}
    /* Example
    {
        "fileFormat": "pdf",
        "reportType": "ui",
        "reportName": "scan-report",
        "data": {
            "scanId": "{{Cx1_ScanId}}",
            "projectId": "{{Cx1_ProjectId}}",
            "branchName": "master",
            "sections": [
                "ScanSummary",
                "ExecutiveSummary",
                "ScanResults"
            ],
            "scanners": [
                "SAST"
            ],
            "host": ""
        }
    } // *//*
    jsonData := map[string]interface{}{
        "scanId":     scanID,
        "reportType": reportType,
        "comment":    "Scan report triggered by Piper",
    }

    jsonValue, _ := json.Marshal(jsonData)

    header := http.Header{}
    header.Set("cxOrigin", cxOrigin)
    header.Set("Content-Type", "application/json")
    data, err := sendRequest(sys, http.MethodPost, "/reports", bytes.NewBuffer(jsonValue), header, []int{})
    if err != nil {
        return report, errors.Wrapf(err, "Failed to trigger report generation for scan %v", scanID)
    }

    json.Unmarshal(data, &report) 
    return report, nil
}
*/
// GetReportStatus returns the status of the report generation process
// TODO - request is sent but the response is not yet stored, "ReportStatusResponse" structure not yet fully defined
/*
func (sys *SystemInstance) GetReportStatus(reportID int) (ReportStatusResponse, error) {
    var response ReportStatusResponse

    header := http.Header{}
    header.Set("Accept", "application/json")
    data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/reports/%v", reportID), nil, header, []int{})
    if err != nil {
        sys.logger.Errorf("Failed to fetch report status for reportID %v: %s", reportID, err)
        return response, errors.Wrapf(err, "failed to fetch report status for reportID %v", reportID)
    }

    json.Unmarshal(data, &response)
    return response, nil
}*/

// GetShortDescription returns the short description for an issue with a scanID and pathID
// TODO - I believe this is quite different in Cx1 as it is a per-query description rather than using the specific Scan & Path ID.
/*
func (sys *SystemInstance) GetShortDescription(scanID int, pathID int) (ShortDescription, error) {
    var shortDescription ShortDescription

    data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/sast/scans/%v/results/%v/shortDescription", scanID, pathID), nil, nil)
    if err != nil {
        sys.logger.Errorf("Failed to get short description for scanID %v and pathID %v: %s", scanID, pathID, err)
        return shortDescription, err
    }

    json.Unmarshal(data, &shortDescription)
    return shortDescription, nil
} */

// DownloadReport downloads the report addressed by reportID and returns the XML contents
// TODO
func (sys *SystemInstance) DownloadReport(reportID int) ([]byte, error) {
    header := http.Header{}
    header.Set("Accept", "application/json")
    data, err := sendRequest(sys, http.MethodGet, fmt.Sprintf("/reports/sastScan/%v", reportID), nil, header, []int{})
    if err != nil {
        return []byte{}, errors.Wrapf(err, "failed to download report with reportID %v", reportID)
    }
    return data, nil
}
