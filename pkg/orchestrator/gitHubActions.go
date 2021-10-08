package orchestrator

import (
	"os"
	"strings"
)

type GitHubActionsConfigProvider struct{}

func (g *GitHubActionsConfigProvider) GetStageName() string {
	return "GITHUB_WORKFLOW" //TODO: is there something like is "stage" in GH Actions?
}

func (g *GitHubActionsConfigProvider) GetBranch() string {
	return strings.TrimPrefix(os.Getenv("GITHUB_REF"), "refs/heads/")
}

func (g *GitHubActionsConfigProvider) GetBuildUrl() string {
	return g.GetRepoUrl() + "/actions/runs/" + os.Getenv("GITHUB_RUN_ID")
}

func (g *GitHubActionsConfigProvider) GetCommit() string {
	return os.Getenv("GITHUB_SHA")
}

func (g *GitHubActionsConfigProvider) GetRepoUrl() string {
	return os.Getenv("GITHUB_SERVER_URL") + os.Getenv("GITHUB_REPOSITORY")
}

func (g *GitHubActionsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: os.Getenv("GITHUB_HEAD_REF"),
		Base:   os.Getenv("GITHUB_BASE_REF"),
		Key:    os.Getenv("GITHUB_EVENT_PULL_REQUEST_NUMBER"),
	}
}

func (g *GitHubActionsConfigProvider) IsPullRequest() bool {
	_, exists := os.LookupEnv("GITHUB_HEAD_REF")
	return exists
}

func isGitHubActions() bool {
	envVars := []string{"GITHUB_ACTION", "GITHUB_ACTIONS"}
	return areIndicatingEnvVarsSet(envVars)
}
