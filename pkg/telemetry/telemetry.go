package telemetry

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
)

const (
	eventType      = "library-os-ng"
	actionName     = "Piper Library OS"
	pipelineIDPath = ".pipeline/commonPipelineEnvironment/custom/cumulusPipelineID"
)

// LibraryRepository that is passed into with -ldflags
var LibraryRepository string

// Telemetry struct which holds necessary infos about telemetry
type Telemetry struct {
	baseData             BaseData
	data                 Data
	provider             orchestrator.ConfigProvider
	disabled             bool
	client               *piperhttp.Client
	CustomReportingDsn   string
	CustomReportingToken string
	customClient         *piperhttp.Client
	BaseURL              string
	Endpoint             string
	SiteID               string
	PendoToken           string
	Pendo                Pendo
}

type Pendo struct {
	Type       string `json:"type"`
	Event      string `json:"event"`
	VisitorID  string `json:"visitorId"`
	AccountID  string `json:"accountId"`
	Timestamp  int64  `json:"timestamp"`
	Properties *Data  `json:"properties"`
}

// Initialize sets up the base telemetry data and is called in generated part of the steps
func (t *Telemetry) Initialize(telemetryDisabled bool, stepName, token string) {
	if token == "" {
		telemetryDisabled = true
	}
	t.disabled = telemetryDisabled

	provider, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		log.Entry().Warningf("could not get orchestrator config provider, leads to insufficient data")
		provider = &orchestrator.UnknownOrchestratorConfigProvider{}
	}
	t.provider = provider

	if t.client == nil {
		t.client = &piperhttp.Client{}
	}

	t.client.SetOptions(piperhttp.ClientOptions{MaxRequestDuration: 5 * time.Second, MaxRetries: -1})

	if t.BaseURL == "" {
		// Pendo baseURL
		t.BaseURL = "https://app.pendo.io"
	}
	if t.Endpoint == "" {
		// Pendo endpoint
		t.Endpoint = "/data/track"
	}
	if len(LibraryRepository) == 0 {
		LibraryRepository = "https://github.com/n/a"
	}
	if t.SiteID == "" {
		t.SiteID = "827e8025-1e21-ae84-c3a3-3f62b70b0130"
	}

	t.PendoToken = token

	t.baseData = BaseData{
		Orchestrator:    t.provider.OrchestratorType(),
		StageName:       t.provider.StageName(),
		URL:             LibraryRepository,
		ActionName:      actionName,
		EventType:       eventType,
		StepName:        stepName,
		SiteID:          t.SiteID,
		PipelineURLHash: t.getPipelineURLHash(), // URL (hashed value) which points to the project’s pipelines
		BuildURLHash:    t.getBuildURLHash(),    // URL (hashed value) which points to the pipeline that is currently running
	}
}

func (t *Telemetry) getPipelineURLHash() string {
	jobURL := t.provider.JobURL()
	return t.toSha1OrNA(jobURL)
}

func (t *Telemetry) getBuildURLHash() string {
	buildURL := t.provider.BuildURL()
	return t.toSha1OrNA(buildURL)
}

func (t *Telemetry) toSha1OrNA(input string) string {
	if len(input) == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%x", sha1.Sum([]byte(input)))
}

// SetData sets the custom telemetry, Pendo and base data
func (t *Telemetry) SetData(customData *CustomData) {
	t.data = Data{
		BaseData:   t.baseData,
		CustomData: *customData,
	}
	pipelineID := readPipelineID(pipelineIDPath)
	t.Pendo = Pendo{
		Type:       "track",
		Event:      t.baseData.StepName,
		AccountID:  pipelineID,
		VisitorID:  pipelineID,
		Timestamp:  time.Now().UnixMilli(),
		Properties: &t.data,
	}
}

// GetData returns telemetryData
func (t *Telemetry) GetData() Data {
	return t.data
}

func (t *Telemetry) GetDataBytes() []byte {
	data, err := json.Marshal(t.data)
	if err != nil {
		log.Entry().WithError(err).Println("Failed to marshal data")
		return []byte{}
	}

	return data
}

// Send telemetry information to SWA
func (t *Telemetry) Send() {
	// always log step telemetry data to logfile used for internal use-case
	t.logStepTelemetryData()

	// skip if telemetry is disabled
	if t.disabled {
		return
	}

	b, err := json.Marshal(t.Pendo)
	if err != nil {
		log.Entry().WithError(err).Println("Failed to marshal data")
		return
	}

	log.Entry().Debug("Sending telemetry data")
	h := http.Header{}
	h.Add("Content-Type", "application/json")
	h.Add("X-Pendo-Integration-Key", t.PendoToken)
	t.client.SendRequest(http.MethodPost, t.BaseURL+t.Endpoint, bytes.NewReader(b), h, nil)
}

func (t *Telemetry) logStepTelemetryData() {

	var fatalError map[string]interface{}
	if t.data.CustomData.ErrorCode != "0" && log.GetFatalErrorDetail() != nil {
		// retrieve the error information from the logCollector
		err := json.Unmarshal(log.GetFatalErrorDetail(), &fatalError)
		if err != nil {
			log.Entry().WithError(err).Warn("could not unmarshal fatal error struct")
		}
	}

	// Subtracts the duration from now to estimate the step start time
	i, err := strconv.ParseInt(t.data.CustomData.Duration, 10, 64)
	duration := time.Millisecond * time.Duration(i)
	starTime := time.Now().UTC().Add(-duration)

	stepTelemetryData := StepTelemetryData{
		StepStartTime:   starTime.String(),
		PipelineURLHash: t.data.PipelineURLHash,
		BuildURLHash:    t.data.BuildURLHash,
		StageName:       t.data.StageName,
		StepName:        t.data.BaseData.StepName,
		ErrorCode:       t.data.CustomData.ErrorCode,
		StepDuration:    t.data.CustomData.Duration,
		ErrorCategory:   t.data.CustomData.ErrorCategory,
		ErrorDetail:     fatalError,
		CorrelationID:   t.provider.BuildURL(),
		PiperCommitHash: t.data.CustomData.PiperCommitHash,
	}
	stepTelemetryJSON, err := json.Marshal(stepTelemetryData)
	if err != nil {
		log.Entry().Error("could not marshal step telemetry data")
		log.Entry().Infof("Step telemetry data: {n/a}")
	} else {
		// log step telemetry data, changes here need to change the regex in the internal piper lib
		log.Entry().Infof("Step telemetry data:%v", string(stepTelemetryJSON))
	}
}

func readPipelineID(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Entry().Debugf("Could not read %v file: %v", filePath, err)
		content = []byte("N/A")
	}
	return string(content)
}
