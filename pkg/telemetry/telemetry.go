package telemetry

import (
	"crypto/sha1"
	"fmt"
	"time"

	"net/http"
	"net/url"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
)

// SWA Reporting
// 1. Step usage, details, ... -> SWA implementierung
// 2. Errors -> Hook
// 3. Notify/Deprecations

var disabled bool
var baseData BaseData
var baseMetaData BaseMetaData = BaseMetaData{
	GitOwnerLabel:          "owner",
	GitRepositoryLabel:     "repository",
	StepNameLabel:          "stepName",
	PipelineURLSha1Label:   "",
	BuildURLSha1Label:      "",
	GitPathSha1Label:       "gitpathsha1",
	GitOwnerSha1Label:      "",
	GitRepositorySha1Label: "",
	JobNameLabel:           "jobName",
	StageNameLabel:         "stageName",
	BuildToolLabel:         "buildTool",
	ScanTypeLabel:          "scanType",
}
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

	gitOwner, gitRepository, gitPath := getGitData(envRootPath, getResourceParameter)

	baseData = BaseData{
		GitOwner:      gitOwner,
		GitRepository: gitRepository,
		StepName:      stepName,
		GitPathSha1:   fmt.Sprintf("%x", sha1.Sum([]byte(gitPath))),

		// ToDo: add further params
	}

	//ToDo: register Logrus Hook
}

func getGitData(envRootPath string, getResourceParameter func(rootPath, resourceName, parameterName string) string) (owner, repository, path string) {

	gitOwner := getResourceParameter(envRootPath, "commonPipelineEnvironment", "github/owner")
	gitRepo := getResourceParameter(envRootPath, "commonPipelineEnvironment", "github/repository")

	if len(gitOwner)+len(gitRepo) == 0 {
		// 1st fallback: try to get repositoryUrl from commonPipelineEnvironment
		gitRepoURL := getResourceParameter(envRootPath, "commonPipelineEnvironment", "git/repositoryUrl")

		// 2nd fallback: get repository url from git
		if len(gitRepoURL) == 0 {
			repo, _ := git.Open(nil, nil)

			remote, _ := repo.Remote(git.DefaultRemoteName)

			urlList := remote.Config().URLs

			for url := range urlList {
				fmt.Print(url)
			}

		}

		//ToDo: get owner and repo from url
	}
	path = fmt.Sprintf("%v/%v", owner, repository)
	return
}

// SWA baseURL
const baseURL = "https://webanalytics.cfapps.eu10.hana.ondemand.com"

// SWA endpoint
const endpoint = "/tracker/log"

// site ID
const siteID = "827e8025-1e21-ae84-c3a3-3f62b70b0130"

// SendTelemetry ...
func SendTelemetry(customData *CustomData) {
	data := Data{BaseData: baseData, CustomData: *customData}

	if disabled {
		return
	}

	request, _ := url.Parse(baseURL)
	request.Path = endpoint
	request.RawQuery = data.toPayloadString()
	// Add logic for sending data to SWA
	client.SendRequest(http.MethodGet, request.String(), nil, nil, nil)
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
