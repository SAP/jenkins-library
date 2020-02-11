package fortify

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	ff "github.com/piper-validation/fortify-client-go/fortify"
	"github.com/piper-validation/fortify-client-go/fortify/attribute_of_project_version_controller"
	"github.com/piper-validation/fortify-client-go/fortify/auth_entity_of_project_version_controller"
	"github.com/piper-validation/fortify-client-go/fortify/project_controller"
	"github.com/piper-validation/fortify-client-go/fortify/project_version_controller"
	"github.com/piper-validation/fortify-client-go/fortify/project_version_of_project_controller"
	"github.com/piper-validation/fortify-client-go/models"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/sirupsen/logrus"
)

// System is the interface abstraction of a specific SystemInstance
type System interface {
}

// SystemInstance is the specific instance
type SystemInstance struct {
	timeout time.Duration
	token   string
	client  *ff.Fortify
	logger  *logrus.Entry
}

// NewSystemInstance - creates an returns a new SystemInstance
func NewSystemInstance(serverURL, endpoint, authToken string, timeout time.Duration) *SystemInstance {
	schemeHost := strings.Split(serverURL, "://")
	hostEndpoint := strings.Split(schemeHost[1], "/")
	format := strfmt.Default
	dateTimeFormat := models.Iso8601MilliDateTime{}
	format.Add("datetime", &dateTimeFormat, models.IsDateTime)
	clientInstance := ff.NewHTTPClientWithConfig(format, &ff.TransportConfig{
		Host:     hostEndpoint[0],
		Schemes:  []string{schemeHost[0]},
		BasePath: fmt.Sprintf("%v/%v", hostEndpoint[1], endpoint)},
	)

	return NewSystemInstanceForClient(clientInstance, authToken, timeout)
}

// NewSystemInstanceForClient - creates a new SystemInstance
func NewSystemInstanceForClient(clientInstance *ff.Fortify, authToken string, requestTimeout time.Duration) *SystemInstance {
	return &SystemInstance{
		timeout: requestTimeout,
		token:   authToken,
		client:  clientInstance,
		logger:  log.Entry().WithField("package", "SAP/jenkins-library/pkg/fortify"),
	}
}

// AuthenticateRequest authenticates the request
func (sys *SystemInstance) AuthenticateRequest(req runtime.ClientRequest, formats strfmt.Registry) error {
	req.SetHeaderParam("Authorization", fmt.Sprintf("FortifyToken %v", sys.token))
	return nil
}

// GetProjectByName returns the project identified by the name provided
func (sys *SystemInstance) GetProjectByName(name string) (*models.Project, error) {
	nameParam := fmt.Sprintf("name=%v", name)
	fullText := true
	params := &project_controller.ListProjectParams{Q: &nameParam, Fulltextsearch: &fullText}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.ProjectController.ListProject(params, sys)
	if err != nil {
		return nil, err
	}
	for _, project := range result.GetPayload().Data {
		if *project.Name == name {
			return project, nil
		}
	}
	return nil, fmt.Errorf("Project with name %v not found in backend", name)
}

//GetProjectVersionDetailsByNameAndProjectID returns the project version details of the project version identified by the id
func (sys *SystemInstance) GetProjectVersionDetailsByNameAndProjectID(id int64, name string) (*models.ProjectVersion, error) {
	nameParam := fmt.Sprintf("name=%v", name)
	fullText := true
	params := &project_version_of_project_controller.ListProjectVersionOfProjectParams{ParentID: id, Q: &nameParam, Fulltextsearch: &fullText}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.ProjectVersionOfProjectController.ListProjectVersionOfProject(params, sys)
	if err != nil {
		return nil, err
	}
	for _, projectVersion := range result.GetPayload().Data {
		if *projectVersion.Name == name {
			return projectVersion, nil
		}
	}
	return nil, fmt.Errorf("Project version with name %v not found in for project with ID %v", name, id)
}

//GetProjectVersionAttributesByID returns the project version attributes of the project version identified by the id
func (sys *SystemInstance) GetProjectVersionAttributesByID(id int64) ([]*models.Attribute, error) {
	params := &attribute_of_project_version_controller.ListAttributeOfProjectVersionParams{ParentID: id}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.AttributeOfProjectVersionController.ListAttributeOfProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
}

//CreateProjectVersion creates the project version with the provided details
func (sys *SystemInstance) CreateProjectVersion(version *models.ProjectVersion) (*models.ProjectVersion, error) {
	params := &project_version_controller.CreateProjectVersionParams{Resource: version}
	params.WithTimeout(sys.timeout)
	result, err := sys.client.ProjectVersionController.CreateProjectVersion(params, sys)
	if err != nil {
		return nil, err
	}
	return result.GetPayload().Data, nil
}

//ProjectVersionCopyFromPartial copies parts of the source project version to the target project version identified by their ids
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

//ProjectVersionCopyCurrentState copies the project version state of sourceID into the new project version addressed by targetID
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

//CopyProjectVersionPermissions copies the authentication entity of the project version addressed by sourceID to the one of targetID
func (sys *SystemInstance) CopyProjectVersionPermissions(sourceID, targetID int64) error {
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

//CommitProjectVersion commits the project version with the provided id
func (sys *SystemInstance) CommitProjectVersion(id int64) (*models.ProjectVersion, error) {
	enabled := true
	update := models.ProjectVersion{Committed: &enabled}
	return sys.updateProjectVersionDetails(id, &update)
}

//InactivateProjectVersion inactivates the project version with the provided id
func (sys *SystemInstance) InactivateProjectVersion(id int64) (*models.ProjectVersion, error) {
	enabled := true
	disabled := false
	update := models.ProjectVersion{Committed: &enabled, Active: &disabled}
	return sys.updateProjectVersionDetails(id, &update)
}
