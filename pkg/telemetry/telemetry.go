package telemetry

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"strings"
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

// BaseData ...
type BaseData struct {
	ActionName string `json:"actionName"`
	EventType  string `json:"eventType"`
	SiteID     string `json:"idsite"`
	URL        string `json:"url"`

	GitOwner               string `json:"e_a"`
	GitOwnerLabel          string `json:"custom1"`
	GitRepository          string `json:"e_2"`
	GitRepositoryLabel     string `json:"custom2"`
	StepName               string `json:"e_3"`
	StepNameLabel          string `json:"custom_3"`
	PipelineURLSha1        string `json:"e_4"` // defaults to env.JOB_URl
	PipelineURLSha1Label   string `json:"custom_4"`
	BuildURLSha1           string `json:"e_5"` // defaults to env.BUILD_URL
	BuildURLSha1Label      string `json:"custom_5"`
	GitPathSha1            string `json:"e_6"`
	GitPathSha1Label       string `json:"custom_6"`
	GitOwnerSha1           string `json:"e_7"`
	GitOwnerSha1Label      string `json:"custom_7"`
	GitRepositorySha1      string `json:"e_8"`
	GitRepositorySha1Label string `json:"custom_8"`
	JobName                string `json:"e_9"`
	JobNameLabel           string `json:"custom_9"`
	StageName              string `json:"e_10"`
	StageNameLabel         string `json:"custom_10"`
}

// CustomData ...
type CustomData struct {
	BuildTool      string `json:"e_11"`
	buildToolLabel string `json:"custom_11"`
	// ...
	ScanType      string `json:"e_24"`
	scanTypeLabel string `json:"custom_24"`
	Custom25      string `json:"e_25"`
	custom25Label string `json:"custom_25"`
	Custom26      string `json:"e_26"`
	custom26Label string `json:"custom_26"`
	Custom27      string `json:"e_27"`
	custom27Label string `json:"custom_27"`
	Custom28      string `json:"e_28"`
	custom28Label string `json:"custom_28"`
	Custom29      string `json:"e_29"`
	custom29Label string `json:"custom_29"`
	Custom30      string `json:"e_30"`
	Custom30Label string `json:"custom_30"`
}

type Data struct {
	BaseData
	CustomData
}

func (d *Data) toMap() (result map[string]string) {
	jsonObj, _ := json.Marshal(d)
	json.Unmarshal(jsonObj, &result)
	return
}

func (d *Data) toPayloadString() string {
	dataList := []string{}

	for key, value := range d.toMap() {
		if len(value) > 0 {
			dataList = append(dataList, fmt.Sprintf("%v=%v", key, value))
		}
	}

	return strings.Join(dataList, "&")
}

var disabled bool
var baseData BaseData
var client piperhttp.Sender

// Initialize sets up the base telemetry data and is called in generated part of the steps
func Initialize(telemetryActive bool, getResourceParameter func(rootPath, resourceName, parameterName string) string, envRootPath, stepName string) {
	//TODO: change parameter semantic to avoid double negation
	disabled = !telemetryActive

	// check if telemetry is disabled
	if disabled {
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

	baseData = BaseData{
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
const endpoint = "https://webanalytics.cfapps.eu10.hana.ondemand.com/tracker/log"

// site ID
const siteID = "827e8025-1e21-ae84-c3a3-3f62b70b0130"

// SendTelemetry ...
func SendTelemetry(customData *CustomData) {
	data := Data{BaseData: baseData, CustomData: *customData}

	if disabled {
		return
	}

	payload := data.toPayloadString()
	// Add logic for sending data to SWA
	client.SendRequest(http.MethodGet, fmt.Sprintf("%v?%v", endpoint, payload), nil, nil, nil)
}

// WARNING ...
const WARNING = "WARNING"

// ERROR ...
const ERROR = "ERROR"

// Notify ...
func Notify(level, message string) {
	data := CustomData{}
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
func (t *BaseData) Fire(entry *logrus.Entry) error {
	telemetryData := CustomData{}
	SendTelemetry(&telemetryData)
	return nil
}

// Levels ...
func (t *BaseData) Levels() (levels []logrus.Level) {
	levels = append(levels, logrus.ErrorLevel)
	levels = append(levels, logrus.FatalLevel)
	//levels = append(levels, logrus.PanicLevel)
	return
}
