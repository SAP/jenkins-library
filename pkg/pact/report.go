package pact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Metrics represents ci report metrics which will makes up part of the report sent to the ci report server
type Metrics struct {
	Type    string   `json:"type"`
	Title   string   `json:"title"`
	Metrics []Metric `json:"metrics"`
}

// Metric represents a single metric which is sent to the report server
type Metric struct {
	Text  string `json:"text"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Level string `json:"level"`
	Link  string `json:"link"`
}

// Report represents the report that is uploaded to the ci report server
type Report struct {
	Data    *ReportData `json:"data"`
	Metrics []Metrics   `json:"metrics"`
}

type ReportData struct {
	OrgOrigin   string `json:"org_origin"`
	OrgAlias    string `json:"org_alias"`
	GitProvider string `json:"git_provider"`
	GitRepo     string `json:"git_repo"`
	GitCommit   string `json:"git_commit"`
	GitPullID   string `json:"git_pull_id"`
	BuildID     string `json:"build_id"`
	GitBranch   string `json:"git_branch"`
}

// ReportClient represents a connection to the ci report server
type ReportClient struct {
	host string
}

// NewReportClient accepts in as an argument systemNamespace. It initializes and returns a ReportClient.
func NewReportClient(systemNamespace string) *ReportClient {
	return &ReportClient{host: fmt.Sprintf("http://dev-ci-report.%s", systemNamespace)}
}

// SendReport sends report to ci report server.
// It returns any errors if encountered.
func (rc *ReportClient) SendReport(reportData *ReportData, text, name, value string, utils Utils) error {
	// Create report and send to CI report server
	report := &Report{
		Data:    reportData,
		Metrics: []Metrics{},
	}
	metrics := Metrics{
		Type:  "contract_tests",
		Title: "",
		Metrics: []Metric{
			{
				Text:  text,
				Name:  name,
				Value: value,
			},
		},
	}
	report.Metrics = append(report.Metrics, metrics)

	url := fmt.Sprintf("%s/api/report", rc.host)
	reportBytes, err := json.Marshal(report)
	if err != nil {
		return err
	}
	resp, err := sendRequest(http.MethodPost, url, bytes.NewReader(reportBytes), utils)
	if err != nil {
		return err
	}
	fmt.Printf("upload contract tests metric response: %s", string(resp))
	return nil
}