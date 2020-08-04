package whitesource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
	RequestType  string `json:"requestType,omitempty"`
	UserKey      string `json:"userKey,omitempty"`
	ProductToken string `json:"productToken,omitempty"`
	ProductName  string `json:"productName,omitempty"`
	ProjectToken string `json:"projectToken,omitempty"`
	OrgToken     string `json:"orgToken,omitempty"`
	Format       string `json:"format,omitempty"`
}

// system defines a WhiteSource system including respective tokens (e.g. org token, user token)
type system struct {
	httpClient piperhttp.Sender
	orgToken   string
	serverURL  string
	userToken  string
}

// NewSystem constructs a new system instance
func NewSystem(serverURL, orgToken, userToken string) System {
	return &system{
		serverURL:  serverURL,
		orgToken:   orgToken,
		userToken:  userToken,
		httpClient: &piperhttp.Client{},
	}
}

// GetProductsMetaInfo retrieves meta information for all WhiteSource products a user has access to
func (s *system) GetProductsMetaInfo() ([]Product, error) {
	wsResponse := struct {
		ProductVitals []Product `json:"productVitals"`
	}{
		ProductVitals: []Product{},
	}

	req := Request{
		RequestType: "getOrganizationProductVitals",
	}

	respBody, err := s.sendRequest(req)
	if err != nil {
		return wsResponse.ProductVitals, errors.Wrap(err, "WhiteSource request failed")
	}

	err = json.Unmarshal(respBody, &wsResponse)
	if err != nil {
		return wsResponse.ProductVitals, errors.Wrap(err, "failed to parse WhiteSource response")
	}

	return wsResponse.ProductVitals, nil
}

// GetProductByName retrieves meta information for a specific WhiteSource product
func (s *system) GetProductByName(productName string) (Product, error) {
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

// GetProjectsMetaInfo retrieves meta information for a specific WhiteSource product
func (s *system) GetProjectsMetaInfo(productToken string) ([]Project, error) {
	wsResponse := struct {
		ProjectVitals []Project `json:"projectVitals"`
	}{
		ProjectVitals: []Project{},
	}

	req := Request{
		RequestType:  "getProductProjectVitals",
		ProductToken: productToken,
	}

	respBody, err := s.sendRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "WhiteSource request failed")
	}

	err = json.Unmarshal(respBody, &wsResponse)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse WhiteSource response")
	}

	return wsResponse.ProjectVitals, nil
}

// GetProjectToken returns the project token for a project with a given name
func (s *system) GetProjectToken(productToken, projectName string) (string, error) {
	var token string
	project, err := s.GetProjectByName(productToken, projectName)
	if err != nil {
		return "", err
	}

	// returns a nil token and no error if not found
	if project != nil {
		token = project.Token
	}

	return token, nil
}

// GetProjectVitals returns project meta info given a project token
func (s *system) GetProjectVitals(projectToken string) (*Project, error) {
	wsResponse := struct {
		ProjectVitals []Project `json:"projectVitals"`
	}{
		ProjectVitals: []Project{},
	}

	req := Request{
		RequestType:  "getProjectVitals",
		ProjectToken: projectToken,
	}

	respBody, err := s.sendRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "WhiteSource request failed")
	}

	err = json.Unmarshal(respBody, &wsResponse)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse WhiteSource response")
	}

	return &wsResponse.ProjectVitals[0], nil
}

// GetProjectByName returns the finds and returns a project by name
func (s *system) GetProjectByName(productToken, projectName string) (*Project, error) {
	var project *Project
	projects, err := s.GetProjectsMetaInfo(productToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve WhiteSource project meta info")
	}

	for _, proj := range projects {
		if projectName == proj.Name {
			project = &proj
			break
		}
	}

	// returns a nil project and no error if no project exists with projectName
	return project, nil
}

// GetProjectTokens returns the project tokens matching a given a slice of project names
func (s *system) GetProjectTokens(productToken string, projectNames []string) ([]string, error) {
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
func (s *system) GetProductName(productToken string) (string, error) {
	wsResponse := struct {
		ProductTags []Product `json:"productTags"`
	}{
		ProductTags: []Product{},
	}

	req := Request{
		RequestType:  "getProductTags",
		ProductToken: productToken,
	}

	respBody, err := s.sendRequest(req)
	if err != nil {
		return "", errors.Wrap(err, "WhiteSource request failed")
	}

	err = json.Unmarshal(respBody, &wsResponse)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse WhiteSource response")
	}

	if len(wsResponse.ProductTags) > 0 {
		return wsResponse.ProductTags[0].Name, nil
	}
	return "", nil
}

// GetProjectRiskReport
func (s *system) GetProjectRiskReport(projectToken string) ([]byte, error) {
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
func (s *system) GetProjectVulnerabilityReport(projectToken string, format string) ([]byte, error) {

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
func (s *system) GetProjectAlerts(projectToken string) ([]Alert, error) {
	wsResponse := struct {
		Alerts []Alert `json:"alerts"`
	}{
		Alerts: []Alert{},
	}

	req := Request{
		RequestType:  "getProjectAlerts",
		ProjectToken: projectToken,
	}

	respBody, err := s.sendRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "WhiteSource request failed")
	}

	err = json.Unmarshal(respBody, &wsResponse)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse WhiteSource response")
	}

	return wsResponse.Alerts, nil
}

// GetProjectLibraryLocations
func (s *system) GetProjectLibraryLocations(projectToken string) ([]Library, error) {
	wsResponse := struct {
		Libraries []Library `json:"libraryLocations"`
	}{
		Libraries: []Library{},
	}

	req := Request{
		RequestType:  "getProjectLibraryLocations",
		ProjectToken: projectToken,
	}

	respBody, err := s.sendRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "WhiteSource request failed")
	}

	err = json.Unmarshal(respBody, &wsResponse)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse WhiteSource response")
	}

	return wsResponse.Libraries, nil
}

func (s *system) sendRequest(req Request) ([]byte, error) {
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

	log.Entry().Debug(string(body))

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
