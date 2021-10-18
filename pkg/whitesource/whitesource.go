package whitesource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// Product defines a WhiteSource product with name and token
type Product struct {
	Name           string `json:"name"`
	Token          string `json:"token"`
	CreationDate   string `json:"creationDate,omitempty"`
	LastUpdateDate string `json:"lastUpdatedDate,omitempty"`
}

// Assignment describes a list of UserAssignments and GroupAssignments which can be attributed to a WhiteSource Product.
type Assignment struct {
	UserAssignments  []UserAssignment  `json:"userAssignments,omitempty"`
	GroupAssignments []GroupAssignment `json:"groupAssignments,omitempty"`
}

// UserAssignment holds an email address for a WhiteSource user
// which can be assigned to a WhiteSource Product in a specific role.
type UserAssignment struct {
	Email string `json:"email,omitempty"`
}

// GroupAssignment refers to the name of a particular group in WhiteSource.
type GroupAssignment struct {
	Name string `json:"name,omitempty"`
}

// Alert
type Alert struct {
	Vulnerability Vulnerability `json:"vulnerability"`
	Library       Library       `json:"library,omitempty"`
	Project       string        `json:"project,omitempty"`
	CreationDate  string        `json:"creation_date,omitempty"`
}

// Library
type Library struct {
	Name     string `json:"name,omitempty"`
	Filename string `json:"filename,omitempty"`
	Version  string `json:"version,omitempty"`
	Project  string `json:"project,omitempty"`
}

// Vulnerability
type Vulnerability struct {
	Name              string  `json:"name,omitempty"`
	Type              string  `json:"type,omitempty"`
	Level             string  `json:"level,omitempty"`
	Description       string  `json:"description,omitempty"`
	Severity          string  `json:"severity,omitempty"`
	CVSS3Severity     string  `json:"cvss3_severity,omitempty"`
	CVSS3Score        float64 `json:"cvss3_score,omitempty"`
	Score             float64 `json:"score,omitempty"`
	FixResolutionText string  `json:"fixResolutionText,omitempty"`
	PublishDate       string  `json:"publishDate,omitempty"`
}

// Project defines a WhiteSource project with name and token
type Project struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	PluginName     string `json:"pluginName"`
	Token          string `json:"token"`
	UploadedBy     string `json:"uploadedBy"`
	CreationDate   string `json:"creationDate,omitempty"`
	LastUpdateDate string `json:"lastUpdatedDate,omitempty"`
}

// Request defines a request object to be sent to the WhiteSource system
type Request struct {
	RequestType          string      `json:"requestType,omitempty"`
	UserKey              string      `json:"userKey,omitempty"`
	ProductToken         string      `json:"productToken,omitempty"`
	ProductName          string      `json:"productName,omitempty"`
	ProjectToken         string      `json:"projectToken,omitempty"`
	OrgToken             string      `json:"orgToken,omitempty"`
	Format               string      `json:"format,omitempty"`
	ProductAdmins        *Assignment `json:"productAdmins,omitempty"`
	ProductMembership    *Assignment `json:"productMembership,omitempty"`
	AlertsEmailReceivers *Assignment `json:"alertsEmailReceivers,omitempty"`
	ProductApprovers     *Assignment `json:"productApprovers,omitempty"`
	ProductIntegrators   *Assignment `json:"productIntegrators,omitempty"`
}

// System defines a WhiteSource System including respective tokens (e.g. org token, user token)
type System struct {
	httpClient piperhttp.Sender
	orgToken   string
	serverURL  string
	userToken  string
	maxRetries int
	retryInterval time.Duration
}

// DateTimeLayout is the layout of the time format used by the WhiteSource API.
const DateTimeLayout = "2006-01-02 15:04:05 -0700"

// NewSystem constructs a new System instance
func NewSystem(serverURL, orgToken, userToken string, timeout time.Duration) *System {
	httpClient := &piperhttp.Client{}
	httpClient.SetOptions(piperhttp.ClientOptions{TransportTimeout: timeout})
	return &System{
		serverURL:  serverURL,
		orgToken:   orgToken,
		userToken:  userToken,
		httpClient: httpClient,
		maxRetries: 10,
		retryInterval: 3 * time.Second,
	}
}

// GetProductsMetaInfo retrieves meta information for all WhiteSource products a user has access to
func (s *System) GetProductsMetaInfo() ([]Product, error) {
	wsResponse := struct {
		ProductVitals []Product `json:"productVitals"`
	}{
		ProductVitals: []Product{},
	}

	req := Request{
		RequestType: "getOrganizationProductVitals",
	}

	err := s.sendRequestAndDecodeJSON(req, &wsResponse)
	if err != nil {
		return wsResponse.ProductVitals, errors.Wrap(err, "WhiteSource request failed")
	}

	return wsResponse.ProductVitals, nil
}

// GetProductByName retrieves meta information for a specific WhiteSource product
func (s *System) GetProductByName(productName string) (Product, error) {
	products, err := s.GetProductsMetaInfo()
	if err != nil {
		return Product{}, errors.Wrap(err, "failed to retrieve WhiteSource products")
	}

	for _, p := range products {
		if p.Name == productName {
			return p, nil
		}
	}

	return Product{}, fmt.Errorf("product '%v' not found in WhiteSource", productName)
}

// CreateProduct creates a new WhiteSource product and returns its product token.
func (s *System) CreateProduct(productName string) (string, error) {
	wsResponse := struct {
		ProductToken string `json:"productToken"`
	}{
		ProductToken: "",
	}

	req := Request{
		RequestType: "createProduct",
		ProductName: productName,
	}

	err := s.sendRequestAndDecodeJSON(req, &wsResponse)
	if err != nil {
		return "", errors.Wrap(err, "WhiteSource request failed")
	}

	return wsResponse.ProductToken, nil
}

// SetProductAssignments assigns various types of membership to a WhiteSource Product.
func (s *System) SetProductAssignments(productToken string, membership, admins, alertReceivers *Assignment) error {
	req := Request{
		RequestType:          "setProductAssignments",
		ProductToken:         productToken,
		ProductMembership:    membership,
		ProductAdmins:        admins,
		AlertsEmailReceivers: alertReceivers,
	}

	err := s.sendRequestAndDecodeJSON(req, nil)
	if err != nil {
		return errors.Wrap(err, "WhiteSource request failed")
	}

	return nil
}

// GetProjectsMetaInfo retrieves the registered projects for a specific WhiteSource product
func (s *System) GetProjectsMetaInfo(productToken string) ([]Project, error) {
	wsResponse := struct {
		ProjectVitals []Project `json:"projectVitals"`
	}{
		ProjectVitals: []Project{},
	}

	req := Request{
		RequestType:  "getProductProjectVitals",
		ProductToken: productToken,
	}

	err := s.sendRequestAndDecodeJSON(req, &wsResponse)
	if err != nil {
		return nil, errors.Wrap(err, "WhiteSource request failed")
	}

	return wsResponse.ProjectVitals, nil
}

// GetProjectToken returns the project token for a project with a given name
func (s *System) GetProjectToken(productToken, projectName string) (string, error) {
	project, err := s.GetProjectByName(productToken, projectName)
	if err != nil {
		return "", err
	}
	return project.Token, nil
}

// GetProjectByToken returns project meta info given a project token
func (s *System) GetProjectByToken(projectToken string) (Project, error) {
	wsResponse := struct {
		ProjectVitals []Project `json:"projectVitals"`
	}{
		ProjectVitals: []Project{},
	}

	req := Request{
		RequestType:  "getProjectVitals",
		ProjectToken: projectToken,
	}

	err := s.sendRequestAndDecodeJSON(req, &wsResponse)
	if err != nil {
		return Project{}, errors.Wrap(err, "WhiteSource request failed")
	}

	if len(wsResponse.ProjectVitals) == 0 {
		return Project{}, errors.Wrapf(err, "no project with token '%s' found in WhiteSource", projectToken)
	}

	return wsResponse.ProjectVitals[0], nil
}

// GetProjectByName fetches all projects and returns the one matching the given projectName, or none, if not found
func (s *System) GetProjectByName(productToken, projectName string) (Project, error) {
	projects, err := s.GetProjectsMetaInfo(productToken)
	if err != nil {
		return Project{}, errors.Wrap(err, "failed to retrieve WhiteSource project meta info")
	}

	for _, project := range projects {
		if projectName == project.Name {
			return project, nil
		}
	}

	// returns empty project and no error. The reason seems to be that it makes polling until the project exists easier.
	return Project{}, nil
}

// GetProjectsByIDs retrieves all projects for the given productToken and filters them by the given project ids
func (s *System) GetProjectsByIDs(productToken string, projectIDs []int64) ([]Project, error) {
	projects, err := s.GetProjectsMetaInfo(productToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve WhiteSource project meta info")
	}

	var projectsMatched []Project
	for _, project := range projects {
		for _, projectID := range projectIDs {
			if projectID == project.ID {
				projectsMatched = append(projectsMatched, project)
				break
			}
		}
	}

	return projectsMatched, nil
}

// GetProjectTokens returns the project tokens matching a given a slice of project names
func (s *System) GetProjectTokens(productToken string, projectNames []string) ([]string, error) {
	projectTokens := []string{}
	projects, err := s.GetProjectsMetaInfo(productToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve WhiteSource project meta info")
	}

	for _, project := range projects {
		for _, projectName := range projectNames {
			if projectName == project.Name {
				projectTokens = append(projectTokens, project.Token)
			}
		}
	}
	return projectTokens, nil
}

// GetProductName returns the product name for a given product token
func (s *System) GetProductName(productToken string) (string, error) {
	wsResponse := struct {
		ProductTags []Product `json:"productTags"`
	}{
		ProductTags: []Product{},
	}

	req := Request{
		RequestType:  "getProductTags",
		ProductToken: productToken,
	}

	err := s.sendRequestAndDecodeJSON(req, &wsResponse)
	if err != nil {
		return "", errors.Wrap(err, "WhiteSource request failed")
	}

	if len(wsResponse.ProductTags) == 0 {
		return "", nil // fmt.Errorf("no product with token '%s' found in WhiteSource", productToken)
	}

	return wsResponse.ProductTags[0].Name, nil
}

// GetProjectRiskReport
func (s *System) GetProjectRiskReport(projectToken string) ([]byte, error) {
	req := Request{
		RequestType:  "getProjectRiskReport",
		ProjectToken: projectToken,
	}

	respBody, err := s.sendRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "WhiteSource getProjectRiskReport request failed")
	}

	return respBody, nil
}

// GetProjectVulnerabilityReport
func (s *System) GetProjectVulnerabilityReport(projectToken string, format string) ([]byte, error) {
	req := Request{
		RequestType:  "getProjectVulnerabilityReport",
		ProjectToken: projectToken,
		Format:       format,
	}

	respBody, err := s.sendRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "WhiteSource getProjectVulnerabilityReport request failed")
	}

	return respBody, nil
}

// GetProjectAlerts
func (s *System) GetProjectAlerts(projectToken string) ([]Alert, error) {
	wsResponse := struct {
		Alerts []Alert `json:"alerts"`
	}{
		Alerts: []Alert{},
	}

	req := Request{
		RequestType:  "getProjectAlerts",
		ProjectToken: projectToken,
	}

	err := s.sendRequestAndDecodeJSON(req, &wsResponse)
	if err != nil {
		return nil, errors.Wrap(err, "WhiteSource request failed")
	}

	return wsResponse.Alerts, nil
}

// GetProjectLibraryLocations
func (s *System) GetProjectLibraryLocations(projectToken string) ([]Library, error) {
	wsResponse := struct {
		Libraries []Library `json:"libraryLocations"`
	}{
		Libraries: []Library{},
	}

	req := Request{
		RequestType:  "getProjectLibraryLocations",
		ProjectToken: projectToken,
	}

	err := s.sendRequestAndDecodeJSON(req, &wsResponse)
	if err != nil {
		return nil, errors.Wrap(err, "WhiteSource request failed")
	}

	return wsResponse.Libraries, nil
}

func (s *System) sendRequestAndDecodeJSON(req Request, result interface{}) error {
	var count int
	return s.sendRequestAndDecodeJSONRecursive(req, result, &count)
}

func (s *System) sendRequestAndDecodeJSONRecursive(req Request, result interface{}, count *int) error {
	respBody, err := s.sendRequest(req)
	if err != nil {
		return errors.Wrap(err, "WhiteSource request failed")
	}

	log.Entry().Debugf("response: %v", string(respBody))

	errorResponse := struct {
		ErrorCode    int    `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}{}

	err = json.Unmarshal(respBody, &errorResponse)
	if err == nil && errorResponse.ErrorCode != 0 {
		if *count < s.maxRetries && errorResponse.ErrorCode == 3000 {
			var initial bool
			if *count == 0 {
				initial = true
			}
			log.Entry().Warnf("backend returned error 3000, retrying in %v", s.retryInterval)
			time.Sleep(s.retryInterval)
			*count = *count + 1
			err = s.sendRequestAndDecodeJSONRecursive(req, result, count)
			if err != nil {
				if initial {
					return errors.Wrapf(err, "%v retries failed", s.maxRetries)
				}
				return err
			}
		}
		return fmt.Errorf("invalid request, error code %v, message '%s'",
			errorResponse.ErrorCode, errorResponse.ErrorMessage)
	}

	if result != nil {
		err = json.Unmarshal(respBody, result)
		if err != nil {
			return errors.Wrap(err, "failed to parse WhiteSource response")
		}
	}
	return nil
}

func (s *System) sendRequest(req Request) ([]byte, error) {
	var responseBody []byte
	if req.UserKey == "" {
		req.UserKey = s.userToken
	}
	if req.OrgToken == "" {
		req.OrgToken = s.orgToken
	}

	body, err := json.Marshal(req)
	if err != nil {
		return responseBody, errors.Wrap(err, "failed to create WhiteSource request")
	}

	log.Entry().Debugf("request: %v", string(body))

	headers := http.Header{}
	headers.Add("Content-Type", "application/json")
	response, err := s.httpClient.SendRequest(http.MethodPost, s.serverURL, bytes.NewBuffer(body), headers, nil)

	if err != nil {
		return responseBody, errors.Wrap(err, "failed to send request to WhiteSource")
	}

	responseBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return responseBody, errors.Wrap(err, "failed to read WhiteSource response")
	}

	return responseBody, nil
}
