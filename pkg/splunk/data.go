package splunk

import "github.com/SAP/jenkins-library/pkg/log"

type Event struct {
	Messages  []log.Message  `json:"messages,omitempty"` // messages
	Telemetry MonitoringData `json:"telemetry"`          // telemetryData
}
type Details struct {
	Host       string `json:"host"`                 // hostname
	Source     string `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      Event  `json:"event"`                // throw any useful key/val pairs here}
}

// MonitoringData definition for monitoring
type MonitoringData struct {
	PipelineUrlHash   string `json:"PipelineUrlHash,omitempty"`
	BuildUrlHash      string `json:"BuildUrlHash,omitempty"`
	Orchestrator      string `json:"Orchestrator,omitempty"`
	TemplateName      string `json:"TemplateName,omitempty"`
	StageTemplateName string `json:"StageTemplateName,omitempty"`
	PiperCommitHash   string `json:"PiperCommitHash,omitempty"`
	StageName         string `json:"StageName,omitempty"`
	StepName          string `json:"StepName,omitempty"`
	ExitCode          string `json:"ExitCode,omitempty"`
	Duration          string `json:"Duration,omitempty"`
	ErrorCode         string `json:"ErrorCode,omitempty"`
	ErrorCategory     string `json:"ErrorCategory,omitempty"`
	ErrorMessage      string `json:"ErrorMessage,omitempty"`
	CorrelationID     string `json:"CorrelationId,omitempty"`
	CommitHash        string `json:"CommitHash,omitempty"`
	Branch            string `json:"Branch,omitempty"`
	GitOwner          string `json:"GitOwner,omitempty"`
	GitRepository     string `json:"GitRepository,omitempty"`
	BinaryVersion     string `json:"BinaryVersion,omitempty"`
	ActionVersion     string `json:"ActionVersion,omitempty"`
	TemplateVersion   string `json:"TemplateVersion,omitempty"`
}

type LogFileEvent struct {
	Event      string `json:"event"`      // messages
	Host       string `json:"host"`       // hostname
	Source     string `json:"source"`     // optional description of the source of the event; typically the app's name
	SourceType string `json:"sourcetype"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string `json:"index"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
}

type DetailsTelemetry struct {
	Host       string         `json:"host"`                 // hostname
	Source     string         `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string         `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string         `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      map[string]any `json:"event,omitempty"`      // throw any useful key/val pairs here}
}
