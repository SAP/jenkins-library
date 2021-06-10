package orchestrator

import (
	"os"
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
	return truthy("CHANGE_ID")
}

func isJenkins() bool {
	envVars := []string{"JENKINS_HOME", "JENKINS_URL"}
	return areIndicatingEnvVarsSet(envVars)
}
