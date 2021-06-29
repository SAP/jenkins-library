package orchestrator

import (
	"os"
	"strings"
)

type AzureDevOpsConfigProvider struct{}

func (a *AzureDevOpsConfigProvider) GetBranch() string {
	tmp := os.Getenv("BUILD_SOURCEBRANCH")
	return strings.TrimPrefix(tmp, "refs/heads/")
}

func (a *AzureDevOpsConfigProvider) GetBuildUrl() string {
	return os.Getenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI") + os.Getenv("SYSTEM_TEAMPROJECT") + "/_build/results?buildId=" + os.Getenv("BUILD_BUILDID")
}

func (a *AzureDevOpsConfigProvider) GetCommit() string {
	return os.Getenv("BUILD_SOURCEVERSION")
}

func (a *AzureDevOpsConfigProvider) GetRepoUrl() string {
	return os.Getenv("BUILD_REPOSITORY_URI")
}

func (a *AzureDevOpsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: os.Getenv("SYSTEM_PULLREQUEST_SOURCEBRANCH"),
		Base:   os.Getenv("SYSTEM_PULLREQUEST_TARGETBRANCH"),
		Key:    os.Getenv("SYSTEM_PULLREQUEST_PULLREQUESTID"),
	}
}

func (a *AzureDevOpsConfigProvider) IsPullRequest() bool {
	return os.Getenv("BUILD_REASON") == "PullRequest"
}

func isAzure() bool {
	envVars := []string{"AZURE_HTTP_USER_AGENT"}
	return areIndicatingEnvVarsSet(envVars)
}
