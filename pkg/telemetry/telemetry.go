package telemetry

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
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
	client               *piperhttp.Client
	CustomReportingDsn   string
	CustomReportingToken string
	customClient         *piperhttp.Client
	BaseURL              string
	Endpoint             string
	SiteID               string
}

// Initialize sets up the base telemetry data and is called in generated part of the steps
func (t *Telemetry) Initialize(stepName string) {
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

	if len(LibraryRepository) == 0 {
		LibraryRepository = "https://github.com/n/a"
	}
	if t.SiteID == "" {
		t.SiteID = "827e8025-1e21-ae84-c3a3-3f62b70b0130"
	}

	t.baseData = BaseData{
		Orchestrator:      t.provider.OrchestratorType(),
		TemplateName:      piperutils.StringWithDefault(os.Getenv("PIPER_PIPELINE_TEMPLATE_NAME"), "n/a"),
		StageTemplateName: piperutils.StringWithDefault(os.Getenv("PIPER_PIPELINE_STAGE_TEMPLATE_NAME"), "n/a"),
		StageName:         t.provider.StageName(),
		URL:               LibraryRepository,
		ActionName:        actionName,
		EventType:         eventType,
		StepName:          stepName,
		SiteID:            t.SiteID,
		PipelineURLHash:   t.getPipelineURLHash(), // URL (hashed value) which points to the project’s pipelines
		BuildURLHash:      t.getBuildURLHash(),    // URL (hashed value) which points to the pipeline that is currently running
		BinaryVersion:     piperutils.GetVersion(),
		ActionVersion:     piperutils.StringWithDefault(os.Getenv("PIPER_ACTION_VERSION"), "n/a"),
		TemplateVersion:   piperutils.StringWithDefault(os.Getenv("PIPER_TEMPLATE_VERSION"), "n/a"),
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

// SetData sets the custom telemetry and base data
func (t *Telemetry) SetData(customData *CustomData) {
	t.data = Data{
		BaseData:   t.baseData,
		CustomData: *customData,
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

// Logs step telemetry data to logfile used for internal use-case
func (t *Telemetry) LogStepTelemetryData() {
	log.Entry().Debug("Logging step telemetry data")

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
		BinaryVersion:   t.data.BinaryVersion,
		ActionVersion:   t.data.ActionVersion,
		TemplateVersion: t.data.TemplateVersion,
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
