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
	ListAnalysesForRepo(ctx context.Context, owner, repo string, opts *github.AnalysesListOptions) ([]*github.ScanningAnalysis, *github.Response, error)
}

const auditStateOpen string = "open"
const perPageCount int = 100

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
	page := 1
	openStateCount := 0
	totalAlerts := 0

	for page != 0 {
		alertOptions := github.AlertListOptions{
			State: "",
			Ref:   analyzedRef,
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPageCount,
			},
		}

		alerts, response, err := codeScanning.ListAlertsForRepo(ctx, codeqlScanAudit.owner, codeqlScanAudit.repository, &alertOptions)
		if err != nil {
			return CodeqlScanning{}, err
		}

		page = response.NextPage

		for _, alert := range alerts {
			totalAlerts += 1
			if *alert.State == auditStateOpen {
				openStateCount += 1
			}
		}
	}

	codeqlScanning := CodeqlScanning{}
	codeqlScanning.Total = totalAlerts
	codeqlScanning.Audited = (totalAlerts - openStateCount)

	return codeqlScanning, nil
}

func getApiUrl(serverUrl string) string {
	if serverUrl == "https://github.com" {
		return "https://api.github.com"
	}

	return (serverUrl + "/api/v3")
}
