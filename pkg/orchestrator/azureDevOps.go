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
		Username: "",
		Password: settings.AzureToken,
	}
	a.client.SetOptions(a.options)
	log.Entry().Debug("Successfully initialized Azure config provider")
}

// OrchestratorVersion returns the agent version on ADO
func (a *AzureDevOpsConfigProvider) OrchestratorVersion() string {
	return getEnv("AGENT_VERSION", "n/a")
}

// OrchestratorType returns the orchestrator name e.g. Azure/GitHubActions/Jenkins
func (a *AzureDevOpsConfigProvider) OrchestratorType() string {
	return "Azure"
}

// GetJobName returns the pipeline job name
func (a *AzureDevOpsConfigProvider) GetJobName() string {
	log.Entry().Debugf("GetJobName() for Azure not yet implemented.")
	return "n/a"
}

// GetLog returns the logfile of the pipeline run so far
func (a *AzureDevOpsConfigProvider) GetLog() ([]byte, error) {

	// ToDo: How to get step specific logs, not only whole log?

	URL := a.GetSystemCollectionURI() + a.GetTeamProjectId() + "/_apis/build/builds/" + a.GetBuildId() + "/logs"

	response, err := a.client.GetRequest(URL, nil, nil)
	logs := []byte{}
	if err != nil {
		fmt.Println(err)
	}
	if response.StatusCode != 200 { //http.StatusNoContent -> also empty log!
		log.Entry().Errorf("Response-Code is %v . \n Could not get log information from AzureDevOps. Returning with empty log.", response.StatusCode)
		return logs, nil
	}
	var responseInterface map[string]interface{}
	err = piperHttp.ParseHTTPResponseBodyJSON(response, &responseInterface)
	if err != nil {
		fmt.Println(err)
	}
	// check if response interface is empty or non-existent
	logCount := int(responseInterface["count"].(float64))

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

// GetPipelineStartTime returns the pipeline start time
func (a *AzureDevOpsConfigProvider) GetPipelineStartTime() time.Time {
	// "2021-10-11 13:49:09+00:00"
	timestamp := getEnv("SYSTEM_PIPELINESTARTTIME", "n/a")
	replaced := strings.Replace(timestamp, " ", "T", 1)
	parsed, err := time.Parse(time.RFC3339, replaced)
	if err != nil {
		log.Entry().Errorf("Could not parse timestamp. %v", err)
		// Return 1970 in case parsing goes wrong
		parsed = time.Date(1970, time.January, 01, 0, 0, 0, 0, time.UTC)
	}
	return parsed
}

func (a *AzureDevOpsConfigProvider) GetSystemCollectionURI() string {
	return getEnv("SYSTEM_COLLECTIONURI", "n/a")
}

func (a *AzureDevOpsConfigProvider) GetTeamProjectId() string {
	return getEnv("SYSTEM_TEAMPROJECTID", "n/a")
}

func (a *AzureDevOpsConfigProvider) GetBuildId() string {
	return getEnv("BUILD_BUILDID", "n/a")
}

func (a *AzureDevOpsConfigProvider) GetStageName() string {
	return getEnv("SYSTEM_STAGEDISPLAYNAME", "n/a")
}

func (a *AzureDevOpsConfigProvider) GetBranch() string {
	tmp := getEnv("BUILD_SOURCEBRANCH", "n/a")
	return strings.TrimPrefix(tmp, "refs/heads/")
}

func (a *AzureDevOpsConfigProvider) GetBuildUrl() string {
	return os.Getenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI") + os.Getenv("SYSTEM_TEAMPROJECT") + "/_build/results?buildId=" + os.Getenv("BUILD_BUILDID")
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
