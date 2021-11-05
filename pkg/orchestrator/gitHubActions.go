package orchestrator

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"os"
	"strings"
	"time"
)

type GitHubActionsConfigProvider struct{}

func (g *GitHubActionsConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	log.Entry().Debug("Successfully initalized GitHubActions config provider")
}

func (g *GitHubActionsConfigProvider) OrchestratorVersion() string {
	return "n/a"
}

func (g *GitHubActionsConfigProvider) OrchestratorType() string {
	return "GitHubActions"
}

func (g *GitHubActionsConfigProvider) GetLog() ([]byte, error) {
	log.Entry().Infof("GetLog() for GitHub Actions not yet implemented.")
	return nil, nil
}

func (g *GitHubActionsConfigProvider) GetBuildId() string {
	log.Entry().Infof("GetBuildId() for GitHub Actions not yet implemented.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) GetPipelineStartTime() time.Time {
	log.Entry().Infof("GetPipelineStartTime() for GitHub Actions not yet implemented.")
	timestamp, _ := time.Parse(time.UnixDate, "Wed Feb 25 11:06:39 PST 1970")
	return timestamp
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

func (g *GitHubActionsConfigProvider) GetJobUrl() string {
	log.Entry().Debugf("Not yet implemented.")
	return g.GetRepoUrl() + "/actions/runs/" + getEnv("GITHUB_RUN_ID", "n/a")
}

func (g *GitHubActionsConfigProvider) GetJobName() string {
	log.Entry().Debugf("GetJobName() for GitHubActions not yet implemented.")
	return "N/A"
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
