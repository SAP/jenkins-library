package orchestrator

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"os"
	"strings"
)

type GitHubActionsConfigProvider struct{}

func (a *GitHubActionsConfigProvider) OrchestratorVersion() string {
	return "n/a"
}

func (a *GitHubActionsConfigProvider) OrchestratorType() string {
	return "GitHub"
}

func (a *GitHubActionsConfigProvider) GetLog() ([]byte, error) {
	log.Entry().Infof("GetLog() for GitHub Actions not yet implemented.")
	return nil, nil
}

func (a *GitHubActionsConfigProvider) GetPipelineStartTime() string {
	log.Entry().Infof("GetPipelineStartTime() for GitHub Actions not yet implemented.")
	return "n/a"
}
func (g *GitHubActionsConfigProvider) GetStageName() string {
	return "GITHUB_WORKFLOW" //TODO: is there something like is "stage" in GH Actions?
}

func (g *GitHubActionsConfigProvider) GetBranch() string {
	return strings.TrimPrefix(getEnv("GITHUB_REF", "n/a"), "refs/heads/")
}

func (g *GitHubActionsConfigProvider) GetBuildUrl() string {
	return g.GetRepoUrl() + "/actions/runs/" + getEnv("GITHUB_RUN_ID", "n/a")
}

func (g *GitHubActionsConfigProvider) GetCommit() string {
	return getEnv("GITHUB_SHA", "n/a")
}

func (g *GitHubActionsConfigProvider) GetRepoUrl() string {
	return getEnv("GITHUB_SERVER_URL", "n/a") + getEnv("GITHUB_REPOSITORY", "n/a")
}

func (g *GitHubActionsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: getEnv("GITHUB_HEAD_REF", "n/a"),
		Base:   getEnv("GITHUB_BASE_REF", "n/a"),
		Key:    getEnv("GITHUB_EVENT_PULL_REQUEST_NUMBER", "n/a"),
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
