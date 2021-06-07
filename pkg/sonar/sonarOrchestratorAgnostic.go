package sonar

import (
	"errors"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type OrchestratorSpecificConfigProviding interface {
	GetBranchBuildConfig() BranchBuildConfig
	GetPullRequestConfig() PullRequestConfig
	IsPullRequest() bool
}

type PullRequestConfig struct {
	Branch string
	Base   string
	Key    string
}

type BranchBuildConfig struct {
	Branch string
}

func NewOrchestratorSpecificConfigProvider() (OrchestratorSpecificConfigProviding, error) {
	o, err := piperutils.DetectOrchestrator()
	if err != nil {
		return nil, err // Don't wrap error here -> Error message wouldn't change
	}

	switch o {
	case piperutils.AzureDevOps:
		return &AzureDevOpsConfigProvider{}, nil
	case piperutils.GitHubActions:
		return &GitHubActionsConfigProvider{}, nil
	case piperutils.Jenkins:
		return &JenkinsConfigProvider{}, nil
	case piperutils.Travis:
		return &TravisConfigProvider{}, nil
	default:
		return nil, errors.New("internal error - Unable to detect orchestrator")
	}
}

// ########################
// #### Azure DevOps ######
// ########################

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

// ########################
// #### GitHub Actions ####
// ########################

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

// ########################
// ####### Jenkins ########
// ########################

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

// ########################
// ######## Travis ########
// ########################

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
	return os.Getenv("TRAVIS_PULL_REQUEST") != "false"
}
