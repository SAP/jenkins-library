package contrast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/log"
)

// StartAsyncSarifGeneration initiates async SARIF report generation for the given application
func (c *Client) StartAsyncSarifGeneration(appUuid string) (string, error) {
	url := fmt.Sprintf("%s/Contrast/api/ng/organizations/%s/applications/%s/sarif/async",
		c.BaseURL, c.OrgID, appUuid)

	payload := map[string]interface{}{
		"severities":  []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "NOTE"},
		"quickFilter": "OPEN",
		"toolTypes":   []string{"ASSESS"},
	}
	bodyBytes, _ := json.Marshal(payload)
	body := bytes.NewBuffer(bodyBytes)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.addAuth(req)

	log.Entry().Debugf("Starting async SARIF generation for application %s", appUuid)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Contrast API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var sarifResp struct {
		Messages []string `json:"messages"`
		Success  bool     `json:"success"`
		Uuid     string   `json:"uuid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sarifResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !sarifResp.Success || sarifResp.Uuid == "" {
		return "", fmt.Errorf("SARIF generation request failed: %v", sarifResp.Messages)
	}

	log.Entry().Infof("SARIF generation started with UUID: %s", sarifResp.Uuid)
	return sarifResp.Uuid, nil
}

// GenerateSarifReport generates a SARIF report for the given application (start, poll, download)
func (c *Client) GenerateSarifReport(appUuid string) ([]byte, error) {
	// Start async generation
	reportUuid, err := c.StartAsyncSarifGeneration(appUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to start SARIF generation: %w", err)
	}
	statusResp, err := c.PollReportStatus(reportUuid, "SARIF")
	if err != nil {
		return nil, fmt.Errorf("failed to poll SARIF status: %w", err)
	}
	if statusResp.DownloadUrl == "" {
		return nil, fmt.Errorf("SARIF download URL not provided")
	}
	return c.DownloadReport(statusResp.DownloadUrl, "SARIF")
}
