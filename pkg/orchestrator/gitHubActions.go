package orchestrator

import (
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

type GitHubActionsConfigProvider struct{}

func (g *GitHubActionsConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	log.Entry().Debug("Successfully initialized GitHubActions config provider")
}

func (g *GitHubActionsConfigProvider) OrchestratorVersion() string {
	return "n/a"
}

func (g *GitHubActionsConfigProvider) OrchestratorType() string {
	return "GitHubActions"
}

func (g *GitHubActionsConfigProvider) GetBuildStatus() string {
	log.Entry().Infof("GetBuildStatus() for GitHub Actions not yet implemented.")
	return "FAILURE"
}

func (g *GitHubActionsConfigProvider) GetLog() ([]byte, error) {
	log.Entry().Infof("GetLog() for GitHub Actions not yet implemented.")
	return []byte{}, nil
}

func (g *GitHubActionsConfigProvider) GetBuildID() string {
	log.Entry().Infof("GetBuildID() for GitHub Actions not yet implemented.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) GetChangeSet() []ChangeSet {
	log.Entry().Warn("GetChangeSet for GitHubActions not yet implemented")
	return []ChangeSet{}
}

func (g *GitHubActionsConfigProvider) GetPipelineStartTime() time.Time {
	log.Entry().Infof("GetPipelineStartTime() for GitHub Actions not yet implemented.")
	return time.Time{}.UTC()
}
func (g *GitHubActionsConfigProvider) GetStageName() string {
	return "GITHUB_WORKFLOW" //TODO: is there something like is "stage" in GH Actions?
}

func (g *GitHubActionsConfigProvider) GetBuildReason() string {
	log.Entry().Infof("GetBuildReason() for GitHub Actions not yet implemented.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) GetBranch() string {
	return strings.TrimPrefix(getEnv("GITHUB_REF", "n/a"), "refs/heads/")
}

func (g *GitHubActionsConfigProvider) GetReference() string {
	return getEnv("GITHUB_REF", "n/a")
}

func (g *GitHubActionsConfigProvider) GetBuildURL() string {
	return g.GetRepoURL() + "/actions/runs/" + getEnv("GITHUB_RUN_ID", "n/a")
}

func (g *GitHubActionsConfigProvider) GetJobURL() string {
	log.Entry().Debugf("Not yet implemented.")
	return g.GetRepoURL() + "/actions/runs/" + getEnv("GITHUB_RUN_ID", "n/a")
}

func (g *GitHubActionsConfigProvider) GetJobName() string {
	log.Entry().Debugf("GetJobName() for GitHubActions not yet implemented.")
	return "n/a"
}

func (g *GitHubActionsConfigProvider) GetCommit() string {
	return getEnv("GITHUB_SHA", "n/a")
}

func (g *GitHubActionsConfigProvider) GetRepoURL() string {
	return getEnv("GITHUB_SERVER_URL", "n/a") + "/" + getEnv("GITHUB_REPOSITORY", "n/a")
}

func (g *GitHubActionsConfigProvider) GetPullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: getEnv("GITHUB_HEAD_REF", "n/a"),
		Base:   getEnv("GITHUB_BASE_REF", "n/a"),
		Key:    getEnv("GITHUB_EVENT_PULL_REQUEST_NUMBER", "n/a"),
	}
}

func (g *GitHubActionsConfigProvider) IsPullRequest() bool {
	return truthy("GITHUB_HEAD_REF")
}

func isGitHubActions() bool {
	envVars := []string{"GITHUB_ACTION", "GITHUB_ACTIONS"}
	return areIndicatingEnvVarsSet(envVars)
}
