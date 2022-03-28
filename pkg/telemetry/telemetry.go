package telemetry

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"strconv"
	"time"

	"net/http"
	"net/url"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

// eventType
const eventType = "library-os-ng"

// actionName
const actionName = "Piper Library OS"

// LibraryRepository that is passed into with -ldflags
var LibraryRepository string

// Telemetry struct which holds necessary infos about telemetry
type Telemetry struct {
	baseData             BaseData
	baseMetaData         BaseMetaData
	data                 Data
	provider             orchestrator.OrchestratorSpecificConfigProviding
	disabled             bool
	client               *piperhttp.Client
	CustomReportingDsn   string
	CustomReportingToken string
	customClient         *piperhttp.Client
	BaseURL              string
	Endpoint             string
	SiteID               string
}

// Initialize sets up the base telemetry data and is called in generated part of the steps
func (t *Telemetry) Initialize(telemetryDisabled bool, stepName string) {
	t.disabled = telemetryDisabled

	provider, err := orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil || provider == nil {
		log.Entry().Warningf("could not get orchestrator config provider, leads to insufficient data")
		provider = &orchestrator.UnknownOrchestratorConfigProvider{}
	}
	t.provider = provider

	if t.client == nil {
		t.client = &piperhttp.Client{}
	}

	t.client.SetOptions(piperhttp.ClientOptions{MaxRequestDuration: 5 * time.Second, MaxRetries: -1})

	if t.BaseURL == "" {
		//SWA baseURL
		t.BaseURL = "https://webanalytics.cfapps.eu10.hana.ondemand.com"
	}
	if t.Endpoint == "" {
		// SWA endpoint
		t.Endpoint = "/tracker/log"
	}
	if len(LibraryRepository) == 0 {
		LibraryRepository = "https://github.com/n/a"
	}

	if t.SiteID == "" {
		t.SiteID = "827e8025-1e21-ae84-c3a3-3f62b70b0130"
	}

	t.baseData = BaseData{
		Orchestrator:    provider.OrchestratorType(),
		StageName:       provider.GetStageName(),
		URL:             LibraryRepository,
		ActionName:      actionName,
		EventType:       eventType,
		StepName:        stepName,
		SiteID:          t.SiteID,
		PipelineURLHash: t.getPipelineURLHash(), // http://server:port/jenkins/job/foo/
		BuildURLHash:    t.getBuildURLHash(),    // http://server:port/jenkins/job/foo/15/
	}
	t.baseMetaData = baseMetaData
}

func (t *Telemetry) getPipelineURLHash() string {
	jobURL := t.provider.GetJobURL()
	return t.toSha1OrNA(jobURL)
}

func (t *Telemetry) getBuildURLHash() string {
	buildURL := t.provider.GetBuildURL()
	return t.toSha1OrNA(buildURL)
}

func (t *Telemetry) toSha1OrNA(input string) string {
	if len(input) == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%x", sha1.Sum([]byte(input)))
}

// SetData sets the custom telemetry data and base data into the Data object
func (t *Telemetry) SetData(customData *CustomData) {
	t.data = Data{
		BaseData:     t.baseData,
		BaseMetaData: t.baseMetaData,
		CustomData:   *customData,
	}
}

// GetData returns telemetryData
func (t *Telemetry) GetData() Data {
	return t.data
}

// Send telemetry information to SWA
func (t *Telemetry) Send() {
	// always log step telemetry data to logfile used for internal use-case
	t.logStepTelemetryData()

	// skip if telemetry is disabled
	if t.disabled {
		return
	}

	request, _ := url.Parse(t.BaseURL)
	request.Path = t.Endpoint
	request.RawQuery = t.data.toPayloadString()
	log.Entry().WithField("request", request.String()).Debug("Sending telemetry data")
	t.client.SendRequest(http.MethodGet, request.String(), nil, nil, nil)
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
		CorrelationID:   t.provider.GetBuildURL(),
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
