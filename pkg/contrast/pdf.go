package contrast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/log"
)

// StartAsyncPdfGeneration initiates async PDF report generation for the given application
func (c *Client) StartAsyncPdfGeneration(appUuid string) (string, error) {
	url := fmt.Sprintf("%s/Contrast/api/ng/%s/applications/%s/attestation", c.BaseURL, c.OrgID, appUuid)
	body := map[string]bool{
		"showVulnerabilitiesDetails": true,
		"showRouteObservations":      true,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.addAuth(req)

	log.Entry().Debugf("Starting async PDF generation for application %s", appUuid)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Contrast API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var pdfResp struct {
		Messages []string `json:"messages"`
		Success  bool     `json:"success"`
		Uuid     string   `json:"uuid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pdfResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !pdfResp.Success || pdfResp.Uuid == "" {
		return "", fmt.Errorf("PDF generation request failed: %v", pdfResp.Messages)
	}

	log.Entry().Infof("PDF generation started with UUID: %s", pdfResp.Uuid)
	return pdfResp.Uuid, nil
}

// GeneratePdfReport generates a PDF attestation report for the given application (start, poll, download)
func (c *Client) GeneratePdfReport(appUuid string) ([]byte, error) {
	reportUuid, err := c.StartAsyncPdfGeneration(appUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to start PDF generation: %w", err)
	}
	_, err = c.PollReportStatus(reportUuid, "PDF")
	if err != nil {
		return nil, fmt.Errorf("failed to poll PDF status: %w", err)
	}
	downloadUrl := fmt.Sprintf("%s/Contrast/api/ng/%s/reports/%s/download", c.BaseURL, c.OrgID, reportUuid)
	return c.DownloadReport(downloadUrl, "PDF")
}
