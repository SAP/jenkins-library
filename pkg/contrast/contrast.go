package contrast

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	StatusReported  = "REPORTED"
	Critical        = "CRITICAL"
	High            = "HIGH"
	Medium          = "MEDIUM"
	AuditAll        = "Audit All"
	Optional        = "Optional"
	pageSize        = 100
	startPage       = 0
	ContentType     = "Content-Type"
	JSONContentType = "application/json"
)

type VulnerabilitiesResponse struct {
	Size            int             `json:"size"`
	TotalElements   int             `json:"totalElements"`
	TotalPages      int             `json:"totalPages"`
	Empty           bool            `json:"empty"`
	First           bool            `json:"first"`
	Last            bool            `json:"last"`
	Vulnerabilities []Vulnerability `json:"content"`
}

type Vulnerability struct {
	Severity string `json:"severity"`
	Status   string `json:"status"`
}

type ApplicationResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Path        string `json:"path"`
	Language    string `json:"language"`
	Importance  string `json:"importance"`
}

type Contrast interface {
	GetVulnerabilities() error
	GetAppInfo(appUIUrl, server string)
}

// Client is the unified Contrast API client for both sync and async operations
type Client struct {
	ApiKey     string
	ServiceKey string
	Username   string
	OrgID      string
	BaseURL    string
	AppURL     string
	Auth       string
	HttpClient *http.Client
}

// ReportStatusResponse represents the response from the report status endpoint
type ReportStatusResponse struct {
	Messages    []string `json:"messages"`
	Success     bool     `json:"success"`
	Status      string   `json:"status"`
	DownloadUrl string   `json:"downloadUrl,omitempty"`
}

// NewClient creates a new unified Contrast API client for both sync and async operations
type pollConfig struct {
	maxTotalWait    time.Duration
	maxPollInterval time.Duration
	initialDelay    time.Duration
	pollInterval    time.Duration
	backoffFactor   float64
}

func newPollConfig() pollConfig {
	return pollConfig{
		maxTotalWait:    5 * time.Minute,
		maxPollInterval: 60 * time.Second,
		initialDelay:    15 * time.Second,
		pollInterval:    5 * time.Second,
		backoffFactor:   1.5,
	}
}

func NewClient(apiKey, serviceKey, username, orgID, baseURL, appURL string) *Client {
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + serviceKey))
	return &Client{
		ApiKey:     apiKey,
		ServiceKey: serviceKey,
		Username:   username,
		OrgID:      orgID,
		BaseURL:    baseURL,
		AppURL:     appURL,
		Auth:       auth,
		HttpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// addAuth adds authentication headers to the request
func (c *Client) addAuth(req *http.Request) {
	req.SetBasicAuth(c.Username, c.ServiceKey)
	req.Header.Set("API-Key", c.ApiKey)
}

// Note: ContrastInstance is deprecated. Use the unified Client from client.go instead.
// The Client now supports both async (SARIF/PDF) and sync (vulnerabilities, app info) operations.
func getApplicationFromClient(client ContrastHttpClient, url string) (*ApplicationInfo, error) {
	var appResponse ApplicationResponse
	err := client.ExecuteRequest(url, nil, &appResponse)
	if err != nil {
		return nil, err
	}

	return &ApplicationInfo{
		Id:   appResponse.Id,
		Name: appResponse.Name,
	}, nil
}

func getVulnerabilitiesFromClient(client ContrastHttpClient, url string, page int) ([]ContrastFindings, error) {
	params := map[string]string{
		"page": fmt.Sprintf("%d", page),
		"size": fmt.Sprintf("%d", pageSize),
	}
	var vulnsResponse VulnerabilitiesResponse
	err := client.ExecuteRequest(url, params, &vulnsResponse)
	if err != nil {
		return nil, err
	}

	if vulnsResponse.Empty {
		log.Entry().Info("empty vulnerabilities response")
	}

	auditAllFindings, optionalFindings := getFindings(vulnsResponse.Vulnerabilities)

	if !vulnsResponse.Last {
		findings, err := getVulnerabilitiesFromClient(client, url, page+1)
		if err != nil {
			return nil, err
		}
		accumulateFindings(auditAllFindings, optionalFindings, findings)
		return findings, nil
	}
	return []ContrastFindings{auditAllFindings, optionalFindings}, nil
}

func getFindings(vulnerabilities []Vulnerability) (ContrastFindings, ContrastFindings) {
	var auditAllFindings, optionalFindings ContrastFindings
	auditAllFindings.ClassificationName = AuditAll
	auditAllFindings.Total = 0
	auditAllFindings.Audited = 0
	optionalFindings.ClassificationName = Optional
	optionalFindings.Total = 0
	optionalFindings.Audited = 0

	for _, vuln := range vulnerabilities {
		if vuln.Severity == Critical || vuln.Severity == High || vuln.Severity == Medium {
			if vuln.Status != StatusReported {
				auditAllFindings.Audited += 1
			}
			auditAllFindings.Total += 1
		} else {
			if vuln.Status != StatusReported {
				optionalFindings.Audited += 1
			}
			optionalFindings.Total += 1
		}
	}
	return auditAllFindings, optionalFindings
}

func accumulateFindings(auditAllFindings, optionalFindings ContrastFindings, contrastFindings []ContrastFindings) {
	for i, fr := range contrastFindings {
		if fr.ClassificationName == AuditAll {
			contrastFindings[i].Total += auditAllFindings.Total
			contrastFindings[i].Audited += auditAllFindings.Audited
		}
		if fr.ClassificationName == Optional {
			contrastFindings[i].Total += optionalFindings.Total
			contrastFindings[i].Audited += optionalFindings.Audited
		}
	}
}

func (c *Client) checkReportStatus(url string) (*ReportStatusResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set(ContentType, JSONContentType)
	c.addAuth(req)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Contrast API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var statusResp ReportStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &statusResp, nil
}

// PollReportStatus polls for report generation completion and returns the status response
func (c *Client) PollReportStatus(reportUuid, reportType string) (*ReportStatusResponse, error) {
	url := fmt.Sprintf("%s/Contrast/api/ng/organizations/%s/reports/%s/status", c.BaseURL, c.OrgID, reportUuid)
	config := newPollConfig()
	totalWaited := time.Duration(0)

	log.Entry().Infof("Waiting %v before first %s poll...", config.initialDelay, reportType)
	time.Sleep(config.initialDelay)
	totalWaited += config.initialDelay

	for totalWaited < config.maxTotalWait {
		statusResp, err := c.checkReportStatus(url)
		if err != nil {
			return nil, err
		}

		if !statusResp.Success {
			return nil, fmt.Errorf("%s status check failed: %v", reportType, statusResp.Messages)
		}

		log.Entry().Debugf("%s generation status: %s", reportType, statusResp.Status)

		if statusResp.Status == "ACTIVE" {
			log.Entry().Infof("%s report is ready for download", reportType)
			return statusResp, nil
		}

		if statusResp.Status != "CREATING" {
			return nil, fmt.Errorf("unexpected %s status: %s", reportType, statusResp.Status)
		}

		totalWaited, config.pollInterval = c.waitAndBackoff(totalWaited, config, reportType)
	}

	return nil, fmt.Errorf("%s generation timed out after waiting %s", reportType, config.maxTotalWait)
}

func (c *Client) waitAndBackoff(totalWaited time.Duration, config pollConfig, reportType string) (time.Duration, time.Duration) {
	log.Entry().Debugf("%s still generating, waiting %v...", reportType, config.pollInterval)
	time.Sleep(config.pollInterval)
	totalWaited += config.pollInterval

	nextInterval := time.Duration(float64(config.pollInterval) * config.backoffFactor)
	if nextInterval > config.maxPollInterval {
		return totalWaited, config.maxPollInterval
	}
	return totalWaited, nextInterval
}

// DownloadReport downloads a report from the given URL
func (c *Client) DownloadReport(downloadUrl, reportType string) ([]byte, error) {
	req, err := http.NewRequest("POST", downloadUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set(ContentType, JSONContentType)
	c.addAuth(req)

	log.Entry().Debugf("Downloading %s report...", reportType)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %w", reportType, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response: %w", reportType, err)
	}

	log.Entry().Infof("%s report downloaded successfully (%d bytes)", reportType, len(data))
	return data, nil
}

// GetVulnerabilities gets vulnerabilities for the application (synchronous)
func (c *Client) GetVulnerabilities() ([]ContrastFindings, error) {
	url := c.AppURL + "/vulnerabilities"
	httpClient := NewContrastHttpClient(c.ApiKey, c.Auth)

	return getVulnerabilitiesFromClient(httpClient, url, startPage)
}

// GetAppInfo gets application information (synchronous)
func (c *Client) GetAppInfo(appUIUrl, server string) (*ApplicationInfo, error) {
	httpClient := NewContrastHttpClient(c.ApiKey, c.Auth)
	app, err := getApplicationFromClient(httpClient, c.AppURL)
	if err != nil {
		log.Entry().Errorf("failed to get application from client: %v", err)
		return nil, err
	}
	app.Url = appUIUrl
	app.Server = server
	return app, nil
}

// AsyncReportConfig contains configuration for async report generation
type AsyncReportConfig struct {
	ReportType         string // "SARIF" or "PDF"
	URLPattern         string // URL pattern for starting async generation
	Payload            map[string]any
	DownloadURLPattern string // Pattern for building download URL
}

// generateAsyncReport is a generic function for async report generation (SARIF, PDF, etc.)
func (c *Client) generateAsyncReport(appUuid string, config AsyncReportConfig) ([]byte, error) {
	reportUuid, err := c.startAsyncReportGeneration(appUuid, config)
	if err != nil {
		return nil, fmt.Errorf("failed to start %s generation: %w", config.ReportType, err)
	}
	statusResp, err := c.PollReportStatus(reportUuid, config.ReportType)
	if err != nil {
		return nil, fmt.Errorf("failed to poll %s status: %w", config.ReportType, err)
	}
	var downloadUrl string
	if statusResp.DownloadUrl != "" {
		downloadUrl = statusResp.DownloadUrl
	} else {
		downloadUrl = fmt.Sprintf(config.DownloadURLPattern, c.BaseURL, c.OrgID, reportUuid)
	}

	if downloadUrl == "" {
		return nil, fmt.Errorf("%s download URL not provided", config.ReportType)
	}
	return c.DownloadReport(downloadUrl, config.ReportType)
}

// startAsyncReportGeneration starts an async report generation request
func (c *Client) startAsyncReportGeneration(appUuid string, config AsyncReportConfig) (string, error) {
	url := fmt.Sprintf(config.URLPattern, c.BaseURL, c.OrgID, appUuid)

	bodyBytes, err := json.Marshal(config.Payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(ContentType, JSONContentType)
	c.addAuth(req)

	log.Entry().Debugf("Starting async %s generation for application %s", config.ReportType, appUuid)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Contrast API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var reportResp struct {
		Messages []string `json:"messages"`
		Success  bool     `json:"success"`
		Uuid     string   `json:"uuid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&reportResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !reportResp.Success || reportResp.Uuid == "" {
		return "", fmt.Errorf("%s generation request failed: %v", config.ReportType, reportResp.Messages)
	}

	log.Entry().Infof("%s generation started with UUID: %s", config.ReportType, reportResp.Uuid)
	return reportResp.Uuid, nil
}

// GenerateSarifReport generates a SARIF report for the given application (start, poll, download)
func (c *Client) GenerateSarifReport(appUuid string) ([]byte, error) {
	config := AsyncReportConfig{
		ReportType: "SARIF",
		URLPattern: "%s/Contrast/api/ng/organizations/%s/applications/%s/sarif/async",
		Payload: map[string]any{
			"severities":  []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "NOTE"},
			"quickFilter": "OPEN",
			"toolTypes":   []string{"ASSESS"},
		},
		DownloadURLPattern: "",
	}
	return c.generateAsyncReport(appUuid, config)
}

// GeneratePdfReport generates a PDF attestation report for the given application (start, poll, download)
func (c *Client) GeneratePdfReport(appUuid string) ([]byte, error) {
	config := AsyncReportConfig{
		ReportType: "PDF",
		URLPattern: "%s/Contrast/api/ng/%s/applications/%s/attestation",
		Payload: map[string]any{
			"showVulnerabilitiesDetails": true,
			"showRouteObservations":      true,
		},
		DownloadURLPattern: "%s/Contrast/api/ng/%s/reports/%s/download",
	}
	return c.generateAsyncReport(appUuid, config)
}

// StartAsyncSarifGeneration initiates async SARIF report generation (wrapper for testing compatibility)
func (c *Client) StartAsyncSarifGeneration(appUuid string) (string, error) {
	config := AsyncReportConfig{
		ReportType: "SARIF",
		URLPattern: "%s/Contrast/api/ng/organizations/%s/applications/%s/sarif/async",
		Payload: map[string]any{
			"severities":  []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "NOTE"},
			"quickFilter": "OPEN",
			"toolTypes":   []string{"ASSESS"},
		},
		DownloadURLPattern: "",
	}
	return c.startAsyncReportGeneration(appUuid, config)
}

// StartAsyncPdfGeneration initiates async PDF report generation (wrapper for testing compatibility)
func (c *Client) StartAsyncPdfGeneration(appUuid string) (string, error) {
	config := AsyncReportConfig{
		ReportType: "PDF",
		URLPattern: "%s/Contrast/api/ng/%s/applications/%s/attestation",
		Payload: map[string]any{
			"showVulnerabilitiesDetails": true,
			"showRouteObservations":      true,
		},
		DownloadURLPattern: "%s/Contrast/api/ng/%s/reports/%s/download",
	}
	return c.startAsyncReportGeneration(appUuid, config)
}
