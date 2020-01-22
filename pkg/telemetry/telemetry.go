package telemetry

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	"net/http"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/fatih/structs"
	"github.com/sirupsen/logrus"
)

// SWA Reporting
// 1. Step usage, details, ... -> SWA implementierung
// 2. Errors -> Hook
// 3. Notify/Deprecations

// BaseData ...
type BaseData struct {
	Active     bool
	ActionName string `structs:"actionName"`
	EventType  string `structs:"eventType"`
	SiteID     string `structs:"idsite"`
	URL        string `structs:"url"`

	GitOwner               string `structs:"e_a"`
	GitOwnerLabel          string `structs:"custom1"`
	GitRepository          string `structs:"e_2"`
	GitRepositoryLabel     string `structs:"custom2"`
	StepName               string `structs:"e_3"`
	StepNameLabel          string `structs:"custom_3"`
	PipelineURLSha1        string `structs:"e_4"` // defaults to env.JOB_URl
	PipelineURLSha1Label   string `structs:"custom_4"`
	BuildURLSha1           string `structs:"e_5"` // defaults to env.BUILD_URL
	BuildURLSha1Label      string `structs:"custom_5"`
	GitPathSha1            string `structs:"e_6"`
	GitPathSha1Label       string `structs:"custom_6"`
	GitOwnerSha1           string `structs:"e_7"`
	GitOwnerSha1Label      string `structs:"custom_7"`
	GitRepositorySha1      string `structs:"e_8"`
	GitRepositorySha1Label string `structs:"custom_8"`
	JobName                string `structs:"e_9"`
	JobNameLabel           string `structs:"custom_9"`
	StageName              string `structs:"e_10"`
	StageNameLabel         string `structs:"custom_10"`
}

// CustomData ...
type CustomData struct {
	BuildTool      string `structs:"e_11"`
	buildToolLabel string `structs:"custom_11"`
	// ...
	ScanType      string `structs:"e_24"`
	scanTypeLabel string `structs:"custom_24"`
	Custom25      string `structs:"e_25"`
	custom25Label string `structs:"custom_25"`
	Custom26      string `structs:"e_26"`
	custom26Label string `structs:"custom_26"`
	Custom27      string `structs:"e_27"`
	custom27Label string `structs:"custom_27"`
	Custom28      string `structs:"e_28"`
	custom28Label string `structs:"custom_28"`
	Custom29      string `structs:"e_29"`
	custom29Label string `structs:"custom_29"`
	Custom30      string `structs:"e_30"`
	Custom30Label string `structs:"custom_30"`
}

type Data struct {
	BaseData
	CustomData
}

func (d *Data) toPayloadString() (payload string) {
	dataMap := structs.Map(d)
	dataList := make([]string, 0)

	for key, value := range dataMap {
		dataList = append(dataList, fmt.Sprintf("%v&%v", key, value))
	}

	payload = strings.Join(dataList, "&")

	return
}

var baseData BaseData
var client piperhttp.Sender

// Initialize sets up the base telemetry data and is called in generated part of the steps
func Initialize(telemetryActive bool, getResourceParameter func(rootPath, resourceName, parameterName string) string, envRootPath, stepName string) {
	// check if telemetry is disabled
	if telemetryActive {
		baseData = BaseData{Active: telemetryActive}
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
const endpoint = "https://webanalytics.cfapps.eu10.hana.ondemand.com/tracker/log"

// site ID
const siteID = "827e8025-1e21-ae84-c3a3-3f62b70b0130"

// SendTelemetry ...
func SendTelemetry(customData *CustomData) {
	data := Data{BaseData: baseData, CustomData: *customData}

	if data.Active != true {
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
