package orchestrator

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"time"
)

type UnknownOrchestratorConfigProvider struct{}

func (u *UnknownOrchestratorConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
}

func (u *UnknownOrchestratorConfigProvider) OrchestratorVersion() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "N/A"
}

func (u *UnknownOrchestratorConfigProvider) GetBuildId() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) GetJobName() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) OrchestratorType() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "Unknown"
}

func (u *UnknownOrchestratorConfigProvider) GetLog() ([]byte, error) {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return nil, nil
}

func (u *UnknownOrchestratorConfigProvider) GetPipelineStartTime() time.Time {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	timestamp, _ := time.Parse(time.UnixDate, "Wed Feb 25 11:06:39 PST 1970")
	return timestamp
}
func (u *UnknownOrchestratorConfigProvider) GetStageName() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) GetBranch() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) GetBuildUrl() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) GetJobUrl() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) GetCommit() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) GetRepoUrl() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) GetPullRequestConfig() PullRequestConfig {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return PullRequestConfig{
		Branch: "n/a",
		Base:   "n/a",
		Key:    "n/a",
	}
}

func (u *UnknownOrchestratorConfigProvider) IsPullRequest() bool {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return false
}
