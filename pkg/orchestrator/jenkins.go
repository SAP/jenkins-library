package orchestrator

import (
	"os"
	"strings"
)

type JenkinsConfigProvider struct{}

func (a *JenkinsConfigProvider) GetBranchBuildConfig() BranchBuildConfig {
	return BranchBuildConfig{Branch: os.Getenv("BRANCH_NAME")}
}

func (a *JenkinsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: os.Getenv("CHANGE_BRANCH"),
		Base:   os.Getenv("CHANGE_TARGET"),
		Key:    os.Getenv("CHANGE_ID"),
	}
}

func (a *JenkinsConfigProvider) IsPullRequest() bool {
	tmp := os.Getenv("BRANCH_NAME")
	return strings.HasPrefix(tmp, "PR")
}
