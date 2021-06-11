package orchestrator

import (
	"errors"
	"os"
)

type Orchestrator int

const (
	Unknown Orchestrator = iota
	AzureDevOps
	GitHubActions
	Jenkins
	Travis
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
	switch DetectOrchestrator() {
	case AzureDevOps:
		return &AzureDevOpsConfigProvider{}, nil
	case GitHubActions:
		return &GitHubActionsConfigProvider{}, nil
	case Jenkins:
		return &JenkinsConfigProvider{}, nil
	case Travis:
		return &TravisConfigProvider{}, nil
	case Unknown:
		fallthrough
	default:
		return nil, errors.New("unable to detect a supported orchestrator (Azure DevOps, GitHub Actions, Jenkins, Travis)")
	}
}

func DetectOrchestrator() Orchestrator {
	if isAzure() {
		return Orchestrator(AzureDevOps)
	} else if isGitHubActions() {
		return Orchestrator(GitHubActions)
	} else if isJenkins() {
		return Orchestrator(Jenkins)
	} else if isTravis() {
		return Orchestrator(Travis)
	} else {
		return Orchestrator(Unknown)
	}
}

func (o Orchestrator) String() string {
	return [...]string{"Unknown", "AzureDevOps", "GitHubActions", "Travis", "Jenkins"}[o]
}

func areIndicatingEnvVarsSet(envVars []string) bool {
	for _, v := range envVars {
		if truthy(v) {
			return true
		}
	}
	return false
}

// Checks if var is set and neither empty nor false
func truthy(key string) bool {
	val, exists := os.LookupEnv(key)
	if !exists {
		return false
	}
	if len(val) == 0 || val == "no" || val == "false" || val == "off" || val == "0" {
		return false
	}

	return true
}
