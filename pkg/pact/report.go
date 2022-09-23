package pact

import (
	"encoding/json"
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

// SaveReport stores a report in the workspace.
// It returns any errors if encountered.
func (r *Report) SaveReport(reportData *ReportData, filePath, text, name, value string, utils Utils) error {
	r.Data =    reportData
	r.Metrics = []Metrics{
		{
			Type:  "contract_tests",
			Title: "",
			Metrics: []Metric{
				{
					Text:  text,
					Name:  name,
					Value: value,
				},
			},
		},
	}
	
	reportBytes, err := json.Marshal(r)
	if err != nil {
		return err
	}

	err = utils.WriteFile(filePath, reportBytes, 0666)
	if err != nil {
		return err
	}

	return nil
}
