package orchestrator

import (
	"fmt"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type AzureDevOpsConfigProvider struct {
	client  piperHttp.Client
	options piperHttp.ClientOptions
}

//InitOrchestratorProvider initializes http client for AzureDevopsConfigProvider
func (a *AzureDevOpsConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	a.client = piperHttp.Client{}
	a.options = piperHttp.ClientOptions{
		Username:         "",
		Password:         settings.AzureToken,
		MaxRetries:       3,
		TransportTimeout: time.Second * 10,
	}
	a.client.SetOptions(a.options)
	log.Entry().Debug("Successfully initialized Azure config provider")
}

var apiInformation map[string]interface{}

func (a *AzureDevOpsConfigProvider) getAPIInformation() {
	// if apiInformation is empty fill it otherwise do nothing
	if len(apiInformation) == 0 {
		log.Entry().Debugf("apiInformation is empty, getting infos from API")
		URL := a.getSystemCollectionURI() + a.getTeamProjectId() + "/_apis/build/builds/" + a.getBuildId() + "/"
		log.Entry().Debugf("API URL: %s", URL)
		response, err := a.client.GetRequest(URL, nil, nil)
		if err != nil {
			log.Entry().Error("failed to get http response, returning empty API information", err)
			apiInformation = map[string]interface{}{}
			return
		}

		if response.StatusCode != 200 { //http.StatusNoContent
			log.Entry().Errorf("Response-Code is %v . \n Could not get API information from AzureDevOps. Returning with empty interface.", response.StatusCode)
			apiInformation = map[string]interface{}{}
			return
		}

		err = piperHttp.ParseHTTPResponseBodyJSON(response, &apiInformation)
		if err != nil {
			log.Entry().Error("failed to parse http response, returning with empty interface", err)
			apiInformation = map[string]interface{}{}
			return
		}
	} else {
		log.Entry().Debugf("apiInformation already set")
	}
}

// GetSystemCollectionURI returns the URI of the TFS collection or Azure DevOps organization e.g. https://dev.azure.com/fabrikamfiber/
func (a *AzureDevOpsConfigProvider) getSystemCollectionURI() string {
	return getEnv("SYSTEM_COLLECTIONURI", "n/a")
}

// GetTeamProjectId is the name of the project that contains this build e.g. 123a4567-ab1c-12a1-1234-123456ab7890
func (a *AzureDevOpsConfigProvider) getTeamProjectId() string {
	return getEnv("SYSTEM_TEAMPROJECTID", "n/a")
}

func (a *AzureDevOpsConfigProvider) getBuildId() string {
	// INFO: Private function only used for API requests, buildId for e.g. reporting
	// is buildNumber to align with the UI of ADO
	return getEnv("BUILD_BUILDID", "n/a")
}

// GetJobName returns the pipeline job name, currently org/repo
func (a *AzureDevOpsConfigProvider) GetJobName() string {
	return getEnv("BUILD_REPOSITORY_NAME", "n/a")
}

// OrchestratorVersion returns the agent version on ADO
func (a *AzureDevOpsConfigProvider) OrchestratorVersion() string {
	return getEnv("AGENT_VERSION", "n/a")
}

// OrchestratorType returns the orchestrator name e.g. Azure/GitHubActions/Jenkins
func (a *AzureDevOpsConfigProvider) OrchestratorType() string {
	return "Azure"
}

//GetBuildStatus returns status of the build. Return variables are aligned with Jenkins build statuses.
func (a *AzureDevOpsConfigProvider) GetBuildStatus() string {
	// cases to align with Jenkins: SUCCESS, FAILURE, NOT_BUILD, ABORTED
	switch buildStatus := getEnv("AGENT_JOBSTATUS", "FAILURE"); buildStatus {
	case "Succeeded":
		return "SUCCESS"
	case "Canceled":
		return "ABORTED"
	default:
		// Failed, SucceededWithIssues
		return "FAILURE"
	}
}

// GetLog returns the whole logfile for the current pipeline run
func (a *AzureDevOpsConfigProvider) GetLog() ([]byte, error) {
	URL := a.getSystemCollectionURI() + a.getTeamProjectId() + "/_apis/build/builds/" + a.GetBuildId() + "/logs"

	response, err := a.client.GetRequest(URL, nil, nil)

	if err != nil {
		log.Entry().Error("failed to get http response", err)
		return []byte{}, nil
	}
	if response.StatusCode != 200 { //http.StatusNoContent -> also empty log!
		log.Entry().Errorf("Response-Code is %v . \n Could not get log information from AzureDevOps. Returning with empty log.", response.StatusCode)
		return []byte{}, nil
	}
	var responseInterface map[string]interface{}
	err = piperHttp.ParseHTTPResponseBodyJSON(response, &responseInterface)
	if err != nil {
		log.Entry().Error("failed to parse http response", err)
		return []byte{}, nil
	}
	// check if response interface is empty or non-existent
	logCount := int(responseInterface["count"].(float64))
	var logs []byte
	for i := 1; i <= logCount; i++ {
		counter := strconv.Itoa(i)
		logURL := URL + "/" + counter
		fmt.Println("logURL: ", logURL)
		log.Entry().Debugf("Getting log no.: %d  from %v", i, logURL)
		response, err := a.client.GetRequest(logURL, nil, nil)
		if err != nil {
			fmt.Println(err)
		}
		content, err := ioutil.ReadAll(response.Body)
		logs = append(logs, content...)
	}

	return logs, nil
}

// GetPipelineStartTime returns the pipeline start time in UTC
func (a *AzureDevOpsConfigProvider) GetPipelineStartTime() time.Time {
	//"2022-03-18T07:30:31.1915758Z"
	a.getAPIInformation()
	if val, ok := apiInformation["startTime"]; ok {
		parsed, err := time.Parse(time.RFC3339, val.(string))
		if err != nil {
			log.Entry().Errorf("could not parse timestamp, %v", err)
			parsed = time.Time{}
		}
		return parsed.UTC()
	}
	return time.Time{}

}

// GetBuildId returns the BuildNumber displayed in the ADO UI
func (a *AzureDevOpsConfigProvider) GetBuildId() string {
	// INFO: ADO has BUILD_ID and buildNumber, as buildNumber is used in the UI we return this value
	// for the buildID used only for API requests we have a private method getBuildId
	// example: buildNumber: 20220318.16 buildId: 76443
	return getEnv("BUILD_BUILDNUMBER", "n/a")
}

// GetStageName returns the human-readable name given to a stage. e.g. "Promote" or "Init"
func (a *AzureDevOpsConfigProvider) GetStageName() string {
	return getEnv("SYSTEM_STAGEDISPLAYNAME", "n/a")
}

func (a *AzureDevOpsConfigProvider) GetBranch() string {
	tmp := getEnv("BUILD_SOURCEBRANCH", "n/a")
	return strings.TrimPrefix(tmp, "refs/heads/")
}

func (a *AzureDevOpsConfigProvider) GetBuildUrl() string {
	return os.Getenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI") + os.Getenv("SYSTEM_TEAMPROJECT") + "/_build/results?buildId=" + a.getBuildId()
}

func (a *AzureDevOpsConfigProvider) GetJobUrl() string {
	// TODO: Check if thi is the correct URL
	return os.Getenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI") + os.Getenv("SYSTEM_TEAMPROJECT")
}

func (a *AzureDevOpsConfigProvider) GetCommit() string {
	return getEnv("BUILD_SOURCEVERSION", "n/a")
}

func (a *AzureDevOpsConfigProvider) GetRepoUrl() string {
	return getEnv("BUILD_REPOSITORY_URI", "n/a")
}

func (a *AzureDevOpsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	prKey := getEnv("SYSTEM_PULLREQUEST_PULLREQUESTID", "n/a")

	// This variable is populated for pull requests which have a different pull request ID and pull request number.
	// In this case the pull request ID will contain an internal numeric ID and the pull request number will be provided
	// as part of the 'SYSTEM_PULLREQUEST_PULLREQUESTNUMBER' environment variable.
	prNumber, prNumberEnvVarSet := os.LookupEnv("SYSTEM_PULLREQUEST_PULLREQUESTNUMBER")
	if prNumberEnvVarSet == true {
		prKey = prNumber
	}

	return PullRequestConfig{
		Branch: os.Getenv("SYSTEM_PULLREQUEST_SOURCEBRANCH"),
		Base:   os.Getenv("SYSTEM_PULLREQUEST_TARGETBRANCH"),
		Key:    prKey,
	}
}

func (a *AzureDevOpsConfigProvider) IsPullRequest() bool {
	return getEnv("BUILD_REASON", "n/a") == "PullRequest"
}

func isAzure() bool {
	envVars := []string{"AZURE_HTTP_USER_AGENT"}
	return areIndicatingEnvVarsSet(envVars)
}
