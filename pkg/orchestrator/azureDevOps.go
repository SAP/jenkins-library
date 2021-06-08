package orchestrator

import (
	"os"
	"strings"
)

type AzureDevOpsConfigProvider struct{}

func (a *AzureDevOpsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: os.Getenv("SYSTEM_PULLREQUEST_SOURCEBRANCH"),
		Base:   os.Getenv("SYSTEM_PULLREQUEST_TARGETBRANCH"),
		Key:    os.Getenv("SYSTEM_PULLREQUEST_PULLREQUESTID"),
	}
}

func (a *AzureDevOpsConfigProvider) GetBranchBuildConfig() BranchBuildConfig {
	tmp := os.Getenv("BUILD_SOURCEBRANCH")
	trimmed := strings.TrimPrefix(tmp, "refs/heads/")
	return BranchBuildConfig{Branch: trimmed}
}

func (a *AzureDevOpsConfigProvider) IsPullRequest() bool {
	return os.Getenv("BUILD_REASON") == "PullRequest"
}

func isAzure() bool {
	envVars := []string{"AZURE_HTTP_USER_AGENT"}
	return areIndicatingEnvVarsSet(envVars)
}
