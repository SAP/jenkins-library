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

const (
	BuildStatusSuccess    = "SUCCESS"
	BuildStatusAborted    = "ABORTED"
	BuildStatusFailure    = "FAILURE"
	BuildStatusInProgress = "IN_PROGRESS"

	BuildReasonManual          = "Manual"
	BuildReasonSchedule        = "Schedule"
	BuildReasonPullRequest     = "PullRequest"
	BuildReasonResourceTrigger = "ResourceTrigger"
	BuildReasonIndividualCI    = "IndividualCI"
	BuildReasonUnknown         = "Unknown"
)

type OrchestratorSpecificConfigProviding interface {
	InitOrchestratorProvider(settings *OrchestratorSettings)
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
		ghProvider := &GitHubActionsConfigProvider{}
		// Temporary workaround: The orchestrator provider is not always initialized after being created,
		// which causes a panic in some places for GitHub Actions provider, as it needs to initialize
		// github sdk client.
		ghProvider.InitOrchestratorProvider(&OrchestratorSettings{})
		return ghProvider, nil
	case Jenkins:
		return &JenkinsConfigProvider{}, nil
	default:
		return &UnknownOrchestratorConfigProvider{}, errors.New("unable to detect a supported orchestrator (Azure DevOps, GitHub Actions, Jenkins)")
	}
}

// DetectOrchestrator returns the name of the current orchestrator e.g. Jenkins, Azure, Unknown
func DetectOrchestrator() Orchestrator {
	if isAzure() {
		return AzureDevOps
	} else if isGitHubActions() {
		return GitHubActions
	} else if isJenkins() {
		return Jenkins
	} else {
		return Unknown
	}
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
