package fortify

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"

	"github.com/SAP/jenkins-library/pkg/fortify/models"
	"github.com/sirupsen/logrus"
)

// ReportsDirectory defines the subfolder for the Fortify reports which are generated
const ReportsDirectory = "fortify"

// System is the interface abstraction of a specific SystemInstance
type System interface {
	GetProjectByName(name string, autoCreate bool, projectVersion string) (*models.Project, error)
	GetProjectVersionDetailsByProjectIDAndVersionName(id int64, name string, autoCreate bool, projectName string) (*models.ProjectVersion, error)
	GetProjectVersionAttributesByProjectVersionID(id int64) ([]*models.Attribute, error)
	SetProjectVersionAttributesByProjectVersionID(id int64, attributes []*models.Attribute) ([]*models.Attribute, error)
	CreateProjectVersionIfNotExist(projectName, projectVersionName, description string) (*models.ProjectVersion, error)
	LookupOrCreateProjectVersionDetailsForPullRequest(projectID int64, masterProjectVersion *models.ProjectVersion, pullRequestName string) (*models.ProjectVersion, error)
	CreateProjectVersion(version *models.ProjectVersion) (*models.ProjectVersion, error)
	ProjectVersionCopyFromPartial(sourceID, targetID int64) error
	ProjectVersionCopyCurrentState(sourceID, targetID int64) error
	ProjectVersionCopyPermissions(sourceID, targetID int64) error
	CommitProjectVersion(id int64) (*models.ProjectVersion, error)
	MergeProjectVersionStateOfPRIntoMaster(downloadEndpoint, uploadEndpoint string, masterProjectID, masterProjectVersionID int64, pullRequestName string) error
	GetArtifactsOfProjectVersion(id int64) ([]*models.Artifact, error)
	GetFilterSetOfProjectVersionByTitle(id int64, title string) (*models.FilterSet, error)
	GetIssueFilterSelectorOfProjectVersionByName(id int64, names []string, options []string) (*models.IssueFilterSelectorSet, error)
	GetFilterSetByDisplayName(issueFilterSelectorSet *models.IssueFilterSelectorSet, name string) *models.IssueFilterSelector
	GetProjectIssuesByIDAndFilterSetGroupedBySelector(id int64, filter, filterSetGUID string, issueFilterSelectorSet *models.IssueFilterSelectorSet) ([]*models.ProjectVersionIssueGroup, error)
	ReduceIssueFilterSelectorSet(issueFilterSelectorSet *models.IssueFilterSelectorSet, names []string, options []string) *models.IssueFilterSelectorSet
	GetIssueStatisticsOfProjectVersion(id int64) ([]*models.IssueStatistics, error)
	GenerateQGateReport(projectID, projectVersionID, reportTemplateID int64, projectName, projectVersionName, reportFormat string) (*models.SavedReport, error)
	GetReportDetails(id int64) (*models.SavedReport, error)
	GetIssueDetails(projectVersionId int64, issueInstanceId string) ([]*models.ProjectVersionIssue, error)
	GetAllIssueDetails(projectVersionId int64) ([]*models.ProjectVersionIssue, error)
	GetIssueComments(parentId int64) ([]*models.IssueAuditComment, error)
	UploadResultFile(endpoint, file string, projectVersionID int64) error
	DownloadReportFile(endpoint string, reportID int64) ([]byte, error)
	DownloadResultFile(endpoint string, projectVersionID int64) ([]byte, error)
}

// SystemInstance is the specific instance
type SystemInstance struct {
	timeout    time.Duration
	token      string
	serverURL  string
	apiBaseURL string
	// apiClient issues the REST API requests without retries, matching the behavior
	// of the formerly used generated swagger client, while httpClient serves the
	// file upload and download endpoints with the default retry handling
	apiClient  *piperHttp.Client
	httpClient *piperHttp.Client
	logger     *logrus.Entry
}

// NewSystemInstance - creates an returns a new SystemInstance
func NewSystemInstance(serverURL, apiEndpoint, authToken, proxyUrl string, timeout time.Duration) *SystemInstance {
	// If serverURL ends in a trailing slash, UploadResultFile() will construct a URL with two or more
	// consecutive slashes and actually fail with a 503. https://github.com/SAP/jenkins-library/issues/1826
	// Also, since the step outputs a lot of URLs to the log, those will look nicer without redundant slashes.
	serverURL = strings.TrimRight(serverURL, "/")
	encodedAuthToken := base64EndodePlainToken(authToken)
	httpClientInstance := &piperHttp.Client{}
	httpClientOptions := piperHttp.ClientOptions{Token: "FortifyToken " + encodedAuthToken, TransportTimeout: timeout}

	if proxyUrl != "" {
		transportProxy, err := url.Parse(proxyUrl)
		if err != nil {
			log.Entry().Warningf("Failed to parse proxy url %v", proxyUrl)
		} else {
			httpClientOptions.TransportProxy = transportProxy
		}
	}

	httpClientInstance.SetOptions(httpClientOptions)
	return NewSystemInstanceForClient(httpClientInstance, serverURL, createAPIBaseURL(serverURL, apiEndpoint), encodedAuthToken, timeout)
}

// createAPIBaseURL constructs the base URL of the SSC REST API from the server URL and
// the API endpoint, defaulting to the https scheme and cleaning up redundant slashes.
func createAPIBaseURL(serverURL, apiEndpoint string) string {
	scheme, host := splitSchemeAndHost(serverURL)
	host, hostEndpoint := splitHostAndEndpoint(host)
	// Cleaning up any slashes here is mostly for cleaner log-output.
	hostEndpoint = strings.TrimRight(hostEndpoint, "/")
	apiEndpoint = strings.Trim(apiEndpoint, "/")
	baseURL := fmt.Sprintf("%v://%v", scheme, host)
	if len(hostEndpoint) > 0 {
		baseURL = fmt.Sprintf("%v/%v", baseURL, hostEndpoint)
	}
	if len(apiEndpoint) > 0 {
		baseURL = fmt.Sprintf("%v/%v", baseURL, apiEndpoint)
	}
	return baseURL
}

func splitSchemeAndHost(url string) (scheme, host string) {
	before, after, ok := strings.Cut(url, "://")
	if ok {
		scheme = before
		host = after
	} else {
		scheme = "https"
		host = url
	}
	return
}

func splitHostAndEndpoint(urlWithoutScheme string) (host, endpoint string) {
	before, after, ok := strings.Cut(urlWithoutScheme, "/")
	if ok {
		host = before
		endpoint = after
	} else {
		host = urlWithoutScheme
		endpoint = ""
	}
	return
}

func base64EndodePlainToken(authToken string) (encodedAuthToken string) {
	isEncoded := strings.Index(authToken, "-") < 0
	if isEncoded {
		return authToken
	}
	return base64.StdEncoding.EncodeToString([]byte(authToken))
}

// NewSystemInstanceForClient - creates a new SystemInstance
func NewSystemInstanceForClient(httpClientInstance *piperHttp.Client, serverURL, apiBaseURL, authToken string, requestTimeout time.Duration) *SystemInstance {
	apiClientInstance := &piperHttp.Client{}
	apiClientInstance.SetOptions(piperHttp.ClientOptions{
		Token:            "FortifyToken " + authToken,
		TransportTimeout: requestTimeout,
		MaxRetries:       -1,
	})
	return &SystemInstance{
		timeout:    requestTimeout,
		token:      authToken,
		serverURL:  serverURL,
		apiBaseURL: apiBaseURL,
		apiClient:  apiClientInstance,
		httpClient: httpClientInstance,
		logger:     log.Entry().WithField("package", "SAP/jenkins-library/pkg/fortify"),
	}
}

// apiResponsePayload is the generic response envelope of the SSC REST API
type apiResponsePayload[T any] struct {
	Data  T     `json:"data"`
	Count int32 `json:"count"`
}

// sendAPIRequest issues a request against the SSC REST API and decodes the enveloped response payload
func sendAPIRequest[T any](sys *SystemInstance, method, endpoint string, query url.Values, requestBody any) (*apiResponsePayload[T], error) {
	requestURL := sys.apiBaseURL + endpoint
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}
	header := http.Header{}
	header.Set("Accept", "application/json")
	var bodyReader io.Reader
	if requestBody != nil {
		// a json.Encoder appends a trailing newline, producing the same request
		// bodies as the formerly used generated swagger client
		body := &bytes.Buffer{}
		if err := json.NewEncoder(body).Encode(requestBody); err != nil {
			return nil, err
		}
		bodyReader = body
		header.Set("Content-Type", "application/json")
	}
	response, err := sys.apiClient.SendRequest(method, requestURL, bodyReader, header, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := &apiResponsePayload[T]{}
	// a json.Decoder decodes the first JSON value and tolerates an empty body as
	// well as trailing content, matching the behavior of the formerly used
	// generated swagger client
	if err := json.NewDecoder(response.Body).Decode(result); err != nil && err != io.EOF {
		return nil, err
	}
	return result, nil
}

// GetProjectByName returns the project identified by the name provided
// autoCreate and projectVersion parameters only used if autoCreate=true
func (sys *SystemInstance) GetProjectByName(projectName string, autoCreate bool, projectVersionName string) (*models.Project, error) {
	query := url.Values{"q": {fmt.Sprintf(`name:"%v"`, projectName)}}
	result, err := sendAPIRequest[[]*models.Project](sys, http.MethodGet, "/projects", query, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	for _, project := range result.Data {
		if *project.Name == projectName {
			return project, nil
		}
	}

	// Project with specified name was NOT found, check if autoCreate flag is set, if not stop otherwise create it automatically
	if !autoCreate {
		log.SetErrorCategory(log.ErrorConfiguration)
		return nil, fmt.Errorf("Project with name %v not found in backend and automatic creation not enabled", projectName)
	}

	log.Entry().Debugf("No projects found with name: %v auto-creating one now...", projectName)
	projectVersion, err := sys.CreateProjectVersionIfNotExist(projectName, projectVersionName, "Created by Go script")
	if err != nil {
		return nil, fmt.Errorf("failed to auto-create new project: %w", err)
	}
	log.Entry().Debugf("Finished creating project: %v", projectVersion)
	return projectVersion.Project, nil
}

// GetProjectVersionDetailsByProjectIDAndVersionName returns the project version details of the project version identified by the id and project versionname
// projectName parameter is only used if autoCreate=true
func (sys *SystemInstance) GetProjectVersionDetailsByProjectIDAndVersionName(id int64, versionName string, autoCreate bool, projectName string) (*models.ProjectVersion, error) {
	query := url.Values{"q": {fmt.Sprintf(`name:"%v"`, versionName)}}
	result, err := sendAPIRequest[[]*models.ProjectVersion](sys, http.MethodGet, fmt.Sprintf("/projects/%v/versions", id), query, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}

	if result.Count > 0 && len(result.Data) > 0 {
		projectVersion := result.Data[0]
		return projectVersion, nil
	}
	// projectVersion not found for specified project id and name, check if autoCreate is enabled
	if !autoCreate {
		log.SetErrorCategory(log.ErrorConfiguration)
		return nil, fmt.Errorf("Project version with name %v not found in project with ID %v and automatic creation not enabled", versionName, id)
	}

	log.Entry().Debugf("Could not find project version with name %v under project %v auto-creating one now...", versionName, projectName)
	version, err := sys.CreateProjectVersionIfNotExist(projectName, versionName, "Created by Go script")
	if err != nil {
		return nil, fmt.Errorf("failed to auto-create project version: %v for project %v: %w", versionName, projectName, err)
	}
	log.Entry().Debugf("Successfully created project version %v for project %v", versionName, projectName)
	return version, nil
}

// CreateProjectVersionIfNotExist creates a new ProjectVersion if it does not already exist.
// If the projectName also does not exist, it will create that as well.
func (sys *SystemInstance) CreateProjectVersionIfNotExist(projectName, projectVersionName, description string) (*models.ProjectVersion, error) {
	var projectID int64 = 0
	// check if project with projectName exists
	projectResp, err := sys.GetProjectByName(projectName, false, "")
	if err == nil {
		// project already exists, all we need to do is append a new ProjectVersion to it
		// save the project id for later
		projectID = projectResp.ID
	}

	issueTemplateID := "4c5799c9-1940-4abe-b57a-3bcad88eb041"
	active := true
	committed := true
	projectVersionDto := &models.ProjectVersion{
		Name:            &projectVersionName,
		Description:     &description,
		IssueTemplateID: &issueTemplateID,
		Active:          &active,
		Committed:       &committed,
		Project:         &models.Project{ID: projectID},
	}

	if projectVersionDto.Project.ID == 0 { // project does not exist, set one up
		projectVersionDto.Project = &models.Project{
			Name:            &projectName,
			Description:     description,
			IssueTemplateID: &issueTemplateID,
		}
	}
	projectVersion, err := sys.CreateProjectVersion(projectVersionDto)
	if err != nil {
		return nil, fmt.Errorf("Failed to create new project version %v for projectName %v: %w", projectVersionName, projectName, err)
	}
	_, err = sys.CommitProjectVersion(projectVersion.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to commit project version %v: %v: %w", projectVersion.ID, err, err)
	}
	return projectVersion, nil
}

// LookupOrCreateProjectVersionDetailsForPullRequest looks up a project version for pull requests or creates it from scratch
func (sys *SystemInstance) LookupOrCreateProjectVersionDetailsForPullRequest(projectID int64, masterProjectVersion *models.ProjectVersion, pullRequestName string) (*models.ProjectVersion, error) {
	projectVersion, _ := sys.GetProjectVersionDetailsByProjectIDAndVersionName(projectID, pullRequestName, false, "")
	if nil != projectVersion {
		return projectVersion, nil
	}

	newVersion := &models.ProjectVersion{}
	newVersion.Name = &pullRequestName
	newVersion.Description = masterProjectVersion.Description
	newVersion.Active = masterProjectVersion.Active
	newVersion.Committed = masterProjectVersion.Committed
	newVersion.Project = &models.Project{}
	newVersion.Project.Name = masterProjectVersion.Project.Name
	newVersion.Project.Description = masterProjectVersion.Project.Description
	newVersion.Project.ID = masterProjectVersion.Project.ID
	newVersion.IssueTemplateID = masterProjectVersion.IssueTemplateID

	projectVersion, err := sys.CreateProjectVersion(newVersion)
	if err != nil {
		return nil, fmt.Errorf("Failed to create new project version for pull request %v: %w", pullRequestName, err)
	}
	attributes, err := sys.GetProjectVersionAttributesByProjectVersionID(masterProjectVersion.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to load project version attributes for master project version %v: %w", masterProjectVersion.ID, err)
	}
	for _, attribute := range attributes {
		attribute.ID = 0
	}
	_, err = sys.SetProjectVersionAttributesByProjectVersionID(projectVersion.ID, attributes)
	if err != nil {
		return nil, fmt.Errorf("Failed to update project version attributes for pull request project version %v: %w", projectVersion.ID, err)
	}
	err = sys.ProjectVersionCopyFromPartial(masterProjectVersion.ID, projectVersion.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to copy from partial of project version %v to %v: %w", masterProjectVersion.ID, projectVersion.ID, err)
	}
	_, err = sys.CommitProjectVersion(projectVersion.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to commit project version %v: %v: %w", projectVersion.ID, err, err)
	}
	err = sys.ProjectVersionCopyCurrentState(masterProjectVersion.ID, projectVersion.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to copy current state of project version %v to %v: %w", masterProjectVersion.ID, projectVersion.ID, err)
	}
	err = sys.ProjectVersionCopyPermissions(masterProjectVersion.ID, projectVersion.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to copy permissions of project version %v to %v: %w", masterProjectVersion.ID, projectVersion.ID, err)
	}
	return projectVersion, nil
}

// GetProjectVersionAttributesByProjectVersionID returns the project version attributes of the project version identified by the id
func (sys *SystemInstance) GetProjectVersionAttributesByProjectVersionID(id int64) ([]*models.Attribute, error) {
	result, err := sendAPIRequest[[]*models.Attribute](sys, http.MethodGet, fmt.Sprintf("/projectVersions/%v/attributes", id), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// SetProjectVersionAttributesByProjectVersionID sets the project version attributes of the project version identified by the id
func (sys *SystemInstance) SetProjectVersionAttributesByProjectVersionID(id int64, attributes []*models.Attribute) ([]*models.Attribute, error) {
	result, err := sendAPIRequest[[]*models.Attribute](sys, http.MethodPut, fmt.Sprintf("/projectVersions/%v/attributes", id), nil, attributes)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// CreateProjectVersion creates the project version with the provided details
func (sys *SystemInstance) CreateProjectVersion(version *models.ProjectVersion) (*models.ProjectVersion, error) {
	result, err := sendAPIRequest[*models.ProjectVersion](sys, http.MethodPost, "/projectVersions", nil, version)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// ProjectVersionCopyFromPartial copies parts of the source project version to the target project version identified by their ids
func (sys *SystemInstance) ProjectVersionCopyFromPartial(sourceID, targetID int64) error {
	enable := true
	settings := models.ProjectVersionCopyPartialRequest{
		ProjectVersionID:            &targetID,
		PreviousProjectVersionID:    &sourceID,
		CopyAnalysisProcessingRules: &enable,
		CopyBugTrackerConfiguration: &enable,
		CopyCustomTags:              &enable,
	}
	_, err := sendAPIRequest[json.RawMessage](sys, http.MethodPost, "/projectVersions/action/copyFromPartial", nil, &settings)
	if err != nil {
		return fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return nil
}

// ProjectVersionCopyCurrentState copies the project version state of sourceID into the new project version addressed by targetID
func (sys *SystemInstance) ProjectVersionCopyCurrentState(sourceID, targetID int64) error {
	settings := models.ProjectVersionCopyCurrentStateRequest{
		ProjectVersionID:         &targetID,
		PreviousProjectVersionID: &sourceID,
	}
	_, err := sendAPIRequest[json.RawMessage](sys, http.MethodPost, "/projectVersions/action/copyCurrentState", nil, &settings)
	if err != nil {
		return fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return nil
}

func (sys *SystemInstance) getAuthEntityOfProjectVersion(id int64) ([]*models.AuthenticationEntity, error) {
	query := url.Values{"embed": {"roles"}}
	result, err := sendAPIRequest[[]*models.AuthenticationEntity](sys, http.MethodGet, fmt.Sprintf("/projectVersions/%v/authEntities", id), query, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

func (sys *SystemInstance) updateCollectionAuthEntityOfProjectVersion(id int64, data []*models.AuthenticationEntity) error {
	_, err := sendAPIRequest[json.RawMessage](sys, http.MethodPut, fmt.Sprintf("/projectVersions/%v/authEntities", id), nil, data)
	if err != nil {
		return fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return nil
}

// ProjectVersionCopyPermissions copies the authentication entity of the project version addressed by sourceID to the one of targetID
func (sys *SystemInstance) ProjectVersionCopyPermissions(sourceID, targetID int64) error {
	result, err := sys.getAuthEntityOfProjectVersion(sourceID)
	if err != nil {
		return err
	}
	err = sys.updateCollectionAuthEntityOfProjectVersion(targetID, result)
	if err != nil {
		return err
	}
	return nil
}

func (sys *SystemInstance) updateProjectVersionDetails(id int64, details *models.ProjectVersion) (*models.ProjectVersion, error) {
	result, err := sendAPIRequest[*models.ProjectVersion](sys, http.MethodPut, fmt.Sprintf("/projectVersions/%v", id), nil, details)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// CommitProjectVersion commits the project version with the provided id
func (sys *SystemInstance) CommitProjectVersion(id int64) (*models.ProjectVersion, error) {
	enabled := true
	update := models.ProjectVersion{Committed: &enabled}
	return sys.updateProjectVersionDetails(id, &update)
}

func (sys *SystemInstance) inactivateProjectVersion(id int64) (*models.ProjectVersion, error) {
	enabled := true
	disabled := false
	update := models.ProjectVersion{Committed: &enabled, Active: &disabled}
	return sys.updateProjectVersionDetails(id, &update)
}

// GetArtifactsOfProjectVersion returns the list of artifacts related to the project version addressed with id
func (sys *SystemInstance) GetArtifactsOfProjectVersion(id int64) ([]*models.Artifact, error) {
	query := url.Values{"embed": {"scans"}}
	result, err := sendAPIRequest[[]*models.Artifact](sys, http.MethodGet, fmt.Sprintf("/projectVersions/%v/artifacts", id), query, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// MergeProjectVersionStateOfPRIntoMaster merges the PR project version's fpr result file into the master project version
func (sys *SystemInstance) MergeProjectVersionStateOfPRIntoMaster(downloadEndpoint, uploadEndpoint string, masterProjectID, masterProjectVersionID int64, pullRequestName string) error {
	log.Entry().Debugf("Looking up project version with name '%v' to merge audit status into master version", pullRequestName)
	prProjectVersion, _ := sys.GetProjectVersionDetailsByProjectIDAndVersionName(masterProjectID, pullRequestName, false, "")
	if nil != prProjectVersion {
		log.Entry().Debugf("Found project version with ID '%v', starting transfer", prProjectVersion.ID)
		data, err := sys.DownloadResultFile(downloadEndpoint, prProjectVersion.ID)
		if err != nil {
			return fmt.Errorf("Failed to download current state FPR of PR project version %v: %w", prProjectVersion.ID, err)
		}
		err = sys.uploadResultFileContent(uploadEndpoint, "prMergeTransfer.fpr", bytes.NewReader(data), masterProjectVersionID)
		if err != nil {
			return fmt.Errorf("Failed to upload PR project version state to master project version %v: %w", masterProjectVersionID, err)
		}
		_, err = sys.inactivateProjectVersion(prProjectVersion.ID)
		if err != nil {
			log.Entry().Warnf("Failed to inactivate merged PR project version %v", prProjectVersion.ID)
		}
	} else {
		log.Entry().Debug("No related project version found in SSC")
	}
	return nil
}

// GetFilterSetOfProjectVersionByTitle returns the filter set with the given title related to the project version addressed with id, if no title is provided the default filter set will be returned
func (sys *SystemInstance) GetFilterSetOfProjectVersionByTitle(id int64, title string) (*models.FilterSet, error) {
	result, err := sendAPIRequest[[]*models.FilterSet](sys, http.MethodGet, fmt.Sprintf("/projectVersions/%v/filterSets", id), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	var defaultFilterSet *models.FilterSet
	for _, filterSet := range result.Data {
		if len(title) > 0 && filterSet.Title == title {
			return filterSet, nil
		}
		if filterSet.DefaultFilterSet {
			defaultFilterSet = filterSet
		}
	}
	if len(title) > 0 {
		log.Entry().Warnf("Failed to load filter set with title '%v', falling back to default filter set", title)
	}
	if nil != defaultFilterSet {
		return defaultFilterSet, nil
	}
	return nil, fmt.Errorf("Failed to identify requested filter set and default filter")
}

// GetIssueFilterSelectorOfProjectVersionByName returns the groupings with the given names related to the project version addressed with id
func (sys *SystemInstance) GetIssueFilterSelectorOfProjectVersionByName(id int64, names []string, options []string) (*models.IssueFilterSelectorSet, error) {
	result, err := sendAPIRequest[*models.IssueFilterSelectorSet](sys, http.MethodGet, fmt.Sprintf("/projectVersions/%v/issueSelectorSet", id), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return sys.ReduceIssueFilterSelectorSet(result.Data, names, options), nil
}

// ReduceIssueFilterSelectorSet filters the set to the relevant filter display names
func (sys *SystemInstance) ReduceIssueFilterSelectorSet(issueFilterSelectorSet *models.IssueFilterSelectorSet, names []string, options []string) *models.IssueFilterSelectorSet {
	groupingList := []*models.IssueSelector{}
	if issueFilterSelectorSet.GroupBySet != nil {
		for _, group := range issueFilterSelectorSet.GroupBySet {
			if slices.Contains(names, *group.DisplayName) {
				log.Entry().Debugf("adding new grouping '%v' to reduced list", *group.DisplayName)
				groupingList = append(groupingList, group)
			}
		}
	}
	filterList := []*models.IssueFilterSelector{}
	if issueFilterSelectorSet.FilterBySet != nil {
		for _, filter := range issueFilterSelectorSet.FilterBySet {
			if slices.Contains(names, filter.DisplayName) {
				newFilter := &models.IssueFilterSelector{}
				newFilter.DisplayName = filter.DisplayName
				newFilter.Description = filter.Description
				newFilter.EntityType = filter.EntityType
				newFilter.FilterSelectorType = filter.FilterSelectorType
				newFilter.GUID = filter.GUID
				newFilter.Value = filter.Value
				newFilter.SelectorOptions = []*models.SelectorOption{}
				for _, option := range filter.SelectorOptions {
					if (nil != options && slices.Contains(options, option.DisplayName)) || options == nil || len(options) == 0 {
						log.Entry().Debugf("adding selector option '%v' to list for filter selector '%v'", option.DisplayName, newFilter.DisplayName)
						newFilter.SelectorOptions = append(newFilter.SelectorOptions, option)
					}
				}
				log.Entry().Debugf("adding new filter '%v' to reduced list with selector options '%v'", newFilter.DisplayName, newFilter.SelectorOptions)
				filterList = append(filterList, newFilter)
			}
		}
	}
	return &models.IssueFilterSelectorSet{GroupBySet: groupingList, FilterBySet: filterList}
}

// GetFilterSetByDisplayName returns the set identified by the provided name or nil
func (sys *SystemInstance) GetFilterSetByDisplayName(issueFilterSelectorSet *models.IssueFilterSelectorSet, name string) *models.IssueFilterSelector {
	if issueFilterSelectorSet.FilterBySet != nil {
		for _, filter := range issueFilterSelectorSet.FilterBySet {
			if filter.DisplayName == name {
				return filter
			}
		}
	}
	return nil
}

func (sys *SystemInstance) getIssuesOfProjectVersion(id int64, filter, filterset, groupingtype string) ([]*models.ProjectVersionIssueGroup, error) {
	query := url.Values{
		"showsuppressed": {"true"},
		"filterset":      {filterset},
		"groupingtype":   {groupingtype},
	}
	if len(filter) > 0 {
		query.Set("filter", filter)
	}
	result, err := sendAPIRequest[[]*models.ProjectVersionIssueGroup](sys, http.MethodGet, fmt.Sprintf("/projectVersions/%v/issueGroups", id), query, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// GetProjectIssuesByIDAndFilterSetGroupedBySelector returns issues of the project version addressed with id filtered with the respective set and grouped by the issue filter selector grouping
func (sys *SystemInstance) GetProjectIssuesByIDAndFilterSetGroupedBySelector(id int64, filter, filterSetGUID string, issueFilterSelectorSet *models.IssueFilterSelectorSet) ([]*models.ProjectVersionIssueGroup, error) {
	groupingTypeGUID := ""
	if issueFilterSelectorSet != nil {
		groupingTypeGUID = *issueFilterSelectorSet.GroupBySet[0].GUID
	}

	result, err := sys.getIssuesOfProjectVersion(id, filter, filterSetGUID, groupingTypeGUID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetIssueStatisticsOfProjectVersion returns the issue statistics related to the project version addressed with id
func (sys *SystemInstance) GetIssueStatisticsOfProjectVersion(id int64) ([]*models.IssueStatistics, error) {
	result, err := sendAPIRequest[[]*models.IssueStatistics](sys, http.MethodGet, fmt.Sprintf("/projectVersions/%v/issueStatistics", id), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// GenerateQGateReport returns the issue statistics related to the project version addressed with id
func (sys *SystemInstance) GenerateQGateReport(projectID, projectVersionID, reportTemplateID int64, projectName, projectVersionName, reportFormat string) (*models.SavedReport, error) {
	paramIdentifier := "projectVersionId"
	paramType := "SINGLE_PROJECT"
	paramName := "Q-gate-report"
	reportType := "PORTFOLIO"
	inputReportParameters := []*models.InputReportParameter{{Name: &paramName, Identifier: &paramIdentifier, ParamValue: projectVersionID, Type: &paramType}}
	reportProjectVersions := []*models.ReportProjectVersion{{ID: projectVersionID, Name: projectVersionName}}
	reportProjects := []*models.ReportProject{{ID: projectID, Name: projectName, Versions: reportProjectVersions}}
	report := models.SavedReport{Name: fmt.Sprintf("FortifyReport: %v:%v", projectName, projectVersionName), Type: &reportType, ReportDefinitionID: &reportTemplateID, Format: &reportFormat, Projects: reportProjects, InputReportParameters: inputReportParameters}
	result, err := sendAPIRequest[*models.SavedReport](sys, http.MethodPost, "/reports", nil, &report)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// GetReportDetails returns the details of the report addressed with id
func (sys *SystemInstance) GetReportDetails(id int64) (*models.SavedReport, error) {
	result, err := sendAPIRequest[*models.SavedReport](sys, http.MethodGet, fmt.Sprintf("/reports/%v", id), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// GetIssueDetails returns the details of an issue with its issueInstanceId and projectVersionId
func (sys *SystemInstance) GetIssueDetails(projectVersionId int64, issueInstanceId string) ([]*models.ProjectVersionIssue, error) {
	query := url.Values{
		"q":              {issueInstanceId},
		"qm":             {"issues"},
		"showsuppressed": {"true"},
	}
	result, err := sendAPIRequest[[]*models.ProjectVersionIssue](sys, http.MethodGet, fmt.Sprintf("/projectVersions/%v/issues", projectVersionId), query, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// GetAllIssueDetails returns the details of all issues of the project with id projectVersionId
func (sys *SystemInstance) GetAllIssueDetails(projectVersionId int64) ([]*models.ProjectVersionIssue, error) {
	query := url.Values{
		"limit":          {"-1"},
		"showsuppressed": {"true"},
	}
	result, err := sendAPIRequest[[]*models.ProjectVersionIssue](sys, http.MethodGet, fmt.Sprintf("/projectVersions/%v/issues", projectVersionId), query, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// GetIssueComments returns the details of an issue comments with its unique parentId
func (sys *SystemInstance) GetIssueComments(parentId int64) ([]*models.IssueAuditComment, error) {
	result, err := sendAPIRequest[[]*models.IssueAuditComment](sys, http.MethodGet, fmt.Sprintf("/issues/%v/comments", parentId), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

func (sys *SystemInstance) invalidateFileTokens() error {
	log.Entry().Debug("invalidating file tokens")
	_, err := sendAPIRequest[json.RawMessage](sys, http.MethodDelete, "/fileTokens", nil, nil)
	if err != nil {
		return fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return nil
}

func (sys *SystemInstance) getFileToken(tokenType string) (*models.FileToken, error) {
	log.Entry().Debugf("fetching file token of type %v", tokenType)
	token := models.FileToken{FileTokenType: &tokenType}
	result, err := sendAPIRequest[*models.FileToken](sys, http.MethodPost, "/fileTokens", nil, &token)
	if err != nil {
		return nil, fmt.Errorf("Error from url %s %w", sys.serverURL, err)
	}
	return result.Data, nil
}

// UploadResultFile uploads a fpr file to the fortify backend
func (sys *SystemInstance) UploadResultFile(endpoint, file string, projectVersionID int64) error {
	fileHandle, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Unable to locate file %v: %w", file, err)
	}
	defer fileHandle.Close()

	return sys.uploadResultFileContent(endpoint, file, fileHandle, projectVersionID)
}

func (sys *SystemInstance) uploadResultFileContent(endpoint, file string, fileContent io.Reader, projectVersionID int64) error {
	token, err := sys.getFileToken("UPLOAD")
	if err != nil {
		return err
	}
	defer sys.invalidateFileTokens()

	header := http.Header{}
	header.Add("Cache-Control", "no-cache, no-store, must-revalidate")
	header.Add("Pragma", "no-cache")

	formFields := map[string]string{}
	formFields["entityId"] = fmt.Sprintf("%v", projectVersionID)

	_, err = sys.httpClient.Upload(piperHttp.UploadRequestData{
		Method:        http.MethodPost,
		URL:           fmt.Sprintf("%v%v?mat=%v", sys.serverURL, endpoint, token.Token),
		File:          file,
		FileFieldName: "file",
		FormFields:    formFields,
		FileContent:   fileContent,
		Header:        header,
	})
	return err
}

// DownloadFile downloads a file from Fortify backend
func (sys *SystemInstance) downloadFile(endpoint, method, acceptType, tokenType string, fileID int64) ([]byte, error) {
	token, err := sys.getFileToken(tokenType)
	if err != nil {
		return nil, fmt.Errorf("Error fetching file token: %w", err)
	}
	defer sys.invalidateFileTokens()

	header := http.Header{}
	header.Add("Cache-Control", "no-cache, no-store, must-revalidate")
	header.Add("Pragma", "no-cache")
	header.Add("Accept", acceptType)
	header.Add("Content-Type", "application/form-data")
	body := url.Values{
		"id":  {fmt.Sprintf("%v", fileID)},
		"mat": {token.Token},
	}
	var response *http.Response
	if method == http.MethodGet {
		response, err = sys.httpClient.SendRequest(method, fmt.Sprintf("%v%v?%v", sys.serverURL, endpoint, body.Encode()), nil, header, nil)
	} else {
		response, err = sys.httpClient.SendRequest(method, fmt.Sprintf("%v%v", sys.serverURL, endpoint), strings.NewReader(body.Encode()), header, nil)
	}
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Error reading the response data: %w", err)
	}
	return data, nil
}

// DownloadReportFile downloads a report file from Fortify backend
func (sys *SystemInstance) DownloadReportFile(endpoint string, reportID int64) ([]byte, error) {
	data, err := sys.downloadFile(endpoint, http.MethodGet, "application/octet-stream", "REPORT_FILE", reportID)
	if err != nil {
		return nil, fmt.Errorf("Error downloading report file: %w", err)
	}
	return data, nil
}

// DownloadResultFile downloads a result file from Fortify backend
func (sys *SystemInstance) DownloadResultFile(endpoint string, projectVersionID int64) ([]byte, error) {
	data, err := sys.downloadFile(endpoint, http.MethodGet, "application/zip", "DOWNLOAD", projectVersionID)
	if err != nil {
		return nil, fmt.Errorf("Error downloading result file: %w", err)
	}
	return data, nil
}
