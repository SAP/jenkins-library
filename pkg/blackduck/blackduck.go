package blackduck

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/package-url/packageurl-go"
)

// ReportsDirectory defines the subfolder for the BlackDuck reports which are generated
const ReportsDirectory = "blackduck"
const maxLimit = 50

const (
	HEADER_PROJECT_DETAILS_V4 = "application/vnd.blackducksoftware.project-detail-4+json"
	HEADER_USER_V4            = "application/vnd.blackducksoftware.user-4+json"
	HEADER_BOM_V6             = "application/vnd.blackducksoftware.bill-of-materials-6+json"
)

// Projects defines the response to a BlackDuck project API request
type Projects struct {
	TotalCount int       `json:"totalCount,omitempty"`
	Items      []Project `json:"items,omitempty"`
}

// Project defines a BlackDuck project
type Project struct {
	Name     string `json:"name,omitempty"`
	Metadata `json:"_meta,omitempty"`
}

// Metadata defines BlackDuck metadata for e.g. projects
type Metadata struct {
	Href  string `json:"href,omitempty"`
	Links []Link `json:"links,omitempty"`
}

// Link defines BlackDuck links to e.g. versions of projects
type Link struct {
	Rel  string `json:"rel,omitempty"`
	Href string `json:"href,omitempty"`
}

// ProjectVersions defines the response to a BlackDuck project version API request
type ProjectVersions struct {
	TotalCount int              `json:"totalCount,omitempty"`
	Items      []ProjectVersion `json:"items,omitempty"`
}

// ProjectVersion defines a version of a BlackDuck project
type ProjectVersion struct {
	Name     string `json:"versionName,omitempty"`
	Metadata `json:"_meta,omitempty"`
}

type Components struct {
	TotalCount int         `json:"totalCount,omitempty"`
	Items      []Component `json:"items,omitempty"`
}

type Component struct {
	Name                string            `json:"componentName,omitempty"`
	Version             string            `json:"componentVersionName,omitempty"`
	ComponentOriginName string            `json:"componentVersionOriginName,omitempty"`
	PrimaryLanguage     string            `json:"primaryLanguage,omitempty"`
	PolicyStatus        string            `json:"policyStatus,omitempty"`
	MatchTypes          []string          `json:"matchTypes,omitempty"`
	Origins             []ComponentOrigin `json:"origins,omitempty"`
	Metadata            `json:"_meta,omitempty"`
}

type ComponentOrigin struct {
	ExternalNamespace string `json:"externalNamespace,omitempty"`
	ExternalID        string `json:"externalId,omitempty"`
}

// ToPackageUrl creates the package URL for the component
func (c *Component) ToPackageUrl() *packageurl.PackageURL {
	purlParts := transformComponentOriginToPurlParts(c)

	// Namespace could not be in purlParts
	var purlType, namespace, name, version string
	if len(purlParts) >= 3 {
		version = purlParts[len(purlParts)-1]
		name = purlParts[len(purlParts)-2]
		purlType = purlParts[0]
	}
	if len(purlParts) == 4 {
		namespace = purlParts[1]
	}

	return packageurl.NewPackageURL(purlType, namespace, name, version, nil, "")
}

// MatchedType returns matched type of component: direct/transitive
func (c *Component) MatchedType() string {
	for _, matchedType := range c.MatchTypes {
		if matchedType == "FILE_DEPENDENCY_DIRECT" {
			return "direct"
		} else if matchedType == "FILE_DEPENDENCY_TRANSITIVE" {
			return "transitive"
		}
	}

	return ""
}

type Vulnerabilities struct {
	TotalCount int             `json:"totalCount,omitempty"`
	Items      []Vulnerability `json:"items,omitempty"`
}

type Vulnerability struct {
	Name                         string `json:"componentName,omitempty"`
	Version                      string `json:"componentVersionName,omitempty"`
	ComponentVersionOriginID     string `json:"componentVersionOriginId,omitempty"`
	ComponentVersionOriginName   string `json:"componentVersionOriginName,omitempty"`
	Ignored                      bool   `json:"ignored,omitempty"`
	VulnerabilityWithRemediation `json:"vulnerabilityWithRemediation,omitempty"`
	Component                    *Component
	projectName                  string
	projectVersion               string
	projectVersionLink           string
}

type VulnerabilityWithRemediation struct {
	VulnerabilityName      string  `json:"vulnerabilityName,omitempty"`
	BaseScore              float32 `json:"baseScore,omitempty"`
	Severity               string  `json:"severity,omitempty"`
	RemediationStatus      string  `json:"remediationStatus,omitempty"`
	RemediationComment     string  `json:"remediationComment,omitempty"`
	Description            string  `json:"description,omitempty"`
	OverallScore           float32 `json:"overallScore,omitempty"`
	CweID                  string  `json:"cweId,omitempty"`
	ExploitabilitySubscore float32 `json:"exploitabilitySubscore,omitempty"`
	ImpactSubscore         float32 `json:"impactSubscore,omitempty"`
	RelatedVulnerability   string  `json:"relatedVulnerability,omitempty"`
	RemidiatedBy           string  `json:"remediationCreatedBy,omitempty"`
}

// Title returns the issue title representation of the contents
func (v Vulnerability) Title() string {
	return v.VulnerabilityWithRemediation.VulnerabilityName
}

// ToMarkdown returns the markdown representation of the contents
func (v Vulnerability) ToMarkdown() ([]byte, error) {
	vul := reporting.VulnerabilityReport{
		ProjectName:          v.projectName,
		ProjectVersion:       v.projectVersion,
		BlackDuckProjectLink: v.projectVersionLink,
		ArtifactID:           v.Component.Name,
		Description:          v.Description,
		DependencyType:       v.Component.MatchedType(),
		Origin:               v.ComponentVersionOriginID,

		// no information available about footer, yet
		Footer: "",

		// no information available about group, yet
		Group: "",

		// no information available about publish date and resolution yet
		PublishDate: "",
		Resolution:  "",

		Score:             float64(v.VulnerabilityWithRemediation.BaseScore),
		Severity:          v.VulnerabilityWithRemediation.Severity,
		Version:           v.Version,
		PackageURL:        v.Component.ToPackageUrl().ToString(),
		VulnerabilityLink: v.RelatedVulnerability,
		VulnerabilityName: v.VulnerabilityName,
	}

	return vul.ToMarkdown()
}

// ToTxt returns the textual representation of the contents
func (v Vulnerability) ToTxt() string {
	return fmt.Sprintf(`Vulnerability %v
Severity: %v
Base (NVD) Score: %v
Temporal Score: %v
Package: %v
Installed Version: %v
Package URL: %v
Description: %v
Fix Resolution: %v
Link: [%v](%v)`,
		v.VulnerabilityName,
		v.Severity,
		v.VulnerabilityWithRemediation.BaseScore,
		v.VulnerabilityWithRemediation.OverallScore,
		v.Name,
		v.Version,
		v.Component.ToPackageUrl().ToString(),
		v.Description,
		"",
		"",
		"",
	)
}

type PolicyStatus struct {
	OverallStatus        string `json:"overallStatus,omitempty"`
	PolicyVersionDetails `json:"componentVersionPolicyViolationDetails,omitempty"`
}

type PolicyVersionDetails struct {
	Name           string           `json:"name,omitempty"`
	SeverityLevels []SeverityLevels `json:"severityLevels,omitEmpty"`
}

type SeverityLevels struct {
	Name  string `json:"name,omitempty"`
	Value int    `json:"value,omitempty"`
}

// Client defines a BlackDuck client
type Client struct {
	BearerToken                 string `json:"bearerToken,omitempty"`
	BearerExpiresInMilliseconds int64  `json:"expiresInMilliseconds,omitempty"`
	lastAuthentication          time.Time
	token                       string
	httpClient                  piperhttp.Sender
	serverURL                   string
	projectVersion              *ProjectVersion
}

// NewClient creates a new BlackDuck client
func NewClient(token, serverURL string, httpClient piperhttp.Sender) Client {
	return Client{
		httpClient: httpClient,
		serverURL:  serverURL,
		token:      token,
	}
}

// GetProject returns a project with a given name
func (b *Client) GetProject(projectName string) (*Project, error) {
	projects, err := b.getProjectByPagination(projectName, 0)
	if err != nil {
		return nil, err
	}
	// even if more than one projects found, let's return the first one with exact project name match
	for _, project := range projects.Items {
		if project.Name == projectName {
			return &project, nil
		}
	}

	if projects.TotalCount > maxLimit {
		offset := maxLimit
		totalProjects := projects.TotalCount
		for offset < totalProjects {
			projects, err = b.getProjectByPagination(projectName, offset)
			if err != nil {
				return nil, err
			}
			// even if more than one projects found, let's return the first one with exact project name match
			for _, project := range projects.Items {
				if project.Name == projectName {
					return &project, nil
				}
			}
			offset += maxLimit
		}
	}

	return nil, fmt.Errorf("project '%v' not found", projectName)
}

func (b *Client) getProjectByPagination(projectName string, offset int) (*Projects, error) {
	if !b.authenticationValid(time.Now()) {
		if err := b.authenticate(); err != nil {
			return nil, err
		}
	}
	headers := http.Header{}
	headers.Add("Accept", HEADER_PROJECT_DETAILS_V4)
	queryParams := map[string]string{
		"q":      fmt.Sprintf("name:%v", projectName),
		"limit":  fmt.Sprint(maxLimit),
		"offset": fmt.Sprint(offset),
		"sort":   "asc",
	}
	respBody, err := b.sendRequest("GET", "/api/projects", queryParams, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to get project '%v': %w", projectName, err)
	}
	projects := Projects{}
	err = json.Unmarshal(respBody, &projects)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve details for project '%v': %w", projectName, err)
	} else if projects.TotalCount == 0 {
		return nil, fmt.Errorf("project '%v' not found", projectName)
	}
	return &projects, nil
}

// GetProjectVersion returns a project version with a given name
func (b *Client) GetProjectVersion(projectName, projectVersion string) (*ProjectVersion, error) {
	// get version from cache if it is there
	if b.projectVersion != nil {
		return b.projectVersion, nil
	}
	project, err := b.GetProject(projectName)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Add("Accept", HEADER_PROJECT_DETAILS_V4)

	var versionPath string
	for _, link := range project.Links {
		if link.Rel == "versions" {
			versionPath = urlPath(link.Href)
			break
		}
	}

	//While sending a request to 'versions', get all 100 versions from that project by setting limit=100
	//More than 100 project versions is currently not supported/recommended by BlackDuck
	respBody, err := b.sendRequest("GET", versionPath, map[string]string{"offset": "0", "limit": "100"}, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to get project version '%v:%v': %w", projectName, projectVersion, err)
	}

	projectVersions := ProjectVersions{}
	err = json.Unmarshal(respBody, &projectVersions)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve details for project version '%v:%v': %w", projectName, projectVersion, err)
	} else if projectVersions.TotalCount == 0 {
		return nil, fmt.Errorf("project version '%v:%v' not found", projectName, projectVersion)
	}

	// even if more than one projects found, let's return the first one with exact project name match
	for _, version := range projectVersions.Items {
		if version.Name == projectVersion {
			// save version to cache in order not to do several same requests
			b.projectVersion = &version
			return &version, nil
		}
	}

	return nil, fmt.Errorf("failed to get project version '%v'", projectVersion)
}

func (b *Client) GetProjectVersionLink(projectName, versionName string) (string, error) {
	projectVersion, err := b.GetProjectVersion(projectName, versionName)
	if err != nil {
		return "", err
	}
	if projectVersion != nil {
		return projectVersion.Href, nil
	}
	return "", nil
}

func (b *Client) GetComponents(projectName, versionName string) (*Components, error) {
	projectVersion, err := b.GetProjectVersion(projectName, versionName)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Add("Accept", HEADER_BOM_V6)

	var componentsPath string
	for _, link := range projectVersion.Links {
		if link.Rel == "components" {
			componentsPath = urlPath(link.Href)
			break
		}
	}

	respBody, err := b.sendRequest("GET", componentsPath, map[string]string{"offset": "0", "limit": "999"}, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("Failed to get components list for project version '%v:%v': %w", projectName, versionName, err)
	}

	components := Components{}
	err = json.Unmarshal(respBody, &components)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve component details for project version '%v:%v': %w", projectName, versionName, err)
	} else if components.TotalCount == 0 {
		return nil, fmt.Errorf("No Components found for project version '%v:%v'", projectName, versionName)
	}

	//Just return the components, the details of the components are not necessary
	return &components, nil
}
func (b *Client) GetComponentsWithLicensePolicyRule(projectName, versionName string) (*Components, error) {
	projectVersion, err := b.GetProjectVersion(projectName, versionName)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Add("Accept", HEADER_BOM_V6)

	var componentsPath string
	for _, link := range projectVersion.Links {
		if link.Rel == "components" {
			componentsPath = urlPath(link.Href)
			break
		}
	}

	respBody, err := b.sendRequest("GET", componentsPath, map[string]string{"offset": "0", "limit": "999", "filter": "policyCategory:license"}, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("Failed to get components list for project version '%v:%v': %w", projectName, versionName, err)
	}

	components := Components{}
	err = json.Unmarshal(respBody, &components)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve component details for project version '%v:%v': %w", projectName, versionName, err)
	}

	//Just return the components, the details of the components are not necessary
	return &components, nil
}

// func (b *Client) GetComponentPolicyStatus(component Component) (ComponentPolicyStatus, error) {
// 	var policyStatusUrl string
// 	for _, link := range component.Links {
// 		if link.Rel == "policy-status" {
// 			policyStatusUrl = urlPath(link.Href)
// 		}
// 	}

// 	headers := http.Header{}
// 	headers.Add("Accept", HEADER_BOM_V6)

// 	respBody, err := b.sendRequest("GET", policyStatusUrl, map[string]string{}, nil, headers)
// }

func (b *Client) GetVulnerabilities(projectName, versionName string) (*Vulnerabilities, error) {
	projectVersion, err := b.GetProjectVersion(projectName, versionName)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Add("Accept", HEADER_BOM_V6)

	var vulnerableComponentsPath string
	for _, link := range projectVersion.Links {
		if link.Rel == "vulnerable-components" {
			vulnerableComponentsPath = urlPath(link.Href)
			break
		}
	}

	respBody, err := b.sendRequest("GET", vulnerableComponentsPath, map[string]string{"offset": "0", "limit": "999"}, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("Failed to get Vulnerabilties for project version '%v:%v': %w", projectName, versionName, err)
	}

	vulnerabilities := Vulnerabilities{}
	err = json.Unmarshal(respBody, &vulnerabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Vulnerability details for project version '%v:%v': %w", projectName, versionName, err)
	}

	return &vulnerabilities, nil
}

func (b *Client) GetPolicyStatus(projectName, versionName string) (*PolicyStatus, error) {
	projectVersion, err := b.GetProjectVersion(projectName, versionName)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Add("Accept", HEADER_BOM_V6)

	var policyStatusPath string
	for _, link := range projectVersion.Links {
		if link.Rel == "policy-status" {
			policyStatusPath = urlPath(link.Href)
			break
		}
	}

	respBody, err := b.sendRequest("GET", policyStatusPath, map[string]string{}, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("Failed to get Policy Violation status for project version '%v:%v': %w", projectName, versionName, err)
	}

	policyStatus := PolicyStatus{}
	err = json.Unmarshal(respBody, &policyStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Policy violation details for project version '%v:%v': %w", projectName, versionName, err)
	}

	return &policyStatus, nil
}

func (b *Client) authenticate() error {
	headers := http.Header{}
	headers.Add("Authorization", fmt.Sprintf("token %v", b.token))
	headers.Add("Accept", HEADER_USER_V4)
	b.lastAuthentication = time.Now()
	respBody, err := b.sendRequest(http.MethodPost, "/api/tokens/authenticate", map[string]string{}, nil, headers)
	if err != nil {
		return fmt.Errorf("authentication to BlackDuck API failed: %w", err)
	}
	err = json.Unmarshal(respBody, b)
	if err != nil {
		return fmt.Errorf("failed to parse BlackDuck response: %w", err)
	}
	return nil
}

func (b *Client) sendRequest(method, apiEndpoint string, params map[string]string, body io.Reader, header http.Header) ([]byte, error) {
	responseBody := []byte{}

	blackDuckAPIUrl, err := b.apiURL(apiEndpoint)
	if err != nil {
		return responseBody, fmt.Errorf("failed to get api url: %w", err)
	}

	q := url.Values{}
	for key, val := range params {
		q.Add(key, val)
	}
	blackDuckAPIUrl.RawQuery = q.Encode()

	if len(b.BearerToken) > 0 {
		header.Add("Authorization", fmt.Sprintf("Bearer %v", b.BearerToken))
	}

	response, err := b.httpClient.SendRequest(method, blackDuckAPIUrl.String(), nil, header, nil)
	if err != nil {
		return responseBody, fmt.Errorf("request to BlackDuck API failed: %w", err)
	}

	responseBody, err = io.ReadAll(response.Body)
	if err != nil {
		return responseBody, fmt.Errorf("reading BlackDuck response failed: %w", err)
	}
	return responseBody, nil
}

func (b *Client) apiURL(apiEndpoint string) (*url.URL, error) {
	blackDuckURL, err := url.Parse(b.serverURL)
	if err != nil {
		return nil, err
	}
	blackDuckURL.Path = path.Join(blackDuckURL.Path, apiEndpoint)
	return blackDuckURL, nil
}

func (b *Client) authenticationValid(now time.Time) bool {
	// check bearer token timeout
	expiryTime := b.lastAuthentication.Add(time.Millisecond * time.Duration(b.BearerExpiresInMilliseconds))
	return now.Sub(expiryTime) < 0
}

func urlPath(fullUrl string) string {
	theUrl, _ := url.Parse(fullUrl)
	return theUrl.Path
}

func transformComponentOriginToPurlParts(component *Component) []string {
	result := []string{}
	purlType := packageurl.TypeGeneric
	gav := []string{"", component.Name, component.Version}
	origins := component.Origins
	if origins != nil && len(origins) > 0 {
		if strings.Contains(origins[0].ExternalID, "/") {
			gav = strings.Split(origins[0].ExternalID, "/")
		} else {
			gav = strings.Split(origins[0].ExternalID, ":")
		}
		switch strings.ToLower(origins[0].ExternalNamespace) {
		case "maven":
			purlType = packageurl.TypeMaven
		case "node":
			purlType = packageurl.TypeNPM
		case "npmjs":
			purlType = packageurl.TypeNPM
		case "golang":
			purlType = packageurl.TypeGolang
		case "docker":
			purlType = packageurl.TypeDocker
		case "":
			purlType = packageurl.TypeGeneric
		default:
			purlType = strings.ToLower(origins[0].ExternalNamespace)
		}
	}
	result = append(result, purlType)
	result = append(result, gav...)

	if len(result) > 0 && !strings.Contains(result[len(result)-1], ".") {
		result = result[:len(result)-1]
	}

	return result
}
