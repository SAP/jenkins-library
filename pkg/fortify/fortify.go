package fortify

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	ff "github.com/piper-validation/fortify-client-go/fortify"
	"github.com/piper-validation/fortify-client-go/fortify/attribute_of_project_version_controller"
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
func NewSystemInstance(serverURL, endpoint, authToken string, requestTimeout time.Duration) *SystemInstance {
	parts := strings.Split(serverURL, "://")
	format := strfmt.Default
	dateTimeFormat := models.Iso8601MilliDateTime{}
	format.Add("datetime", &dateTimeFormat, models.IsDateTime)

	sys := &SystemInstance{
		timeout: requestTimeout,
		token:   authToken,
		client: ff.NewHTTPClientWithConfig(format, &ff.TransportConfig{
			Host:     parts[1],
			Schemes:  []string{parts[0]},
			BasePath: endpoint},
		),
		logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/fortify"),
	}

	return sys
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
	if result.GetPayload().ResponseCode != 200 {
		return nil, fmt.Errorf("Backend returned HTTP response code %v", result.GetPayload().ResponseCode)
	}
	if result.GetPayload().ErrorCode != 0 {
		return nil, fmt.Errorf("Backend returned error code %v with message %v", result.GetPayload().ErrorCode, result.GetPayload().Message)
	}
	if len(result.GetPayload().Data) == 0 {
		return nil, fmt.Errorf("Project with name %v not found in backend", name)
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
	if result.GetPayload().ResponseCode != 200 {
		return nil, fmt.Errorf("Backend returned HTTP response code %v", result.GetPayload().ResponseCode)
	}
	if result.GetPayload().ErrorCode != 0 {
		return nil, fmt.Errorf("Backend returned error code %v with message %v", result.GetPayload().ErrorCode, result.GetPayload().Message)
	}
	if len(result.GetPayload().Data) == 0 {
		return nil, fmt.Errorf("Project version with name %v not found in for project with ID %v", name, id)
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
	if result.GetPayload().ResponseCode != 200 {
		return nil, fmt.Errorf("Backend returned HTTP response code %v", result.GetPayload().ResponseCode)
	}
	if result.GetPayload().ErrorCode != 0 {
		return nil, fmt.Errorf("Backend returned error code %v with message %v", result.GetPayload().ErrorCode, result.GetPayload().Message)
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
	if result.GetPayload().ResponseCode != 200 {
		return nil, fmt.Errorf("Backend returned HTTP response code %v", result.GetPayload().ResponseCode)
	}
	if result.GetPayload().ErrorCode != 0 {
		return nil, fmt.Errorf("Backend returned error code %v with message %v", result.GetPayload().ErrorCode, result.GetPayload().Message)
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
	result, err := sys.client.ProjectVersionController.CopyProjectVersion(params, sys)
	if err != nil {
		return err
	}
	if result.GetPayload().ResponseCode != 200 {
		return fmt.Errorf("Backend returned HTTP response code %v", result.GetPayload().ResponseCode)
	}
	if result.GetPayload().ErrorCode != 0 {
		return fmt.Errorf("Backend returned error code %v with message %v", result.GetPayload().ErrorCode, result.GetPayload().Message)
	}
	return nil
}
