package sonar

import (
	sonarAPI "github.com/magicsong/sonargo/sonar"
	"github.com/pkg/errors"
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
)

func (api *IssueService) GetNumberOfBlockerIssues() (int, error) {
	return api.getIssueCount(issueSeverityBlocker)
}

func (api *IssueService) GetNumberOfCriticalIssues() (int, error) {
	return api.getIssueCount(issueSeverityCritical)
}

func (api *IssueService) getIssueCount(severity issueSeverity) (int, error) {
	if api.client == nil {
		if err := api.createClient(); err != nil {
			return -1, err
		}
	}

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
