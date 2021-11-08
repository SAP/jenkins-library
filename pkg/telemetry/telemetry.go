package telemetry

import (
	"crypto/sha1"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
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

// SiteID ...
var SiteID string

var disabled bool
var client piperhttp.Sender

// SWA baseURL
const baseURL = "https://webanalytics.cfapps.eu10.hana.ondemand.com"

// SWA endpoint
const endpoint = "/tracker/log"

// Telemetry struct which holds necessary infos about telemetry
type Telemetry struct {
	baseData     BaseData
	baseMetaData BaseMetaData
	data         Data
	provider     orchestrator.OrchestratorSpecificConfigProviding
	disabled     bool
}

// Initialize sets up the base telemetry data and is called in generated part of the steps
func (t *Telemetry) Initialize(telemetryDisabled bool, stepName string) {
	t.disabled = telemetryDisabled
	// skip if telemetry is disabled
	if telemetryDisabled {
		log.Entry().Info("Telemetry reporting deactivated")
		return
	}
	provider, err := orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil || provider == nil {
		log.Entry().Warningf("could not get orchestrator config provider, leads to insufficient data")
		provider = &orchestrator.UnknownOrchestratorConfigProvider{}
	}
	t.provider = provider

	if client == nil {
		client = &piperhttp.Client{}
	}

	client.SetOptions(piperhttp.ClientOptions{MaxRequestDuration: 5 * time.Second, MaxRetries: -1})

	if len(LibraryRepository) == 0 {
		LibraryRepository = "https://github.com/n/a"
	}

	if len(SiteID) == 0 {
		SiteID = "827e8025-1e21-ae84-c3a3-3f62b70b0130"
	}

	t.baseData = BaseData{
		Orchestrator:    provider.OrchestratorType(),
		StageName:       provider.GetStageName(),
		URL:             LibraryRepository,
		ActionName:      actionName,
		EventType:       eventType,
		StepName:        stepName,
		SiteID:          SiteID,
		PipelineURLHash: t.getPipelineURLHash(), // http://server:port/jenkins/job/foo/
		BuildURLHash:    t.getBuildURLHash(),    // http://server:port/jenkins/job/foo/15/
	}
	t.baseMetaData = baseMetaData
}

func (t *Telemetry) getPipelineURLHash() string {
	jobUrl := t.provider.GetJobUrl()
	return t.toSha1OrNA(jobUrl)
}

func (t *Telemetry) getBuildURLHash() string {
	buildUrl := t.provider.GetBuildUrl()
	return t.toSha1OrNA(buildUrl)
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

	// skip if telemetry is disabled
	if t.disabled {
		return
	}

	request, _ := url.Parse(baseURL)
	request.Path = endpoint
	request.RawQuery = t.data.toPayloadString()
	log.Entry().WithField("request", request.String()).Debug("Sending telemetry data")
	client.SendRequest(http.MethodGet, request.String(), nil, nil, nil)
}
