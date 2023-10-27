package orchestrator

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"sync"
	"time"
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
		User      string
		AuthToken string
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
			provider = &azureDevopsConfigProvider{}
		case GitHubActions:
			provider = &githubActionsConfigProvider{}
		case Jenkins:
			provider = &jenkinsConfigProvider{}
		default:
			provider = &UnknownOrchestratorConfigProvider{}
			err = errors.New("unable to detect a supported orchestrator (Azure DevOps, GitHub Actions, Jenkins)")
		}

		if opts == nil {
			log.Entry().Debug("ConfigProvider initialized without options. Some data may be unavailable")
			return
		}

		if cfgErr := provider.Configure(opts); cfgErr != nil {
			err = errors.Wrap(cfgErr, "provider configuration failed")
		}
	})

	return provider, err
}

// DetectOrchestrator function determines in which orchestrator Piper is running by examining environment variables.
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
