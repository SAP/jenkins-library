package telemetry

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
	BuildTool             string `json:"buildTool,omitempty"` // artifactPrepVersion
	FilePath              string `json:"filePath,omitempty"`
	DeployTool            string `json:"deployTool,omitempty"`            // k8sDeploy
	ContainerBuildOptions string `json:"containerBuildOptions,omitempty"` // kaniko
	IsScheduled           bool   `json:"isScheduled,omitempty"`           // sapInit
	IsOptimized           bool   `json:"isOptimized,omitempty"`           // sapInit
	ProxyLogFile          string `json:"proxyLogFile,omitempty"`          // xmake
	BuildType             string `json:"buildType,omitempty"`
	BuildQuality          string `json:"buildQuality,omitempty"`
	LegacyJobNameTemplate string `json:"legacyJobNameTemplate,omitempty"`
	LegacyJobName         string `json:"legacyJobName,omitempty"`
	DeployType            string `json:"deployType,omitempty"`
	CnbBuildStepData      string `json:"cnbBuildStepData,omitempty"`
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
