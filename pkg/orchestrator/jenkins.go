package orchestrator

import (
	"os"
)

type JenkinsConfigProvider struct{}

func (a *JenkinsConfigProvider) GetStageName() string {
	return os.Getenv("STAGE_NAME")
}

func (j *JenkinsConfigProvider) GetBranch() string {
	return os.Getenv("GIT_BRANCH")
}

func (j *JenkinsConfigProvider) GetBuildUrl() string {
	return os.Getenv("BUILD_URL")
}

func (j *JenkinsConfigProvider) GetCommit() string {
	return os.Getenv("GIT_COMMIT")
}

func (j *JenkinsConfigProvider) GetRepoUrl() string {
	return os.Getenv("GIT_URL")
}

func (j *JenkinsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: os.Getenv("CHANGE_BRANCH"),
		Base:   os.Getenv("CHANGE_TARGET"),
		Key:    os.Getenv("CHANGE_ID"),
	}
}

func (j *JenkinsConfigProvider) IsPullRequest() bool {
	return truthy("CHANGE_ID")
}

func isJenkins() bool {
	envVars := []string{"JENKINS_HOME", "JENKINS_URL"}
	return areIndicatingEnvVarsSet(envVars)
}
