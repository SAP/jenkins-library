package orchestrator

import "os"

type GitHubActionsConfigProvider struct{}

func (a *GitHubActionsConfigProvider) GetBranchBuildConfig() BranchBuildConfig {
	return BranchBuildConfig{Branch: os.Getenv("GITHUB_REF")}
}

func (a *GitHubActionsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: os.Getenv("GITHUB_HEAD_REF"),
		Base:   os.Getenv("GITHUB_BASE_REF"),
		Key:    os.Getenv("GITHUB_EVENT_PULL_REQUEST_NUMBER"),
	}
}

func (a *GitHubActionsConfigProvider) IsPullRequest() bool {
	_, exists := os.LookupEnv("GITHUB_HEAD_REF")
	return exists
}

func isGitHubActions() bool {
	envVars := []string{"GITHUB_ACTION", "GITHUB_ACTIONS"}
	return areIndicatingEnvVarsSet(envVars)
}
