package sonar

import (
	sonarAPI "github.com/magicsong/sonargo/sonar"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/log"
)

type IssueService struct {
	Host    string
	Token   string
	Project string
	client  *sonarAPI.Client
}

type issueSeverity string

func (s issueSeverity) ToString() string {
	return string(s)
}

const (
	issueSeverityBlocker  issueSeverity = "BLOCKER"
	issueSeverityCritical issueSeverity = "CRITICAL"
	issueSeverityMajor    issueSeverity = "MAJOR"
	issueSeverityMinor    issueSeverity = "MINOR"
	issueSeverityInfo     issueSeverity = "INFO"
)

func (api *IssueService) GetNumberOfBlockerIssues() (int, error) {
	return api.getIssueCount(issueSeverityBlocker)
}

func (api *IssueService) GetNumberOfCriticalIssues() (int, error) {
	return api.getIssueCount(issueSeverityCritical)
}

func (api *IssueService) GetNumberOfMajorIssues() (int, error) {
	return api.getIssueCount(issueSeverityMajor)
}

func (api *IssueService) GetNumberOfMinorIssues() (int, error) {
	return api.getIssueCount(issueSeverityMinor)
}

func (api *IssueService) GetNumberOfInfoIssues() (int, error) {
	return api.getIssueCount(issueSeverityInfo)
}

func (api *IssueService) getIssueCount(severity issueSeverity) (int, error) {
	if api.client == nil {
		if err := api.createClient(); err != nil {
			return -1, err
		}
	}
	log.Entry().Debugf("using api client for '%s'", api.Host)
	result, _, err := api.client.Issues.Search(&sonarAPI.IssuesSearchOption{
		ComponentKeys: api.Project,
		Severities:    severity.ToString(),
		Resolved:      "false",
		Ps:            "1",
	})
	if err != nil {
		return -1, errors.Wrapf(err, "failed to fetch the numer of '%s' issues", severity)
	}
	return result.Total, nil
}
