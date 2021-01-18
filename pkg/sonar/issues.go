package sonar

import (
	"strings"

	sonargo "github.com/magicsong/sonargo/sonar"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/log"
)

func NewIssuesService(host, token, project string, client Sender) *IssueService {
	// Make sure the given URL end with a slash
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
	// Make sure the given URL end with a api/
	if !strings.HasSuffix(host, "api/") {
		host += "api/"
	}
	return &IssueService{
		Host:      host,
		Token:     token,
		Project:   project,
		apiClient: NewBasicAuthClient(token, "", host, client),
	}
}

type IssueService struct {
	Host      string
	Token     string
	Project   string
	apiClient *Requester
}

type issueSeverity string

func (s issueSeverity) ToString() string {
	return string(s)
}

const (
	blocker  issueSeverity = "BLOCKER"
	critical issueSeverity = "CRITICAL"
	major    issueSeverity = "MAJOR"
	minor    issueSeverity = "MINOR"
	info     issueSeverity = "INFO"
)

func (service *IssueService) GetNumberOfBlockerIssues() (int, error) {
	return service.getIssueCount(blocker)
}

func (service *IssueService) GetNumberOfCriticalIssues() (int, error) {
	return service.getIssueCount(critical)
}

func (service *IssueService) GetNumberOfMajorIssues() (int, error) {
	return service.getIssueCount(major)
}

func (service *IssueService) GetNumberOfMinorIssues() (int, error) {
	return service.getIssueCount(minor)
}

func (service *IssueService) GetNumberOfInfoIssues() (int, error) {
	return service.getIssueCount(info)
}

func (service *IssueService) getIssueCount(severity issueSeverity) (int, error) {
	log.Entry().Debugf("using api client for '%s'", service.Host)
	result, _, err := service.apiClient.SearchIssues(&sonargo.IssuesSearchOption{
		ComponentKeys: service.Project,
		Severities:    severity.ToString(),
		Resolved:      "false",
		Ps:            "1",
	})
	if err != nil {
		return -1, errors.Wrapf(err, "failed to fetch the numer of '%s' issues", severity)
	}
	return result.Total, nil
}
