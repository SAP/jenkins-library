package telemetry

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"io"
	"io/ioutil"
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
	PipelineTelemetry    *PipelineTelemetry
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

	if t.data.ErrorCode == "1" && len(t.CustomReportingDsn) > 0 {
		// check if reporting dsn is available in the config and then send the payload to the specific URL
		log.Entry().Infof("Step %v exited with errorcode 1, sending additional reporting infos to %v", t.data.StepName, t.CustomReportingDsn)
		t.sendCustom()
	}

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

func (t *Telemetry) sendCustom() {
	if t.customClient == nil {
		t.customClient = &piperhttp.Client{}
	}
	t.provider.GetBuildUrl()
	data := t.data
	data.BuildURLHash = t.provider.GetBuildUrl()
	data.PipelineURLHash = t.provider.GetBuildUrl()

	var payload []byte
	var err error
	if t.PipelineTelemetry != nil {
		var m map[string]string

		jPipelineTelemetry, _ := json.Marshal(t.PipelineTelemetry)
		json.Unmarshal(jPipelineTelemetry, &m)
		jData, _ := json.Marshal(data)
		json.Unmarshal(jData, &m)

		payload, err = json.Marshal(m)
		if err != nil {
			log.Entry().Errorf("error while marshalling reporting details, %v", err)
		}
	} else {
		payload, err = json.Marshal(data)
		if err != nil {
			log.Entry().Errorf("error while marshalling reporting details, %v", err)
		}
	}

	log.Entry().Debugf("Sending the following payload: %v", string(payload))
	resp, err := t.customClient.SendRequest(http.MethodPost, t.CustomReportingDsn, bytes.NewBuffer(payload), nil, nil)

	if resp != nil {
		if resp.StatusCode != http.StatusOK {
			// log it to stdout
			rdr := io.LimitReader(resp.Body, 1000)
			body, errRead := ioutil.ReadAll(rdr)
			log.Entry().Infof("%v: error logging failed - %v", resp.Status, string(body))
			if errRead != nil {
				log.Entry().Errorf("Error reading response body. %v", errRead)
			}
			log.Entry().Errorf("%v: error logging failed - %v, %v", err, resp.Status, string(body))
		}
	}

	if err != nil {
		log.Entry().Errorf("error sending the requests: %v", err)
	}
}
