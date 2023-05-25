package codeql

import (
	"context"

	sapgithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/google/go-github/v45/github"
)

type CodeqlScanAudit interface {
	GetVulnerabilities(analyzedRef string, state string) error
}

type githubCodeqlScanningService interface {
	ListAlertsForRepo(ctx context.Context, owner, repo string, opts *github.AlertListOptions) ([]*github.Alert, *github.Response, error)
}

const auditStateOpen = "open"

func NewCodeqlScanAuditInstance(serverUrl, owner, repository, token string, trustedCerts []string) CodeqlScanAuditInstance {
	return CodeqlScanAuditInstance{serverUrl: serverUrl, owner: owner, repository: repository, token: token, trustedCerts: trustedCerts}
}

type CodeqlScanAuditInstance struct {
	serverUrl        string
	owner            string
	repository       string
	token            string
	trustedCerts     []string
	alertListoptions github.AlertListOptions
}

func (codeqlScanAudit *CodeqlScanAuditInstance) GetVulnerabilities(analyzedRef string) (CodeqlScanning, error) {
	apiUrl := getApiUrl(codeqlScanAudit.serverUrl)
	ctx, client, err := sapgithub.NewClient(codeqlScanAudit.token, apiUrl, "", codeqlScanAudit.trustedCerts)
	if err != nil {
		return CodeqlScanning{}, err
	}

	return getVulnerabilitiesFromClient(ctx, client.CodeScanning, analyzedRef, codeqlScanAudit)
}

func getVulnerabilitiesFromClient(ctx context.Context, codeScanning githubCodeqlScanningService, analyzedRef string, codeqlScanAudit *CodeqlScanAuditInstance) (CodeqlScanning, error) {
	alertOptions := github.AlertListOptions{
		State:       "",
		Ref:         analyzedRef,
		ListOptions: github.ListOptions{},
	}

	alerts, _, err := codeScanning.ListAlertsForRepo(ctx, codeqlScanAudit.owner, codeqlScanAudit.repository, &alertOptions)
	if err != nil {
		return CodeqlScanning{}, err
	}

	openStateCount := 0
	for _, alert := range alerts {
		if *alert.State == auditStateOpen {
			openStateCount = openStateCount + 1
		}
	}

	codeqlScanning := CodeqlScanning{}
	codeqlScanning.Total = len(alerts)
	codeqlScanning.Audited = (codeqlScanning.Total - openStateCount)
	return codeqlScanning, nil
}

func getApiUrl(serverUrl string) string {
	if serverUrl == "https://github.com" {
		return "https://api.github.com"
	}

	return (serverUrl + "/api/v3")
}
