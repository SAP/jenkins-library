package codeql

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	sapgithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/google/go-github/v45/github"
)

type CodeqlScanAudit interface {
	GetVulnerabilities(analyzedRef string, state string) error
}

type CodeqlSarifUploader interface {
	GetSarifStatus() (SarifFileInfo, error)
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

func (codeqlScanAudit *CodeqlScanAuditInstance) GetVulnerabilities(analyzedRef string) (CodeqlScanning, error) {
	apiUrl := getApiUrl(codeqlScanAudit.serverUrl)
	ctx, client, err := sapgithub.NewClient(codeqlScanAudit.token, apiUrl, "", codeqlScanAudit.trustedCerts)
	if err != nil {
		return CodeqlScanning{}, err
	}

	return getVulnerabilitiesFromClient(ctx, client.CodeScanning, analyzedRef, codeqlScanAudit)
}

func NewCodeqlSarifUploaderInstance(url, token string) CodeqlSarifUploaderInstance {
	return CodeqlSarifUploaderInstance{url: url, token: token}
}

type CodeqlSarifUploaderInstance struct {
	url   string
	token string
}

func (codeqlSarifUploader *CodeqlSarifUploaderInstance) GetSarifStatus() (SarifFileInfo, error) {
	return getSarifUploadingStatus(codeqlSarifUploader.url, codeqlSarifUploader.token)
}

type SarifFileInfo struct {
	ProcessingStatus string   `json:"processing_status"`
	Errors           []string `json:"errors"`
}

func getSarifUploadingStatus(sarifURL, token string) (SarifFileInfo, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", sarifURL, nil)
	if err != nil {
		return SarifFileInfo{}, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return SarifFileInfo{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	sarifInfo := SarifFileInfo{}
	err = json.Unmarshal(body, &sarifInfo)
	if err != nil {
		return SarifFileInfo{}, err
	}
	return sarifInfo, nil
}

func getVulnerabilitiesFromClient(ctx context.Context, codeScanning githubCodeqlScanningService, analyzedRef string, codeqlScanAudit *CodeqlScanAuditInstance) (CodeqlScanning, error) {
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
			return CodeqlScanning{}, err
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

	codeqlScanning := CodeqlScanning{}
	codeqlScanning.Total = totalAlerts
	codeqlScanning.Audited = audited

	return codeqlScanning, nil
}

func getApiUrl(serverUrl string) string {
	if serverUrl == "https://github.com" {
		return "https://api.github.com"
	}

	return (serverUrl + "/api/v3")
}
