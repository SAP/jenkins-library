package orchestrator

import (
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/environment"
	"github.com/SAP/jenkins-library/pkg/log"
)

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

var (
	provider     ConfigProvider
	providerOnce sync.Once
)

type ConfigProvider interface {
	Configure(opts *Options) error
	OrchestratorType() string
	OrchestratorVersion() string
	StageName() string
	Branch() string
	GitReference() string
	RepoURL() string
	BuildURL() string
	BuildID() string
	BuildStatus() string
	BuildReason() string
	JobURL() string
	JobName() string
	CommitSHA() string
	PullRequestConfig() PullRequestConfig
	IsPullRequest() bool
	FullLogs() ([]byte, error)
	PipelineStartTime() time.Time
	ChangeSets() []ChangeSet
}

type (
	Orchestrator int

	// Options used to set orchestrator specific settings.
	Options struct {
		JenkinsUsername string
		JenkinsToken    string
		AzureToken      string
		GitHubToken     string
	}

	PullRequestConfig struct {
		Branch string
		Base   string
		Key    string
	}

	ChangeSet struct {
		CommitId  string
		Timestamp string
		PrNumber  int
	}
)

func GetOrchestratorConfigProvider(opts *Options) (ConfigProvider, error) {
	var err error
	providerOnce.Do(func() {
		switch DetectOrchestrator() {
		case AzureDevOps:
			provider = newAzureDevopsConfigProvider()
		case GitHubActions:
			provider = newGithubActionsConfigProvider()
		case Jenkins:
			provider = newJenkinsConfigProvider()
		default:
			provider = newUnknownOrchestratorConfigProvider()
			err = errors.New("unable to detect a supported orchestrator (Azure DevOps, GitHub Actions, Jenkins)")
		}
	})
	if err != nil {
		return provider, err
	}

	if opts == nil {
		log.Entry().Debug("ConfigProvider options are not set. Provider configuration is skipped.")
		return provider, nil
	}

	// This allows configuration of the provider during initialization and/or after it (reconfiguration)
	if cfgErr := provider.Configure(opts); cfgErr != nil {
		return provider, errors.Wrap(cfgErr, "provider configuration failed")
	}

	return provider, nil
}

// DetectOrchestrator function determines in which orchestrator Piper is running by examining environment variables.
func DetectOrchestrator() Orchestrator {
	if isAzure() {
		return AzureDevOps
	} else if environment.IsGitHubActions() {
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

// ResetConfigProvider is intended to be used only for unit tests because some of these tests
// run with different environment variables (for example, mock runs in various orchestrators).
// Usage in production code is not recommended.
func ResetConfigProvider() {
	provider = nil
	providerOnce = sync.Once{}
}
