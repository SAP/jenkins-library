package orchestrator

import "os"

type TravisConfigProvider struct{}

func (a *TravisConfigProvider) GetBranchBuildConfig() BranchBuildConfig {
	return BranchBuildConfig{Branch: os.Getenv("TRAVIS_BRANCH")}
}

func (a *TravisConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: os.Getenv("TRAVIS_PULL_REQUEST_BRANCH"),
		Base:   os.Getenv("TRAVIS_BRANCH"),
		Key:    os.Getenv("TRAVIS_PULL_REQUEST"),
	}
}

func (a *TravisConfigProvider) IsPullRequest() bool {
	return truthy("TRAVIS_PULL_REQUEST")
}

func isTravis() bool {
	envVars := []string{"TRAVIS"}
	return areIndicatingEnvVarsSet(envVars)
}
