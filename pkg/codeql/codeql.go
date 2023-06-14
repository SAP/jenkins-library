package codeql

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

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
	totalAlerts, err := getTotalAlertsFromClient(ctx, client.CodeScanning, analyzedRef, codeqlScanAudit)

	return getVulnerabilitiesFromClient(ctx, client.CodeScanning, analyzedRef, codeqlScanAudit, totalAlerts)
}

type SarifFileInfo struct {
	ProcessingStatus string `json:"processing_status"`
	Errors []string `json:"errors"`
}

func (codeqlScanAudit *CodeqlScanAuditInstance) GetSarifUploadingStatus(sarifURL string) (SarifFileInfo, error) {
	client := http.Client{}
	req , err := http.NewRequest("GET", sarifURL, nil)
	if err != nil {
		return SarifFileInfo{}, err
	}
	req.Header.Add("Authorization", "Bearer "+codeqlScanAudit.token)
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

func getTotalAlertsFromClient(ctx context.Context, codeScannning githubCodeqlScanningService, analyzedRef string, codeqlScanAudit *CodeqlScanAuditInstance) (int, error) {
	analysesOptions := github.AnalysesListOptions{
		Ref: &analyzedRef,
	}
	analyses, _, err := codeScannning.ListAnalysesForRepo(ctx, codeqlScanAudit.owner, codeqlScanAudit.repository, &analysesOptions)
	if err != nil {
		return 0, err
	}
	if len(analyses) < 1 {
		return 0, errors.New("analyses for ref not found")
	}
	return *analyses[0].ResultsCount, nil
}

func getVulnerabilitiesFromClient(ctx context.Context, codeScanning githubCodeqlScanningService, analyzedRef string, codeqlScanAudit *CodeqlScanAuditInstance, totalAlerts int) (CodeqlScanning, error) {
	pages := totalAlerts/perPageCount + 1
	errChan := make(chan error)
	openStateCountChan := make(chan int)
	for page := 1; page <= pages; page++ {
		go func(i int) {
			alertOptions := github.AlertListOptions{
				State: "",
				Ref:   analyzedRef,
				ListOptions: github.ListOptions{
					Page:    i,
					PerPage: perPageCount,
				},
			}

			alerts, _, err := codeScanning.ListAlertsForRepo(ctx, codeqlScanAudit.owner, codeqlScanAudit.repository, &alertOptions)
			if err != nil {
				errChan <- err
				return
			}

			openStateCount := 0
			for _, alert := range alerts {
				if *alert.State == auditStateOpen {
					openStateCount = openStateCount + 1
				}
			}
			openStateCountChan <- len(alerts) - openStateCount
		}(page)
	}

	codeqlScanning := CodeqlScanning{}
	codeqlScanning.Total = totalAlerts
	for i := 0; i < pages; i++ {
		select {
		case openStateCount := <-openStateCountChan:
			codeqlScanning.Audited += openStateCount
		case err := <-errChan:
			return CodeqlScanning{}, err
		}
	}

	return codeqlScanning, nil
}

func getApiUrl(serverUrl string) string {
	if serverUrl == "https://github.com" {
		return "https://api.github.com"
	}

	return (serverUrl + "/api/v3")
}
