package sonar

import (
	"net/http"

	sonargo "github.com/magicsong/sonargo/sonar"
	"github.com/pkg/errors"
)

// EndpointIssuesSearch API endpoint for https://sonarcloud.io/web_api/api/issues/search
const EndpointIssuesSearch = "issues/search"

// IssueService ...
type IssueService struct {
	Organization string
	Project      string
	Branch       string
	PullRequest  string
	apiClient    *Requester
}

// SearchIssues ...
func (service *IssueService) SearchIssues(options *IssuesSearchOption) (*sonargo.IssuesSearchObject, *http.Response, error) {
	request, err := service.apiClient.create("GET", EndpointIssuesSearch, options)
	if err != nil {
		return nil, nil, err
	}
	// use custom HTTP client to send request
	response, err := service.apiClient.send(request)
	if err != nil {
		return nil, nil, err
	}
	// reuse response verrification from sonargo
	err = sonargo.CheckResponse(response)
	if err != nil {
		return nil, response, err
	}
	// decode JSON response
	result := new(sonargo.IssuesSearchObject)
	err = service.apiClient.decode(response, result)
	if err != nil {
		return nil, response, err
	}
	return result, response, nil
}

func (service *IssueService) getIssueCount(severity issueSeverity) (int, error) {
	options := &IssuesSearchOption{
		ComponentKeys: service.Project,
		Severities:    severity.ToString(),
		Resolved:      "false",
		Ps:            "1",
	}
	if len(service.Organization) > 0 {
		options.Organization = service.Organization
	}
	// if PR, ignore branch name and consider PR branch name. If not PR, consider branch name
	if len(service.PullRequest) > 0 {
		options.PullRequest = service.PullRequest
	} else if len(service.Branch) > 0 {
		options.Branch = service.Branch
	}
	result, _, err := service.SearchIssues(options)
	if err != nil {
		return -1, errors.Wrapf(err, "failed to fetch the numer of '%s' issues", severity)
	}
	return result.Total, nil
}

// GetNumberOfBlockerIssues returns the number of issue with BLOCKER severity.
func (service *IssueService) GetNumberOfBlockerIssues() (int, error) {
	return service.getIssueCount(blocker)
}

// GetNumberOfCriticalIssues returns the number of issue with CRITICAL severity.
func (service *IssueService) GetNumberOfCriticalIssues() (int, error) {
	return service.getIssueCount(critical)
}

// GetNumberOfMajorIssues returns the number of issue with MAJOR severity.
func (service *IssueService) GetNumberOfMajorIssues() (int, error) {
	return service.getIssueCount(major)
}

// GetNumberOfMinorIssues returns the number of issue with MINOR severity.
func (service *IssueService) GetNumberOfMinorIssues() (int, error) {
	return service.getIssueCount(minor)
}

// GetNumberOfInfoIssues returns the number of issue with INFO severity.
func (service *IssueService) GetNumberOfInfoIssues() (int, error) {
	return service.getIssueCount(info)
}

// NewIssuesService returns a new instance of a service for the issues API endpoint.
func NewIssuesService(host, token, project, organization, branch, pullRequest string, client Sender) *IssueService {
	return &IssueService{
		Organization: organization,
		Project:      project,
		Branch:       branch,
		PullRequest:  pullRequest,
		apiClient:    NewAPIClient(host, token, client),
	}
}
