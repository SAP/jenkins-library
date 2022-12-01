package orchestrator

import (
	"errors"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

type Orchestrator int

const (
	Unknown Orchestrator = iota
	AzureDevOps
	GitHubActions
	Jenkins
)

type OrchestratorSpecificConfigProviding interface {
	OrchestratorType() string
	OrchestratorVersion() string
	GetStageName() string
	GetBranch() string
	GetReference() string
	GetBuildURL() string
	GetBuildID() string
	GetJobURL() string
	GetJobName() string
	GetCommit() string
	GetPullRequestConfig() PullRequestConfig
	GetRepoURL() string
	IsPullRequest() bool
	GetLog() ([]byte, error)
	GetPipelineStartTime() time.Time
	GetBuildStatus() string
	GetBuildReason() string
	GetChangeSet() []ChangeSet
}

type PullRequestConfig struct {
	Branch string
	Base   string
	Key    string
}

type ChangeSet struct {
	CommitId  string
	Timestamp string
	PrNumber  int
}

// OrchestratorSettings struct to set orchestrator specific settings e.g. Jenkins credentials
type OrchestratorSettings struct {
	JenkinsUser  string
	JenkinsToken string
	AzureToken   string
	GitHubToken  string
}

func NewOrchestratorSpecificConfigProvider() (OrchestratorSpecificConfigProviding, error) {
	switch DetectOrchestrator() {
	case AzureDevOps:
		return &AzureDevOpsConfigProvider{}, nil
	case GitHubActions:
		provider := &GitHubActionsConfigProvider{}
		err := provider.initOrchestratorProvider(&OrchestratorSettings{
			GitHubToken: getEnv("GITHUB_TOKEN", ""),
		})
		return provider, err
	case Jenkins:
		return &JenkinsConfigProvider{}, nil
	default:
		return &UnknownOrchestratorConfigProvider{}, errors.New("unable to detect a supported orchestrator (Azure DevOps, GitHub Actions, Jenkins)")
	}
}

// DetectOrchestrator returns the name of the current orchestrator e.g. Jenkins, Azure, Unknown
func DetectOrchestrator() Orchestrator {
	if isAzure() {
		return Orchestrator(AzureDevOps)
	}
	if isGitHubActions() {
		return Orchestrator(GitHubActions)
	}
	if isJenkins() {
		return Orchestrator(Jenkins)
	}
	return Orchestrator(Unknown)
}

func (o Orchestrator) String() string {
	return [...]string{"Unknown", "AzureDevOps", "GitHubActions", "Jenkins"}[o]
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

// Wrapper function to read env variable and set default value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		log.Entry().Debugf("For: %s, found: %s", key, value)
		return value
	}
	log.Entry().Debugf("Could not read env variable %v using fallback value %v", key, fallback)
	return fallback
}
