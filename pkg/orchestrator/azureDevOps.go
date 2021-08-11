package orchestrator

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"os"
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
	log.Entry().Infof("GetLog() for Azure not yet implemented.")
	return nil, nil
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
