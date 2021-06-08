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
	o, err := DetectOrchestrator()
	if err != nil {
		return nil, err // Don't wrap error here -> Error message wouldn't change
	}

	switch o {
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
		return nil, errors.New("unable to detect orchestrator")
	}
}

func DetectOrchestrator() (Orchestrator, error) {
	if isAzure() {
		return Orchestrator(AzureDevOps), nil
	} else if isGitHubActions() {
		return Orchestrator(GitHubActions), nil
	} else if isJenkins() {
		return Orchestrator(Jenkins), nil
	} else if isTravis() {
		return Orchestrator(Travis), nil
	} else {
		return Orchestrator(Unknown), errors.New("unable to detect a supported orchestrator (Azure DevOps, GitHub Actions, Jenkins, Travis)")
	}
}

func (o Orchestrator) String() string {
	return [...]string{"Unknown", "AzureDevOps", "GitHubActions", "Travis", "Jenkins"}[o]
}

func areIndicatingEnvVarsSet(envVars []string) bool {
	found := false
	for _, v := range envVars {
		found = truthy(v)
	}
	return found
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
