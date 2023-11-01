package codeql

import (
	"context"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/google/go-github/v45/github"
)

type CodeqlScanAudit interface {
	GetVulnerabilities(analyzedRef string, state string) error
}

type githubCodeqlScanningService interface {
	ListAlertsForRepo(ctx context.Context, owner, repo string, opts *github.AlertListOptions) ([]*github.Alert, *github.Response, error)
}

const auditStateOpen string = "open"
const auditStateDismissed string = "dismissed"
const codeqlToolName string = "CodeQL"
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

func (codeqlScanAudit *CodeqlScanAuditInstance) GetVulnerabilities(analyzedRef string) ([]CodeqlFindings, error) {
	apiUrl := getApiUrl(codeqlScanAudit.serverUrl)
	ctx, client, err := piperGithub.
		NewClientBuilder(codeqlScanAudit.token, apiUrl).
		WithTrustedCerts(codeqlScanAudit.trustedCerts).Build()
	if err != nil {
		return []CodeqlFindings{}, err
	}

	return getVulnerabilitiesFromClient(ctx, client.CodeScanning, analyzedRef, codeqlScanAudit)
}

func getVulnerabilitiesFromClient(ctx context.Context, codeScanning githubCodeqlScanningService, analyzedRef string, codeqlScanAudit *CodeqlScanAuditInstance) ([]CodeqlFindings, error) {
	page := 1
	audited := 0
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
			return []CodeqlFindings{}, err
		}

		page = response.NextPage

		for _, alert := range alerts {
			if *alert.Tool.Name != codeqlToolName {
				continue
			}

			if *alert.State == auditStateDismissed {
				audited += 1
				totalAlerts += 1
			}

			if *alert.State == auditStateOpen {
				totalAlerts += 1
			}
		}
	}

	auditAll := CodeqlFindings{
		ClassificationName: "Audit All",
		Total:              totalAlerts,
		Audited:            audited,
	}
	codeqlScanning := []CodeqlFindings{auditAll}

	return codeqlScanning, nil
}

func getApiUrl(serverUrl string) string {
	if serverUrl == "https://github.com" {
		return "https://api.github.com"
	}

	return (serverUrl + "/api/v3")
}
