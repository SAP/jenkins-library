package telemetry

import (
	"crypto/sha1"
	"fmt"
	"os"
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

// Initialize sets up the base telemetry data and is called in generated part of the steps
func Initialize(telemetryDisabled bool, stepName string) {
	disabled = telemetryDisabled

	// skip if telemetry is disabled
	if disabled {
		log.Entry().Info("Telemetry reporting deactivated")
		return
	}

	if client == nil {
		client = &piperhttp.Client{}
	}

	client.SetOptions(piperhttp.ClientOptions{MaxRequestDuration: 5 * time.Second})

	if len(LibraryRepository) == 0 {
		LibraryRepository = "https://github.com/n/a"
	}

	if len(SiteID) == 0 {
		SiteID = "827e8025-1e21-ae84-c3a3-3f62b70b0130"
	}

	baseData = BaseData{
		URL:             LibraryRepository,
		ActionName:      actionName,
		EventType:       eventType,
		StepName:        stepName,
		SiteID:          SiteID,
		PipelineURLHash: getPipelineURLHash(), // http://server:port/jenkins/job/foo/
		BuildURLHash:    getBuildURLHash(),    // http://server:port/jenkins/job/foo/15/
	}
	//ToDo: register Logrus Hook

}

func getPipelineURLHash() string {
	return toSha1OrNA(os.Getenv("JOB_URL"))
}

func getBuildURLHash() string {
	return toSha1OrNA(os.Getenv("BUILD_URL"))
}

func toSha1OrNA(input string) string {
	if len(input) == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%x", sha1.Sum([]byte(input)))
}

// SWA baseURL
const baseURL = "https://webanalytics.cfapps.eu10.hana.ondemand.com"

// SWA endpoint
const endpoint = "/tracker/log"

//type TelemetryHook interface {
//	Send(data Data) error
//}

func GetData(customData *CustomData) Data {
	return Data{
		BaseData:     baseData,
		BaseMetaData: baseMetaData,
		CustomData:   *customData,
	}
}

// Send ...
func Send(customData *CustomData) {
	data := GetData(customData)

	// skip if telemetry is disabled
	if disabled {
		return
	}

	request, _ := url.Parse(baseURL)
	request.Path = endpoint
	request.RawQuery = data.toPayloadString()
	log.Entry().WithField("request", request.String()).Debug("Sending telemetry data")
	client.SendRequest(http.MethodGet, request.String(), nil, nil, nil)
}
