package fortify

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	ff "github.com/piper-validation/fortify-client-go/fortify"
	"github.com/piper-validation/fortify-client-go/fortify/artifact_of_project_version_controller"
	"github.com/piper-validation/fortify-client-go/fortify/attribute_of_project_version_controller"
	"github.com/piper-validation/fortify-client-go/fortify/auth_entity_of_project_version_controller"
	"github.com/piper-validation/fortify-client-go/fortify/file_token_controller"
	"github.com/piper-validation/fortify-client-go/fortify/filter_set_of_project_version_controller"
	"github.com/piper-validation/fortify-client-go/fortify/issue_group_of_project_version_controller"
	"github.com/piper-validation/fortify-client-go/fortify/issue_selector_set_of_project_version_controller"
	"github.com/piper-validation/fortify-client-go/fortify/issue_statistics_of_project_version_controller"
	"github.com/piper-validation/fortify-client-go/fortify/project_controller"
	"github.com/piper-validation/fortify-client-go/fortify/project_version_controller"
	"github.com/piper-validation/fortify-client-go/fortify/project_version_of_project_controller"
	"github.com/piper-validation/fortify-client-go/fortify/saved_report_controller"
	"github.com/piper-validation/fortify-client-go/models"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

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
	UploadResultFile(endpoint, file string, projectVersionID int64) error
	DownloadReportFile(endpoint string, reportID int64) ([]byte, error)
	DownloadResultFile(endpoint string, projectVersionID int64) ([]byte, error)
}

// SystemInstance is the specific instance
type SystemInstance struct {
	timeout    time.Duration
	token      string
	serverURL  string
	client     *ff.Fortify
	httpClient *piperHttp.Client
	logger     *logrus.Entry
}

// NewSystemInstance - creates an returns a new SystemInstance
func NewSystemInstance(serverURL, apiEndpoint, authToken string, timeout time.Duration) *SystemInstance {
	// If serverURL ends in a trailing slash, UploadResultFile() will construct a URL with two or more
	// consecutive slashes and actually fail with a 503. https://github.com/SAP/jenkins-library/issues/1826
	// Also, since the step outputs a lot of URLs to the log, those will look nicer without redundant slashes.
	serverURL = strings.TrimRight(serverURL, "/")
	format := strfmt.Default
	dateTimeFormat := models.Iso8601MilliDateTime{}
	format.Add("datetime", &dateTimeFormat, models.IsDateTime)
	clientInstance := ff.NewHTTPClientWithConfig(format, createTransportConfig(serverURL, apiEndpoint))
	httpClientInstance := &piperHttp.Client{}
	httpClientOptions := piperHttp.ClientOptions{Token: "FortifyToken " + authToken, TransportTimeout: timeout}
	httpClientInstance.SetOptions(httpClientOptions)

	return NewSystemInstanceForClient(clientInstance, httpClientInstance, serverURL, authToken, timeout)
}

func createTransportConfig(serverURL, apiEndpoint string) *ff.TransportConfig {
	scheme, host := splitSchemeAndHost(serverURL)
	host, hostEndpoint := splitHostAndEndpoint(host)
	// Cleaning up any slashes here is mostly for cleaner log-output.
	hostEndpoint = strings.TrimRight(hostEndpoint, "/")
	apiEndpoint = strings.Trim(apiEndpoint, "/")
	return &ff.TransportConfig{
		Host:     host,
		Schemes:  []string{scheme},
		BasePath: fmt.Sprintf("%v/%v", hostEndpoint, apiEndpoint)}
}

func splitSchemeAndHost(url string) (scheme, host string) {
	schemeEnd := strings.Index(url, "://")
	if schemeEnd >= 0 {
		scheme = url[0:schemeEnd]
		host = url[schemeEnd+3:]
	} else {
		scheme = "https"
		host = url
	}
	return
}

func splitHostAndEndpoint(urlWithoutScheme string) (host, endpoint string) {
	hostEnd := strings.Index(urlWithoutScheme, "/")
	if hostEnd >= 0 {
		host = urlWithoutScheme[0:hostEnd]
		endpoint = urlWithoutScheme[hostEnd+1:]
	} else {
		host = urlWithoutScheme
		endpoint = ""
	}
	return
}

// NewSystemInstanceForClient - creates a new SystemInstance
func NewSystemInstanceForClient(clientInstance *ff.Fortify, httpClientInstance *piperHttp.Client, serverURL, authToken string, requestTimeout time.Duration) *SystemInstance {
	return &SystemInstance{
		timeout:    requestTimeout,
		token:      authToken,
		serverURL:  serverURL,
		client:     clientInstance,
		httpClient: httpClientInstance,
		logger:     log.Entry().WithField("package", "SAP/jenkins-library/pkg/fortify"),
	}
}

// AuthenticateRequest authenticates the request
func (sys *SystemInstance) AuthenticateRequest(req runtime.ClientRequest, formats strfmt.Registry) error {
	req.SetHeaderParam("Authorization", fmt.Sprintf("FortifyToken %v", sys.token))
	return nil
}

// GetProjectByName returns the project identified by the name provided
// autoCreate and projectVersion parameters only used if autoCreate=true
func (sys *SystemInstance) GetProjectByName(projectName string, autoCreate bool, projectVersionName string) (*models.Project, error) {
	nameParam := fmt.Sprintf("name=%v", projectName)
	fullText := true
	params := &project_controller.ListProjectParams{Q: &nameParam, Fulltextsearch: &fullText}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.ProjectController.ListProject(params, sys)
	if err != nil {
		return nil, err
	}
	for _, project := range result.GetPayload().Data {
		if *project.Name == projectName {
			return project, nil
		}
	}

	// Project with specified name was NOT found, check if autoCreate flag is set, if not stop otherwise create it automatically
	if !autoCreate {
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
	nameParam := fmt.Sprintf("name=%v", versionName)
	fullText := true
	params := &project_version_of_project_controller.ListProjectVersionOfProjectParams{ParentID: id, Q: &nameParam, Fulltextsearch: &fullText}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.ProjectVersionOfProjectController.ListProjectVersionOfProject(params, sys)
	if err != nil {
		return nil, err
	}
	for _, projectVersion := range result.GetPayload().Data {
		if *projectVersion.Name == versionName {
			return projectVersion, nil
		}
	}
	// projectVersion not found for specified project id and name, check if autoCreate is enabled
	if !autoCreate {
		return nil, errors.New(fmt.Sprintf("Project version with name %v not found in project with ID %v and automatic creation not enabled", versionName, id))
	}

	log.Entry().Debugf("Could not find project version with name %v under project %v auto-creating one now...", versionName, projectName)
	version, err := sys.CreateProjectVersionIfNotExist(projectName, versionName, "Created by Go script")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to auto-create project version: %v for project %v", versionName, projectName)
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
		return nil, errors.Wrapf(err, "Failed to create new project version %v for projectName %v", projectVersionName, projectName)
	}
	_, err = sys.CommitProjectVersion(projectVersion.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to commit project version %v: %v", projectVersion.ID, err)
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
		return nil, errors.Wrapf(err, "Failed to create new project version for pull request %v", pullRequestName)
	}
	attributes, err := sys.GetProjectVersionAttributesByProjectVersionID(masterProjectVersion.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to load project version attributes for master project version %v", masterProjectVersion.ID)
	}
	for _, attribute := range attributes {
		attribute.ID = 0
	}
	_, err = sys.SetProjectVersionAttributesByProjectVersionID(projectVersion.ID, attributes)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to update project version attributes for pull request project version %v", projectVersion.ID)
	}
	err = sys.ProjectVersionCopyFromPartial(masterProjectVersion.ID, projectVersion.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to copy from partial of project version %v to %v", masterProjectVersion.ID, projectVersion.ID)
	}
	_, err = sys.CommitProjectVersion(projectVersion.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to commit project version %v: %v", projectVersion.ID, err)
	}
	err = sys.ProjectVersionCopyCurrentState(masterProjectVersion.ID, projectVersion.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to copy current state of project version %v to %v", masterProjectVersion.ID, projectVersion.ID)
	}
	err = sys.ProjectVersionCopyPermissions(masterProjectVersion.ID, projectVersion.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to copy permissions of project version %v to %v", masterProjectVersion.ID, projectVersion.ID)
	}
	return projectVersion, nil
}

// GetProjectVersionAttributesByProjectVersionID returns the project version attributes of the project version identified by the id
func (sys *SystemInstance) GetProjectVersionAttributesByProjectVersionID(id int64) ([]*models.Attribute, error) {
	params := &attribute_of_project_version_controller.ListAttributeOfProjectVersionParams{ParentID: id}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.AttributeOfProjectVersionController.ListAttributeOfProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
}

// SetProjectVersionAttributesByProjectVersionID sets the project version attributes of the project version identified by the id
func (sys *SystemInstance) SetProjectVersionAttributesByProjectVersionID(id int64, attributes []*models.Attribute) ([]*models.Attribute, error) {
	params := &attribute_of_project_version_controller.UpdateCollectionAttributeOfProjectVersionParams{ParentID: id, Data: attributes}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.AttributeOfProjectVersionController.UpdateCollectionAttributeOfProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
}

// CreateProjectVersion creates the project version with the provided details
func (sys *SystemInstance) CreateProjectVersion(version *models.ProjectVersion) (*models.ProjectVersion, error) {
	params := &project_version_controller.CreateProjectVersionParams{Resource: version}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.ProjectVersionController.CreateProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
}

// ProjectVersionCopyFromPartial copies parts of the source project version to the target project version identified by their ids
func (sys *SystemInstance) ProjectVersionCopyFromPartial(sourceID, targetID int64) error {
	enable := true
	settings := models.ProjectVersionCopyPartialRequest{
		ProjectVersionID:            &targetID,
		PreviousProjectVersionID:    &sourceID,
		CopyAnalysisProcessingRules: &enable,
		CopyBugTrackerConfiguration: &enable,
		CopyCurrentStateFpr:         &enable,
		CopyCustomTags:              &enable,
	}
	params := &project_version_controller.CopyProjectVersionParams{Resource: &settings}
	params.WithTimeout(sys.timeout)
	_, err := sys.client.ProjectVersionController.CopyProjectVersion(params, sys)
	if err != nil {
		return err
	}
	return nil
}

// ProjectVersionCopyCurrentState copies the project version state of sourceID into the new project version addressed by targetID
func (sys *SystemInstance) ProjectVersionCopyCurrentState(sourceID, targetID int64) error {
	enable := true
	settings := models.ProjectVersionCopyCurrentStateRequest{
		ProjectVersionID:         &targetID,
		PreviousProjectVersionID: &sourceID,
		CopyCurrentStateFpr:      &enable,
	}
	params := &project_version_controller.CopyCurrentStateForProjectVersionParams{Resource: &settings}
	params.WithTimeout(sys.timeout)
	_, err := sys.client.ProjectVersionController.CopyCurrentStateForProjectVersion(params, sys)
	if err != nil {
		return err
	}
	return nil
}

func (sys *SystemInstance) getAuthEntityOfProjectVersion(id int64) ([]*models.AuthenticationEntity, error) {
	embed := "roles"
	params := &auth_entity_of_project_version_controller.ListAuthEntityOfProjectVersionParams{Embed: &embed, ParentID: id}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.AuthEntityOfProjectVersionController.ListAuthEntityOfProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
}

func (sys *SystemInstance) updateCollectionAuthEntityOfProjectVersion(id int64, data []*models.AuthenticationEntity) error {
	params := &auth_entity_of_project_version_controller.UpdateCollectionAuthEntityOfProjectVersionParams{ParentID: id, Data: data}
	params.WithTimeout(sys.timeout)
	_, err := sys.client.AuthEntityOfProjectVersionController.UpdateCollectionAuthEntityOfProjectVersion(params, sys)
	if err != nil {
		return err
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
	params := &project_version_controller.UpdateProjectVersionParams{ID: id, Resource: details}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.ProjectVersionController.UpdateProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
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
	scans := "scans"
	params := &artifact_of_project_version_controller.ListArtifactOfProjectVersionParams{ParentID: id, Embed: &scans}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.ArtifactOfProjectVersionController.ListArtifactOfProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
}

// MergeProjectVersionStateOfPRIntoMaster merges the PR project version's fpr result file into the master project version
func (sys *SystemInstance) MergeProjectVersionStateOfPRIntoMaster(downloadEndpoint, uploadEndpoint string, masterProjectID, masterProjectVersionID int64, pullRequestName string) error {
	log.Entry().Debugf("Looking up project version with name '%v' to merge audit status into master version", pullRequestName)
	prProjectVersion, _ := sys.GetProjectVersionDetailsByProjectIDAndVersionName(masterProjectID, pullRequestName, false, "")
	if nil != prProjectVersion {
		log.Entry().Debugf("Found project version with ID '%v', starting transfer", prProjectVersion.ID)
		data, err := sys.DownloadResultFile(downloadEndpoint, prProjectVersion.ID)
		if err != nil {
			return errors.Wrapf(err, "Failed to download current state FPR of PR project version %v", prProjectVersion.ID)
		}
		err = sys.uploadResultFileContent(uploadEndpoint, "prMergeTransfer.fpr", bytes.NewReader(data), masterProjectVersionID)
		if err != nil {
			return errors.Wrapf(err, "Failed to upload PR project version state to master project version %v", masterProjectVersionID)
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
	params := &filter_set_of_project_version_controller.ListFilterSetOfProjectVersionParams{ParentID: id}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.FilterSetOfProjectVersionController.ListFilterSetOfProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	var defaultFilterSet *models.FilterSet
	for _, filterSet := range result.GetPayload().Data {
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
	params := &issue_selector_set_of_project_version_controller.GetIssueSelectorSetOfProjectVersionParams{ParentID: id}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.IssueSelectorSetOfProjectVersionController.GetIssueSelectorSetOfProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	return sys.ReduceIssueFilterSelectorSet(result.GetPayload().Data, names, options), nil
}

// ReduceIssueFilterSelectorSet filters the set to the relevant filter display names
func (sys *SystemInstance) ReduceIssueFilterSelectorSet(issueFilterSelectorSet *models.IssueFilterSelectorSet, names []string, options []string) *models.IssueFilterSelectorSet {
	groupingList := []*models.IssueSelector{}
	if issueFilterSelectorSet.GroupBySet != nil {
		for _, group := range issueFilterSelectorSet.GroupBySet {
			if piperutils.ContainsString(names, *group.DisplayName) {
				log.Entry().Debugf("adding new grouping '%v' to reduced list", *group.DisplayName)
				groupingList = append(groupingList, group)
			}
		}
	}
	filterList := []*models.IssueFilterSelector{}
	if issueFilterSelectorSet.FilterBySet != nil {
		for _, filter := range issueFilterSelectorSet.FilterBySet {
			if piperutils.ContainsString(names, filter.DisplayName) {
				newFilter := &models.IssueFilterSelector{}
				newFilter.DisplayName = filter.DisplayName
				newFilter.Description = filter.Description
				newFilter.EntityType = filter.EntityType
				newFilter.FilterSelectorType = filter.FilterSelectorType
				newFilter.GUID = filter.GUID
				newFilter.Value = filter.Value
				newFilter.SelectorOptions = []*models.SelectorOption{}
				for _, option := range filter.SelectorOptions {
					if (nil != options && piperutils.ContainsString(options, option.DisplayName)) || options == nil || len(options) == 0 {
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

//GetFilterSetByDisplayName returns the set identified by the provided name or nil
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
	enable := true
	params := &issue_group_of_project_version_controller.ListIssueGroupOfProjectVersionParams{ParentID: id, Showsuppressed: &enable, Filterset: &filterset, Groupingtype: &groupingtype}
	params.WithTimeout(sys.timeout)
	if len(filter) > 0 {
		params.WithFilter(&filter)
	}
	result, err := sys.client.IssueGroupOfProjectVersionController.ListIssueGroupOfProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
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
	params := &issue_statistics_of_project_version_controller.ListIssueStatisticsOfProjectVersionParams{ParentID: id}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.IssueStatisticsOfProjectVersionController.ListIssueStatisticsOfProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
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
	params := &saved_report_controller.CreateSavedReportParams{Resource: &report}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.SavedReportController.CreateSavedReport(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
}

// GetReportDetails returns the details of the report addressed with id
func (sys *SystemInstance) GetReportDetails(id int64) (*models.SavedReport, error) {
	params := &saved_report_controller.ReadSavedReportParams{ID: id}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.SavedReportController.ReadSavedReport(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
}

func (sys *SystemInstance) invalidateFileTokens() error {
	log.Entry().Debug("invalidating file tokens")
	params := &file_token_controller.MultiDeleteFileTokenParams{}
	params.WithTimeout(sys.timeout)
	_, err := sys.client.FileTokenController.MultiDeleteFileToken(params, sys)
	return err
}

func (sys *SystemInstance) getFileToken(tokenType string) (*models.FileToken, error) {
	log.Entry().Debugf("fetching file token of type %v", tokenType)
	token := models.FileToken{FileTokenType: &tokenType}
	params := &file_token_controller.CreateFileTokenParams{Resource: &token}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.FileTokenController.CreateFileToken(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
}

// UploadResultFile uploads a fpr file to the fortify backend
func (sys *SystemInstance) UploadResultFile(endpoint, file string, projectVersionID int64) error {
	fileHandle, err := os.Open(file)
	if err != nil {
		return errors.Wrapf(err, "Unable to locate file %v", file)
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
		return nil, errors.Wrap(err, "Error fetching file token")
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
	data, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "Error reading the response data")
	}
	return data, nil
}

// DownloadReportFile downloads a report file from Fortify backend
func (sys *SystemInstance) DownloadReportFile(endpoint string, reportID int64) ([]byte, error) {
	data, err := sys.downloadFile(endpoint, http.MethodGet, "application/octet-stream", "REPORT_FILE", reportID)
	if err != nil {
		return nil, errors.Wrap(err, "Error downloading report file")
	}
	return data, nil
}

// DownloadResultFile downloads a result file from Fortify backend
func (sys *SystemInstance) DownloadResultFile(endpoint string, projectVersionID int64) ([]byte, error) {
	data, err := sys.downloadFile(endpoint, http.MethodGet, "application/zip", "DOWNLOAD", projectVersionID)
	if err != nil {
		return nil, errors.Wrap(err, "Error downloading result file")
	}
	return data, nil
}
