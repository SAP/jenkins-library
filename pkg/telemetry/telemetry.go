package telemetry

import (
	"crypto/sha1"
	"fmt"
	"time"

	"net/http"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
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
var client piperhttp.Sender

// InitializeTelemetry sets up the base telemetry data and is called in generated part of the steps
func InitializeTelemetry(telemetryActive bool, getResourceParameter func(rootPath, resourceName, parameterName string) string, envRootPath, stepName string) {
	// check if telemetry is disabled
	if telemetryActive {
		data = TelemetryBaseData{Active: telemetryActive}
		return
	}

	client := piperhttp.Client{}
	client.SetOptions(piperhttp.ClientOptions{Timeout: time.Second * 5})

	gitOwner := getResourceParameter(envRootPath, "commonPipelineEnvironment", "github/owner")
	gitRepo := getResourceParameter(envRootPath, "commonPipelineEnvironment", "github/repository")

	if len(gitOwner)+len(gitRepo) == 0 {
		// 1st fallback: try to get repositoryUrl from commonPipelineEnvironment
		gitRepoURL := getResourceParameter(envRootPath, "commonPipelineEnvironment", "git/repositoryUrl")

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

// SWA endpoint
const ENDPOINT = "https://webanalytics.cfapps.eu10.hana.ondemand.com/tracker/log"
const SITE_ID = "827e8025-1e21-ae84-c3a3-3f62b70b0130"

// SendTelemetry ...
func SendTelemetry(customData *TelemetryCustomData) {
	payload := ""

	// Add logic for sending data to SWA
	client.SendRequest(http.MethodGet, fmt.Sprintf("%v?%v", ENDPOINT, payload), nil, nil, nil)
}

// WARNING ...
const WARNING = "WARNING"

// ERROR ...
const ERROR = "ERROR"

// Notify ...
func Notify(level, message string) {
	data := TelemetryCustomData{}
	SendTelemetry(&data)

	notification := log.Entry().WithField("type", "notification")

	switch level {
	case WARNING:
		notification.Warning(message)
	case ERROR:
		notification.Fatal(message)
	}
}

// Fire ...
func (t *TelemetryBaseData) Fire(entry *logrus.Entry) error {
	telemetryData := TelemetryCustomData{}
	SendTelemetry(&telemetryData)
	return nil
}

// Levels ...
func (t *TelemetryBaseData) Levels() (levels []logrus.Level) {
	levels = append(levels, logrus.ErrorLevel)
	levels = append(levels, logrus.FatalLevel)
	//levels = append(levels, logrus.PanicLevel)
	return
}
