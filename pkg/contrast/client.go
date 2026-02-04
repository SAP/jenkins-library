package contrast

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

// Client is the Contrast API client for report generation
type Client struct {
	ApiKey     string
	ServiceKey string
	Username   string
	OrgID      string
	BaseURL    string
	HttpClient *http.Client
}

// ReportStatusResponse represents the response from the report status endpoint
type ReportStatusResponse struct {
	Messages    []string `json:"messages"`
	Success     bool     `json:"success"`
	Status      string   `json:"status"`
	DownloadUrl string   `json:"downloadUrl,omitempty"`
}

// NewClient creates a new Contrast API client
func NewClient(apiKey, serviceKey, username, orgID, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://cs003.contrastsecurity.com"
	}
	return &Client{
		ApiKey:     apiKey,
		ServiceKey: serviceKey,
		Username:   username,
		OrgID:      orgID,
		BaseURL:    baseURL,
		HttpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// addAuth adds authentication headers to the request
func (c *Client) addAuth(req *http.Request) {
	req.SetBasicAuth(c.Username, c.ServiceKey)
	req.Header.Set("API-Key", c.ApiKey)
}

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

func (c *Client) checkReportStatus(url string) (*ReportStatusResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
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
	req.Header.Set("Content-Type", "application/json")
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
