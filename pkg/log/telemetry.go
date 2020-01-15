package log

import (
	"github.com/sirupsen/logrus"
)

// SWA Reporting
// 1. Step usage, details, ... -> SWA implementierung
// 2. Errors -> Hook
// 3. Notify/Deprecations

// TelemetryBaseData ...
type telemetryBaseData struct {
	Active     bool
	ActionName string `json:"actionName,omitempty"`
	EventType  string `json:"eventType,omitempty"`
	SiteID     string `json:"idsite,omitempty"`
	URL        string `json:"url,omitempty"`

	GitOwner               string `json:"e_a,omitempty"`
	GitOwnerLabel          string `json:"custom1,omitempty"`
	GitRepository          string `json:"e_2,omitempty"`
	GitRepositoryLabel     string `json:"custom2,omitempty"`
	StepName               string `json:"e_3,omitempty"`
	StepNameLabel          string `json:"custom_3,omitempty"`
	PipelineURLSha1        string `json:"e_4,omitempty"` // defaults to env.JOB_URl
	PipelineURLSha1Label   string `json:"custom_4,omitempty"`
	BuildURLSha1           string `json:"e_5,omitempty"` // defaults to env.BUILD_URL
	BuildURLSha1Label      string `json:"custom_5,omitempty"`
	GitPathSha1            string `json:"e_6,omitempty"`
	GitPathSha1Label       string `json:"custom_6,omitempty"`
	GitOwnerSha1           string `json:"e_7,omitempty"`
	GitOwnerSha1Label      string `json:"custom_7,omitempty"`
	GitRepositorySha1      string `json:"e_8,omitempty"`
	GitRepositorySha1Label string `json:"custom_8,omitempty"`
	JobName                string `json:"e_9,omitempty"`
	JobNameLabel           string `json:"custom_9,omitempty"`
	StageName              string `json:"e_10,omitempty"`
	StageNameLabel         string `json:"custom_10,omitempty"`
}

// TelemetryStepData ...
type TelemetryStepData struct {
	BuildTool      string `json:"e_11,omitempty"`
	buildToolLabel string `json:"custom_11,omitempty"`
	// ...
	ScanType      string `json:"e_24,omitempty"`
	scanTypeLabel string `json:"custom_24,omitempty"`
	Custom25      string `json:"e_25,omitempty"`
	custom25Label string `json:"custom_25,omitempty"`
	Custom26      string `json:"e_26,omitempty"`
	custom26Label string `json:"custom_26,omitempty"`
	Custom27      string `json:"e_27,omitempty"`
	custom27Label string `json:"custom_27,omitempty"`
	Custom28      string `json:"e_28,omitempty"`
	custom28Label string `json:"custom_28,omitempty"`
	Custom29      string `json:"e_29,omitempty"`
	custom29Label string `json:"custom_29,omitempty"`
	Custom30      string `json:"e_30,omitempty"`
	Custom30Label string `json:"custom_30,omitempty"`
}

var data telemetryBaseData

// Initialize sets up the base telemetry data and is call in generated part of the steps
func Initialize(telemetryActive bool) {
	if telemetryActive {
		data = telemetryBaseData{Active: telemetryActive}
		return
	}

	data = telemetryBaseData{
		Active: telemetryActive,

		GitOwner: 
		GitOwnerLabel: "owner",
		// ToDo ...

		GitRepositoryLabel: "repository",
	}
}

// SendTelemetry ...
func SendTelemetry(stepData *TelemetryStepData) {

}

const WARNING = "WARNING"
const ERROR = "ERROR"

func Notify(level, message string) {
	data := TelemetryData{}
	data.Send()

	switch level {
	case WARNING:
		Entry().WithField("type", "notification").Warning("")
	case ERROR:
		Entry().WithField("type", "notification").Fatal("")
	}
}

func test() {
	//Errors
	Entry().Error("error")
	Entry().Fatal("error")
}

func (t *TelemetryData) Fire(entry *logrus.Entry) error {
	//fields := entry.Data
	return nil
}

func (t *TelemetryData) Levels() []logrus.Level {
	//not all levels, only Error, Fatal?
	return logrus.AllLevels
}
