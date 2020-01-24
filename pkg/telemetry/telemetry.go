package telemetry

import (
	"crypto/sha1"
	"fmt"
	"os"
	"time"

	"net/http"
	"net/url"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

// site ID
const siteID = "827e8025-1e21-ae84-c3a3-3f62b70b0130"

// LibraryRepository that is passed into with -ldflags
var LibraryRepository string

var disabled bool
var client piperhttp.Sender

// Initialize sets up the base telemetry data and is called in generated part of the steps
func Initialize(telemetryActive bool, _ func(rootPath, resourceName, parameterName string) string, _, stepName string) {
	//TODO: change parameter semantic to avoid double negation
	disabled = !telemetryActive

	// skip if telemetry is dieabled
	if disabled {
		return
	}

	if client == nil {
		client = &piperhttp.Client{}
	}

	client.SetOptions(piperhttp.ClientOptions{Timeout: time.Second * 5})

	baseData = BaseData{
		URL:             LibraryRepository,
		ActionName:      "Piper Library OS",
		EventType:       "library-os",
		StepName:        stepName,
		SiteID:          siteID,
		PipelineURLHash: getPipelineURLHash(), // http://server:port/jenkins/job/foo/
		BuildURLHash:    getBuildURLHash(),    // http://server:port/jenkins/job/foo/15/
		//GitOwnerHash:
		//GitOwner:      gitOwner,
		//GitRepositoryHash:
		//GitRepository: gitRepository,

		// JOB_URL - JOB_BASE_NAME = count repositories
		// JOB_URL - JOB_BASE_NAME - "/job/.*/job/" = count orgs

		//JOB_NAME
		//Projektname des Builds, z.B. "foo" oder "foo/bar". (Um in einem Bourne Shell-Script den Pfadanteil abzuschneiden, probieren Sie: ${JOB_NAME##*/})
		//JOB_BASE_NAME
		//Short Name of the project of this build stripping off folder paths, such as "foo" for "bar/foo".

		// JENKINS_URL // http://server:port/jenkins/
		//GitPathSha1:   fmt.Sprintf("%x", sha1.Sum([]byte(gitPath))),
		// ToDo: add further params
	}
	//ToDo: register Logrus Hook
}

func getPipelineURLHash() string {
	return toSha1(os.Getenv("JOB_URL"))
}

func getBuildURLHash() string {
	return toSha1(os.Getenv("BUILD_URL"))
}

func toSha1(input string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(input)))
}

// SWA baseURL
const baseURL = "https://webanalytics.cfapps.eu10.hana.ondemand.com"

// SWA endpoint
const endpoint = "/tracker/log"

// SendTelemetry ...
func SendTelemetry(customData *CustomData) {
	data := Data{
		BaseData:     baseData,
		BaseMetaData: baseMetaData,
		CustomData:   *customData,
	}

	// skip if telemetry is dieabled
	if disabled {
		return
	}

	request, _ := url.Parse(baseURL)
	request.Path = endpoint
	request.RawQuery = data.toPayloadString()
	client.SendRequest(http.MethodGet, request.String(), nil, nil, nil)
}
