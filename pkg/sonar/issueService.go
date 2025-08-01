package sonar

import (
	"net/http"
	"net/http/httputil"

	"github.com/SAP/jenkins-library/pkg/log"
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

	// log response
	log.Entry().Debugf("HTTP Response: %v", func() string { rsp, _ := httputil.DumpResponse(response, true); return string(rsp) }())

	// decode JSON response
	result := new(sonargo.IssuesSearchObject)
	err = service.apiClient.decode(response, result)
	if err != nil {
		return nil, response, err
	}
	return result, response, nil
}

func (service *IssueService) getIssueCount(severity issueSeverity, categories *[]Severity) (int, error) {
	options := &IssuesSearchOption{
		ComponentKeys: service.Project,
		Severities:    severity.ToString(),
		Resolved:      "false",
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

	table := map[string]int{}
	service.updateIssueTypesTable(result.Issues, table)
	for issueType, issuesCount := range table {
		var severityResult Severity
		severityResult.SeverityType = severity.ToString()
		severityResult.IssueType = issueType
		severityResult.IssueCount = issuesCount
		*categories = append(*categories, severityResult)
	}
	return result.Total, nil
}

func (service *IssueService) updateIssueTypesTable(issues []*sonargo.Issue, table map[string]int) {
	for _, issue := range issues {
		table[issue.Type]++
	}
	delete(table, "") // remove undefined key if any exists in response
}

// GetNumberOfBlockerIssues returns the number of issue with BLOCKER severity.
func (service *IssueService) GetNumberOfBlockerIssues(categories *[]Severity) (int, error) {
	return service.getIssueCount(blocker, categories)
}

// GetNumberOfCriticalIssues returns the number of issue with CRITICAL severity.
func (service *IssueService) GetNumberOfCriticalIssues(categories *[]Severity) (int, error) {
	return service.getIssueCount(critical, categories)
}

// GetNumberOfMajorIssues returns the number of issue with MAJOR severity.
func (service *IssueService) GetNumberOfMajorIssues(categories *[]Severity) (int, error) {
	return service.getIssueCount(major, categories)
}

// GetNumberOfMinorIssues returns the number of issue with MINOR severity.
func (service *IssueService) GetNumberOfMinorIssues(categories *[]Severity) (int, error) {
	return service.getIssueCount(minor, categories)
}

// GetNumberOfInfoIssues returns the number of issue with INFO severity.
func (service *IssueService) GetNumberOfInfoIssues(categories *[]Severity) (int, error) {
	return service.getIssueCount(info, categories)
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
