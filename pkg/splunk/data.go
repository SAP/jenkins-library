package splunk

import "github.com/SAP/jenkins-library/pkg/log"

type Event struct {
	Messages  []log.Message  `json:"messages,omitempty"`  // messages
	Telemetry MonitoringData `json:"telemetry,omitempty"` // telemetryData
}
type Details struct {
	Host       string `json:"host"`                 // hostname
	Source     string `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      Event  `json:"event,omitempty"`      // throw any useful key/val pairs here}
}

// MonitoringData definition for monitoring
type MonitoringData struct {
	PipelineUrlHash string `json:"PipelineUrlHash,omitempty"`
	BuildUrlHash    string `json:"BuildUrlHash,omitempty"`
	StageName       string `json:"StageName,omitempty"`
	StepName        string `json:"StepName,omitempty"`
	ExitCode        string `json:"ExitCode,omitempty"`
	Duration        string `json:"Duration,omitempty"`
	ErrorCode       string `json:"ErrorCode,omitempty"`
	ErrorCategory   string `json:"ErrorCategory,omitempty"`
	CorrelationID   string `json:"CorrelationID,omitempty"`
	CommitHash      string `json:"CommitHash,omitempty"`
	Branch          string `json:"Branch,omitempty"`
	GitOwner        string `json:"GitOwner,omitempty"`
	GitRepository   string `json:"GitRepository,omitempty"`
}

type LogFileEvents struct {
	Messages  []string          `json:"messages,omitempty"`  // messages
	Telemetry PipelineTelemetry `json:"telemetry,omitempty"` // telemetryData
}
type DetailsLog struct {
	Host       string        `json:"host"`                 // hostname
	Source     string        `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string        `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string        `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      LogFileEvents `json:"event,omitempty"`      // throw any useful key/val pairs here}
}

type DetailsTelemetry struct {
	Host       string            `json:"host"`                 // hostname
	Source     string            `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string            `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string            `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      PipelineTelemetry `json:"event,omitempty"`      // throw any useful key/val pairs here}
}

type StepData struct {
}

// PipelineTelemetry object to store pipeline related telemetry information
type PipelineTelemetry struct {
	CorrelationId       string             `json:"CorrelationId"`       // CorrelationId
	Duration            string             `json:"PipelineDuration"`    // Duration of the pipeline in milliseconds
	Orchestrator        string             `json:"Orchestrator"`        // Orchestrator, e.g. Jenkins or Azure
	OrchestratorVersion string             `json:"OrchestratorVersion"` // OrchestratorVersion
	PipelineStartTime   string             `json:"PipelineStartTime"`   // PipelineStartTime Pipeline start time
	Errors              []log.ErrorDetails `json:"ErrorDetails"`        // Errors Error Codes found in errorsJson
	BuildId             string             `json:"BuildId"`             // BuildId of the pipeline run
	JobName             string             `json:"JobName"`
	PipelineStatus      string             `json:"PipelineStatus"`
	PipelineName        string             `json:"PipelineName"`
	GitInstance         string             `json:"GitInstance"`
	JobURL              string             `json:"JobURL"`
	CommitHash          string             `json:"CommitHash"`
	Branch              string             `json:"Branch"`
	GitOwner            string             `json:"GitOwner"`
	GitRepository       string             `json:"GitRepository"`
	StepData            []StepData         `json:"StepData"`
}
