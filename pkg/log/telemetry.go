package log

import (
	"crypto/sha1"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/sirupsen/logrus"
)

// SWA Reporting
// 1. Step usage, details, ... -> SWA implementierung
// 2. Errors -> Hook
// 3. Notify/Deprecations

// TelemetryBaseData ...
type TelemetryBaseData struct {
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

// TelemetryCustomData ...
type TelemetryCustomData struct {
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

var data TelemetryBaseData

// Initialize sets up the base telemetry data and is call in generated part of the steps
func Initialize(telemetryActive bool, path, stepName string) {
	if telemetryActive {
		data = TelemetryBaseData{Active: telemetryActive}
		return
	}

	gitOwner := piperenv.GetResourceParameter(path, "commonPipelineEnvironment", "github/owner")
	gitRepo := piperenv.GetResourceParameter(path, "commonPipelineEnvironment", "github/repository")

	if len(gitOwner)+len(gitRepo) == 0 {
		// 1st fallback: try to get repositoryUrl from commonPipelineEnvironment
		gitRepoURL := piperenv.GetResourceParameter(path, "commonPipelineEnvironment", "git/repositoryUrl")

		// 2nd fallback: get repository url from git
		if len(gitRepoURL) == 0 {

		}

		//ToDo: get owner and repo from url
	}

	gitPath := fmt.Sprintf("%v/%v", gitOwner, gitRepo)

	data = TelemetryBaseData{
		Active: telemetryActive,

		GitOwner:           gitOwner,
		GitOwnerLabel:      "owner",
		GitRepository:      gitRepo,
		GitRepositoryLabel: "repository",
		StepName:           stepName,
		StepNameLabel:      "stepName",
		GitPathSha1:        fmt.Sprintf("%x", sha1.Sum([]byte(gitPath))),
		GitPathSha1Label:   "gitpathsha1",

		// ToDo: add further params
	}

	//ToDo: register Logrus Hook
}

// SendTelemetry ...
func SendTelemetry(customData *TelemetryCustomData) {
	// Add logic for sending data to SWA
}

// WARNING ...
const WARNING = "WARNING"

// ERROR ...
const ERROR = "ERROR"

// Notify ...
func Notify(level, message string) {
	data := TelemetryCustomData{}
	SendTelemetry(&data)

	switch level {
	case WARNING:
		Entry().WithField("type", "notification").Warning("")
	case ERROR:
		Entry().WithField("type", "notification").Fatal("")
	}
}

// Fire ...
func (t *TelemetryBaseData) Fire(entry *logrus.Entry) error {
	telemetryData := TelemetryCustomData{}
	SendTelemetry(&telemetryData)
	return nil
}

// Levels ...
func (t *TelemetryBaseData) Levels() []logrus.Level {
	//not all levels, only Error, Fatal?
	return logrus.AllLevels
}
