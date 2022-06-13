package telemetry

import (
	"encoding/json"
	"net/url"
)

// BaseData object definition containing the base data and it's mapping information
type BaseData struct {
	// SWA receives the fields custom1 - custom30 and e_a, e_2 - e_30 for custom values.
	ActionName      string `json:"action_name"`
	EventType       string `json:"event_type"`
	SiteID          string `json:"idsite"`
	URL             string `json:"url"`
	StepName        string `json:"e_3"` // set by step generator
	StageName       string `json:"e_10"`
	PipelineURLHash string `json:"e_4"`  // defaults to sha1 of provider.GetBuildURL()
	BuildURLHash    string `json:"e_5"`  // defaults to sha1 of provider.GetJobURL()
	Orchestrator    string `json:"e_14"` // defaults to provider.OrchestratorType()
}

var baseData BaseData

// BaseMetaData object definition containing the labels for the base data, and it's mapping information
type BaseMetaData struct {
	// SWA receives the fields custom1 - custom30 and e_a, e_2 - e_30 for custom values.
	StepNameLabel        string `json:"custom3"`
	StageNameLabel       string `json:"custom10"`
	PipelineURLHashLabel string `json:"custom4"`
	BuildURLHashLabel    string `json:"custom5"`
	DurationLabel        string `json:"custom11,omitempty"`
	ExitCodeLabel        string `json:"custom12,omitempty"`
	ErrorCategoryLabel   string `json:"custom13,omitempty"`
	OrchestratorLabel    string `json:"custom14,omitempty"`
	PiperCommitHashLabel string `json:"custom15,omitempty"`
}

// baseMetaData object containing the labels for the base data
var baseMetaData BaseMetaData = BaseMetaData{
	StepNameLabel:        "stepName",
	StageNameLabel:       "stageName",
	PipelineURLHashLabel: "pipelineUrlHash",
	BuildURLHashLabel:    "buildUrlHash",
	DurationLabel:        "duration",
	ExitCodeLabel:        "exitCode",
	ErrorCategoryLabel:   "errorCategory",
	OrchestratorLabel:    "orchestrator",
	PiperCommitHashLabel: "piperCommitHash",
}

// CustomData object definition containing the data that can be set by a step, and it's mapping information
type CustomData struct {
	// SWA receives the fields custom1 - custom30 and e_a, e_2 - e_30 for custom values.
	// Piper uses the values custom11 - custom25 & e_11 - e_25 for library related reporting
	// and custom26 - custom30 & e_26 - e_30 for step  related reporting.
	Duration        string `json:"e_11,omitempty"`
	ErrorCode       string `json:"e_12,omitempty"`
	ErrorCategory   string `json:"e_13,omitempty"`
	PiperCommitHash string `json:"e_15,omitempty"`
	Custom1Label    string `json:"custom26,omitempty"`
	Custom2Label    string `json:"custom27,omitempty"`
	Custom3Label    string `json:"custom28,omitempty"`
	Custom4Label    string `json:"custom29,omitempty"`
	Custom5Label    string `json:"custom30,omitempty"`
	Custom1         string `json:"e_26,omitempty"`
	Custom2         string `json:"e_27,omitempty"`
	Custom3         string `json:"e_28,omitempty"`
	Custom4         string `json:"e_29,omitempty"`
	Custom5         string `json:"e_30,omitempty"`
}

// StepTelemetryData definition for telemetry reporting and monitoring
type StepTelemetryData struct {
	StepStartTime   string                 `json:"StepStartTime"`
	PipelineURLHash string                 `json:"PipelineURLHash"`
	BuildURLHash    string                 `json:"BuildURLHash"`
	StageName       string                 `json:"StageName"`
	StepName        string                 `json:"StepName"`
	ErrorCode       string                 `json:"ErrorCode"`
	StepDuration    string                 `json:"StepDuration"`
	ErrorCategory   string                 `json:"ErrorCategory"`
	CorrelationID   string                 `json:"CorrelationID"`
	PiperCommitHash string                 `json:"PiperCommitHash"`
	ErrorDetail     map[string]interface{} `json:"ErrorDetail"`
}

// Data object definition containing all telemetry data
type Data struct {
	BaseData
	BaseMetaData
	CustomData
}

// toMap transfers the data object into a map using JSON tags
func (d *Data) toMap() (result map[string]string) {
	jsonObj, _ := json.Marshal(d)
	json.Unmarshal(jsonObj, &result)
	return
}

// toPayloadString transfers the data object into a 'key=value&..' string
func (d *Data) toPayloadString() string {
	parameters := url.Values{}

	for key, value := range d.toMap() {
		if len(value) > 0 {
			parameters.Add(key, value)
		}
	}

	return parameters.Encode()
}
