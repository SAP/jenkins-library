package telemetry

import (
	"encoding/json"
	"net/url"
)

// BaseData object definition containing the base data
type BaseData struct {
	ActionName      string `json:"actionName"`
	EventType       string `json:"eventType"`
	SiteID          string `json:"idsite"`
	URL             string `json:"url"`
	StepName        string `json:"stepName"` // set by step generator
	StageName       string `json:"stageName"`
	PipelineURLHash string `json:"pipelineUrlHash"` // defaults to sha1 of provider.GetBuildURL()
	BuildURLHash    string `json:"buildUrlHash"`    // defaults to sha1 of provider.GetJobURL()
	Orchestrator    string `json:"orchestrator"`    // defaults to provider.OrchestratorType()
}

var baseData BaseData

// CustomData object definition containing the data that can be set by a step
type CustomData struct {
	Duration              string `json:"duration,omitempty"`
	ErrorCode             string `json:"errorCode,omitempty"`
	ErrorCategory         string `json:"errorCategory,omitempty"`
	PiperCommitHash       string `json:"piperCommitHash,omitempty"`
	BuildTool             string `json:"buildTool,omitempty"`
	FilePath              string `json:"filePath,omitempty"`
	DeployTool            string `json:"deployTool,omitempty"`
	ContainerBuildOptions string `json:"containerBuildOptions,omitempty"`
	IsScheduled           bool   `json:"isScheduled,omitempty"`
	IsOptimized           bool   `json:"isOptimized,omitempty"`
	ProxyLogFile          string `json:"proxyLogFile,omitempty"`
	BuildType             string `json:"buildType,omitempty"`
	BuildQuality          string `json:"buildQuality,omitempty"`
	LegacyJobNameTemplate string `json:"legacyJobNameTemplate,omitempty"`
	LegacyJobName         string `json:"legacyJobName,omitempty"`
	DeployType            string `json:"deployType,omitempty"`
	CnbBuilder            string `json:"cnbBuilder,omitempty"`
	CnbRunImage           string `json:"cnbRunImage,omitempty"`
	ServerURL             string `json:"serverURL,omitempty"`
	ECCNMessageStatus     string `json:"eccnMessageStatus,omitempty"`
	ChangeRequestUpload   string `json:"changeRequestUpload,omitempty"`
	BuildVersionCreation  string `json:"buildVersionCreation,omitempty"`
	PullRequestMode       string `json:"pullRequestMode,omitempty"`
	GroovyTemplateUsed    string `json:"groovyTemplateUsed,omitempty"`
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
