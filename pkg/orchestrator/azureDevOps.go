package orchestrator

import (
	"fmt"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type AzureDevOpsConfigProvider struct{}

func (a *AzureDevOpsConfigProvider) OrchestratorVersion() string {
	return "123"
}

func (a *AzureDevOpsConfigProvider) OrchestratorType() string {
	return "Azure"
}

func (a *AzureDevOpsConfigProvider) GetLog() ([]byte, error) {

	URL := a.GetSystemCollectionURI() + a.GetTeamProjectId() + "/_apis/build/builds/" + a.GetBuildId() + "/logs"
	//URL = "https://dev.azure.com/hyperspace-pipelines/8d6e7755-9b5a-4036-a67e-33b95cda3a3f/_apis/build/builds/7804/logs"

	client := &piperHttp.Client{}
	options := piperHttp.ClientOptions{
		// We do not need a username, but the http package does only create the base64 encoded
		// string if the username is larger than 0. So we misuse the username for the PAT.
		Username: "fbntcoh4ttplq6xa4uqjpwaqckmdmnvvu3gebpuah7zmycskygla",
		Password: "",
	}
	client.SetOptions(options)
	response, err := client.GetRequest(URL, nil, nil)
	if err != nil {
		fmt.Println(err)
	}
	var responseInterface map[string]interface{}
	err = piperHttp.ParseHTTPResponseBodyJSON(response, &responseInterface)
	if err != nil {
		fmt.Println(err)
	}
	// check if response interface is empty or non-existent
	logCount := int(responseInterface["count"].(float64))

	logs := []byte{}

	for i := 1; i <= logCount; i++ {
		counter := strconv.Itoa(i)
		logURL := URL + "/" + counter
		fmt.Println("logURL: ", logURL)
		log.Entry().Debugf("Getting log no.: %d  from %v", i, logURL)
		response, err := client.GetRequest(logURL, nil, nil)
		if err != nil {
			fmt.Println(err)
		}
		content, err := ioutil.ReadAll(response.Body)
		logs = append(logs, content...)
	}

	return logs, nil
}

func (a *AzureDevOpsConfigProvider) GetPipelineStartTime() string {
	return getEnv("SYSTEM_PIPELINESTARTTIME", "n/a")
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

func (a *AzureDevOpsConfigProvider) GetBranch() string {
	tmp := getEnv("BUILD_SOURCEBRANCH", "n/a")
	return strings.TrimPrefix(tmp, "refs/heads/")
}

func (a *AzureDevOpsConfigProvider) GetBuildUrl() string {
	return os.Getenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI") + os.Getenv("SYSTEM_TEAMPROJECT") + "/_build/results?buildId=" + os.Getenv("BUILD_BUILDID")
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
