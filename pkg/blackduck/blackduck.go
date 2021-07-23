package blackduck

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
)

const (
	HEADER_PROJECT_DETAILS_V4 = "application/vnd.blackducksoftware.project-detail-4+json"
	HEADER_USER_V4            = "application/vnd.blackducksoftware.user-4+json"
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

// Client defines a BlackDuck client
type Client struct {
	BearerToken                 string `json:"bearerToken,omitempty"`
	BearerExpiresInMilliseconds int64  `json:"expiresInMilliseconds,omitempty"`
	lastAuthentication          time.Time
	token                       string
	httpClient                  piperhttp.Sender
	serverURL                   string
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
	if !b.authenticationValid(time.Now()) {
		if err := b.authenticate(); err != nil {
			return nil, err
		}
	}
	headers := http.Header{}
	headers.Add("Accept", HEADER_PROJECT_DETAILS_V4)
	respBody, err := b.sendRequest("GET", "/api/projects", map[string]string{"q": fmt.Sprintf("name:%v", projectName)}, nil, headers)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project '%v'", projectName)
	}

	projects := Projects{}
	err = json.Unmarshal(respBody, &projects)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve details for project '%v'", projectName)
	} else if projects.TotalCount == 0 {
		return nil, fmt.Errorf("project '%v' not found", projectName)
	}

	// even if more than one projects found, let's return the first one with exact project name match
	for _, project := range projects.Items {
		if project.Name == projectName {
			return &project, nil
		}
	}

	return nil, fmt.Errorf("project '%v' not found", projectName)
}

// GetProjectVersion returns a project with a given name
func (b *Client) GetProjectVersion(projectName, projectVersion string) (*ProjectVersion, error) {
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

	respBody, err := b.sendRequest("GET", versionPath, map[string]string{}, nil, headers)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project version '%v:%v'", projectName, projectVersion)
	}

	projectVersions := ProjectVersions{}
	err = json.Unmarshal(respBody, &projectVersions)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve details for project version '%v:%v'", projectName, projectVersion)
	} else if projectVersions.TotalCount == 0 {
		return nil, fmt.Errorf("project version '%v:%v' not found", projectName, projectVersion)
	}

	// even if more than one projects found, let's return the first one with exact project name match
	for _, version := range projectVersions.Items {
		if version.Name == projectVersion {
			return &version, nil
		}
	}

	return nil, fmt.Errorf("failed to get project version '%v'", projectVersion)
}

func (b *Client) authenticate() error {
	headers := http.Header{}
	headers.Add("Authorization", fmt.Sprintf("token %v", b.token))
	headers.Add("Accept", HEADER_USER_V4)
	b.lastAuthentication = time.Now()
	respBody, err := b.sendRequest(http.MethodPost, "/api/tokens/authenticate", map[string]string{}, nil, headers)
	if err != nil {
		return errors.Wrap(err, "authentication to BlackDuck API failed")
	}
	err = json.Unmarshal(respBody, b)
	if err != nil {
		return errors.Wrap(err, "failed to parse BlackDuck response")
	}
	return nil
}

func (b *Client) sendRequest(method, apiEndpoint string, params map[string]string, body io.Reader, header http.Header) ([]byte, error) {
	responseBody := []byte{}

	blackDuckAPIUrl, err := b.apiURL(apiEndpoint)
	if err != nil {
		return responseBody, errors.Wrap(err, "failed to get api url")
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
		return responseBody, errors.Wrap(err, "request to BlackDuck API failed")
	}

	responseBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return responseBody, errors.Wrap(err, "reading BlackDuck response failed")
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
	// //check bearer token timeout
	expiryTime := b.lastAuthentication.Add(time.Millisecond * time.Duration(b.BearerExpiresInMilliseconds))
	return now.Sub(expiryTime) < 0
}

func urlPath(fullUrl string) string {
	theUrl, _ := url.Parse(fullUrl)
	return theUrl.Path
}
