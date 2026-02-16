package whitesource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/format"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/package-url/packageurl-go"
)

// ReportsDirectory defines the subfolder for the WhiteSource reports which are generated
const ReportsDirectory = "whitesource"

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
	*format.Assessment
	Vulnerability    Vulnerability `json:"vulnerability"`
	Type             string        `json:"type,omitempty"`
	Level            string        `json:"level,omitempty"`
	Library          Library       `json:"library,omitempty"`
	Project          string        `json:"project,omitempty"`
	DirectDependency bool          `json:"directDependency,omitempty"`
	Description      string        `json:"description,omitempty"`
	CreationDate     string        `json:"date,omitempty"`
	ModifiedDate     string        `json:"modifiedDate,omitempty"`
	Status           string        `json:"status,omitempty"`
	Comments         string        `json:"comments,omitempty"`
}

// DependencyType returns type of dependency: direct/transitive
func (a *Alert) DependencyType() string {
	if a.DirectDependency == true {
		return "direct"
	}
	return "transitive"
}

// Title returns the issue title representation of the contents
func (a Alert) Title() string {
	if a.Type == "SECURITY_VULNERABILITY" {
		return fmt.Sprintf("Security Vulnerability %v %v", a.Vulnerability.Name, a.Library.ArtifactID)
	} else if a.Type == "REJECTED_BY_POLICY_RESOURCE" {
		return fmt.Sprintf("Policy Violation %v %v", a.Vulnerability.Name, a.Library.ArtifactID)
	}
	return fmt.Sprintf("%v %v %v ", a.Type, a.Vulnerability.Name, a.Library.ArtifactID)
}

func (a *Alert) ContainedIn(assessments *[]format.Assessment) (bool, error) {
	localPurl := a.Library.ToPackageUrl().ToString()
	for _, assessment := range *assessments {
		if assessment.Vulnerability == a.Vulnerability.Name {
			for _, purl := range assessment.Purls {
				assessmentPurl, err := purl.ToPackageUrl()
				assessmentPurlStr := assessmentPurl.ToString()
				if err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					log.Entry().WithError(err).Errorf("assessment from file ignored due to invalid packageUrl '%s'", purl)
					return false, err
				}
				if assessmentPurlStr == localPurl {
					log.Entry().Debugf("matching assessment %v on package %v detected for alert %v", assessment.Vulnerability, assessmentPurlStr, a.Vulnerability.Name)
					a.Assessment = &assessment
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func transformLibToPurlType(libType string) string {
	log.Entry().Debugf("LibType reported as %v", libType)
	switch strings.ToLower(libType) {
	case "java":
		fallthrough
	case "maven_artifact":
		return packageurl.TypeMaven
	case "javascript/node.js":
		fallthrough
	case "node_packaged_module":
		return packageurl.TypeNPM
	case "javascript/bower":
		return "bower"
	case "go":
		fallthrough
	case "go_package":
		return packageurl.TypeGolang
	case "python":
		fallthrough
	case "python_package":
		return packageurl.TypePyPi
	case "debian":
		fallthrough
	case "debian_package":
		return packageurl.TypeDebian
	case "docker":
		return packageurl.TypeDocker
	case ".net":
		fallthrough
	case "dot_net_resource":
		return packageurl.TypeNuget
	}
	return packageurl.TypeGeneric
}

func consolidate(cvss2severity, cvss3severity string, cvss2score, cvss3score float64) string {
	cvssseverity := consolidateSeverities(cvss2severity, cvss3severity)
	switch cvssseverity {
	case "low":
		return "LOW"
	case "medium":
		return "MEDIUM"
	case "high":
		if cvss3score >= 9 || cvss2score >= 9 {
			return "CRITICAL"
		}
		return "HIGH"
	}
	return "none"
}

// ToMarkdown returns the markdown representation of the contents
func (a Alert) ToMarkdown() ([]byte, error) {

	if a.Type == "SECURITY_VULNERABILITY" {
		score := consolidateScores(a.Vulnerability.Score, a.Vulnerability.CVSS3Score)

		vul := reporting.VulnerabilityReport{
			ArtifactID: a.Library.ArtifactID,
			// no information available about branch and commit, yet
			Branch:         "",
			CommitID:       "",
			Description:    a.Vulnerability.Description,
			DependencyType: a.DependencyType(),
			// no information available about footer, yet
			Footer: "",
			Group:  a.Library.GroupID,
			// no information available about pipeline name and link, yet
			PipelineName:      "",
			PipelineLink:      "",
			PublishDate:       a.Vulnerability.PublishDate,
			Resolution:        a.Vulnerability.TopFix.FixResolution,
			Score:             score,
			Severity:          consolidate(a.Vulnerability.Severity, a.Vulnerability.CVSS3Severity, a.Vulnerability.Score, a.Vulnerability.CVSS3Score),
			Version:           a.Library.Version,
			PackageURL:        a.Library.ToPackageUrl().ToString(),
			VulnerabilityLink: a.Vulnerability.URL,
			VulnerabilityName: a.Vulnerability.Name,
		}
		return vul.ToMarkdown()
	} else if a.Type == "REJECTED_BY_POLICY_RESOURCE" {
		policyReport := reporting.PolicyViolationReport{
			ArtifactID: a.Library.ArtifactID,
			// no information available about branch and commit, yet
			Branch:           "",
			CommitID:         "",
			Description:      a.Vulnerability.Description,
			DirectDependency: fmt.Sprint(a.DirectDependency),
			// no information available about footer, yet
			Footer: "",
			Group:  a.Library.GroupID,
			// no information available about pipeline name and link, yet
			PipelineName: "",
			PipelineLink: "",
			Version:      a.Library.Version,
			PackageURL:   a.Library.ToPackageUrl().ToString(),
		}
		return policyReport.ToMarkdown()
	}

	return []byte{}, nil
}

// ToTxt returns the textual representation of the contents
func (a Alert) ToTxt() string {
	score := consolidateScores(a.Vulnerability.Score, a.Vulnerability.CVSS3Score)
	return fmt.Sprintf(`Vulnerability %v
Severity: %v
Base (NVD) Score: %v
Package: %v
Installed Version: %v
Package URL: %v
Description: %v
Fix Resolution: %v
Link: [%v](%v)`,
		a.Vulnerability.Name,
		a.Vulnerability.Severity,
		score,
		a.Library.ArtifactID,
		a.Library.Version,
		a.Library.ToPackageUrl().ToString(),
		a.Vulnerability.Description,
		a.Vulnerability.TopFix.FixResolution,
		a.Vulnerability.Name,
		a.Vulnerability.URL,
	)
}

func consolidateScores(cvss2score, cvss3score float64) float64 {
	score := cvss3score
	if score == 0 {
		score = cvss2score
	}
	return score
}

// Library
type Library struct {
	KeyUUID      string    `json:"keyUuid,omitempty"`
	KeyID        int       `json:"keyId,omitempty"`
	Name         string    `json:"name,omitempty"`
	Filename     string    `json:"filename,omitempty"`
	ArtifactID   string    `json:"artifactId,omitempty"`
	GroupID      string    `json:"groupId,omitempty"`
	Version      string    `json:"version,omitempty"`
	Sha1         string    `json:"sha1,omitempty"`
	LibType      string    `json:"type,omitempty"`
	Coordinates  string    `json:"coordinates,omitempty"`
	Dependencies []Library `json:"dependencies,omitempty"`
}

// ToPackageUrl constructs and returns the package URL of the library
func (l Library) ToPackageUrl() *packageurl.PackageURL {
	return packageurl.NewPackageURL(transformLibToPurlType(l.LibType), l.GroupID, l.ArtifactID, l.Version, nil, "")
}

// Vulnerability defines a vulnerability as returned by WhiteSource
type Vulnerability struct {
	Name              string      `json:"name,omitempty"`
	Type              string      `json:"type,omitempty"`
	Severity          string      `json:"severity,omitempty"`
	Score             float64     `json:"score,omitempty"`
	CVSS3Severity     string      `json:"cvss3_severity,omitempty"`
	CVSS3Score        float64     `json:"cvss3_score,omitempty"`
	PublishDate       string      `json:"publishDate,omitempty"`
	URL               string      `json:"url,omitempty"`
	Description       string      `json:"description,omitempty"`
	TopFix            Fix         `json:"topFix,omitempty"`
	AllFixes          []Fix       `json:"allFixes,omitempty"`
	FixResolutionText string      `json:"fixResolutionText,omitempty"`
	References        []Reference `json:"references,omitempty"`
}

// Fix defines a Fix as returned by WhiteSource
type Fix struct {
	Vulnerability string `json:"vulnerability,omitempty"`
	Type          string `json:"type,omitempty"`
	Origin        string `json:"origin,omitempty"`
	URL           string `json:"url,omitempty"`
	FixResolution string `json:"fixResolution,omitempty"`
	Date          string `json:"date,omitempty"`
	Message       string `json:"message,omitempty"`
	ExtraData     string `json:"extraData,omitempty"`
}

// Reference defines a reference for the library affected
type Reference struct {
	URL                 string `json:"url,omitempty"`
	Homepage            string `json:"homepage,omitempty"`
	GenericPackageIndex string `json:"genericPackageIndex,omitempty"`
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
	AlertType            string      `json:"alertType,omitempty"`
	ProductAdmins        *Assignment `json:"productAdmins,omitempty"`
	ProductMembership    *Assignment `json:"productMembership,omitempty"`
	AlertsEmailReceivers *Assignment `json:"alertsEmailReceivers,omitempty"`
	ProductApprovers     *Assignment `json:"productApprovers,omitempty"`
	ProductIntegrators   *Assignment `json:"productIntegrators,omitempty"`
	IncludeInHouseData   bool        `json:"includeInHouseData,omitempty"`
}

// System defines a WhiteSource System including respective tokens (e.g. org token, user token)
type System struct {
	httpClient    piperhttp.Sender
	orgToken      string
	serverURL     string
	userToken     string
	maxRetries    int
	retryInterval time.Duration
}

// DateTimeLayout is the layout of the time format used by the WhiteSource API.
const DateTimeLayout = "2006-01-02 15:04:05 -0700"

// NewSystem constructs a new System instance
func NewSystem(serverURL, orgToken, userToken string, timeout time.Duration) *System {
	httpClient := &piperhttp.Client{}
	httpClient.SetOptions(piperhttp.ClientOptions{TransportTimeout: timeout})
	return &System{
		serverURL:     serverURL,
		orgToken:      orgToken,
		userToken:     userToken,
		httpClient:    httpClient,
		maxRetries:    10,
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
		return wsResponse.ProductVitals, err
	}

	return wsResponse.ProductVitals, nil
}

// GetProductByName retrieves meta information for a specific WhiteSource product
func (s *System) GetProductByName(productName string) (Product, error) {
	products, err := s.GetProductsMetaInfo()
	if err != nil {
		return Product{}, fmt.Errorf("failed to retrieve WhiteSource products: %w", err)
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
		return "", err
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
		return err
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
		return nil, err
	}

	return wsResponse.ProjectVitals, nil
}

// GetProjectHierarchy retrieves the full set of libraries that the project depends on
func (s *System) GetProjectHierarchy(projectToken string, includeInHouse bool) ([]Library, error) {
	wsResponse := struct {
		Libraries []Library `json:"libraries"`
	}{
		Libraries: []Library{},
	}

	req := Request{
		RequestType:        "getProjectHierarchy",
		ProjectToken:       projectToken,
		IncludeInHouseData: includeInHouse,
	}

	err := s.sendRequestAndDecodeJSON(req, &wsResponse)
	if err != nil {
		return nil, err
	}

	return wsResponse.Libraries, nil
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
		return Project{}, err
	}

	if len(wsResponse.ProjectVitals) == 0 {
		return Project{}, fmt.Errorf("no project with token '%s' found in WhiteSource: %w", projectToken, err)
	}

	return wsResponse.ProjectVitals[0], nil
}

// GetProjectByName fetches all projects and returns the one matching the given projectName, or none, if not found
func (s *System) GetProjectByName(productToken, projectName string) (Project, error) {
	projects, err := s.GetProjectsMetaInfo(productToken)
	if err != nil {
		return Project{}, fmt.Errorf("failed to retrieve WhiteSource project meta info: %w", err)
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
		return nil, fmt.Errorf("failed to retrieve WhiteSource project meta info: %w", err)
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
		return nil, fmt.Errorf("failed to retrieve WhiteSource project meta info: %w", err)
	}

	for _, project := range projects {
		for _, projectName := range projectNames {
			if projectName == project.Name {
				projectTokens = append(projectTokens, project.Token)
			}
		}
	}

	if len(projectNames) > 0 && len(projectTokens) == 0 {
		return projectTokens, fmt.Errorf("no project token(s) found for provided projects")
	}

	if len(projectNames) > 0 && len(projectNames) != len(projectTokens) {
		return projectTokens, fmt.Errorf("not all project token(s) found for provided projects")
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
		return "", err
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
		return nil, fmt.Errorf("WhiteSource getProjectRiskReport request failed: %w", err)
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
		return nil, fmt.Errorf("WhiteSource getProjectVulnerabilityReport request failed: %w", err)
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
		return nil, err
	}

	return wsResponse.Alerts, nil
}

// GetProjectAlertsByType returns all alerts of a certain type for a given project
func (s *System) GetProjectAlertsByType(projectToken, alertType string) ([]Alert, error) {
	wsResponse := struct {
		Alerts []Alert `json:"alerts"`
	}{
		Alerts: []Alert{},
	}

	req := Request{
		RequestType:  "getProjectAlertsByType",
		ProjectToken: projectToken,
		AlertType:    alertType,
	}

	err := s.sendRequestAndDecodeJSON(req, &wsResponse)
	if err != nil {
		return nil, err
	}

	return wsResponse.Alerts, nil
}

// GetProjectIgnoredAlertsByType returns all ignored alerts of a certain type for a given project
func (s *System) GetProjectIgnoredAlertsByType(projectToken string, alertType string) ([]Alert, error) {
	wsResponse := struct {
		Alerts []Alert `json:"alerts"`
	}{
		Alerts: []Alert{},
	}

	req := Request{
		RequestType:  "getProjectIgnoredAlerts",
		ProjectToken: projectToken,
	}

	err := s.sendRequestAndDecodeJSON(req, &wsResponse)
	if err != nil {
		return nil, err
	}

	alerts := make([]Alert, 0)
	for _, alert := range wsResponse.Alerts {
		if alert.Type == alertType {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
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
		return nil, err
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
		return fmt.Errorf("sending whiteSource request failed: %w", err)
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
					return fmt.Errorf("WhiteSource request failed after %v retries: %w", s.maxRetries, err)
				}
				return err
			}
		}
		return fmt.Errorf("invalid request, error code %v, message '%s'", errorResponse.ErrorCode, errorResponse.ErrorMessage)
	}

	if result != nil {
		err = json.Unmarshal(respBody, result)
		if err != nil {
			return fmt.Errorf("failed to parse WhiteSource response: %w", err)
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
		return responseBody, fmt.Errorf("failed to create WhiteSource request: %w", err)
	}

	log.Entry().Debugf("request: %v", string(body))

	headers := http.Header{}
	headers.Add("Content-Type", "application/json")
	response, err := s.httpClient.SendRequest(http.MethodPost, s.serverURL, bytes.NewBuffer(body), headers, nil)
	if err != nil {
		return responseBody, fmt.Errorf("failed to send request to WhiteSource: %w", err)
	}
	defer response.Body.Close()
	responseBody, err = io.ReadAll(response.Body)
	if err != nil {
		return responseBody, fmt.Errorf("failed to read WhiteSource response: %w", err)
	}

	return responseBody, nil
}
