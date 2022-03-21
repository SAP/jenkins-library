package orchestrator

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"time"
)

type UnknownOrchestratorConfigProvider struct{}

// InitOrchestratorProvider returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
}

// OrchestratorVersion returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) OrchestratorVersion() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

// GetBuildStatus returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetBuildStatus() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "FAILURE"
}

// GetBuildID returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetBuildID() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

// GetJobName returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetJobName() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

// OrchestratorType returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) OrchestratorType() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "Unknown"
}

// GetLog returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetLog() ([]byte, error) {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return []byte{}, nil
}

// GetPipelineStartTime returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetPipelineStartTime() time.Time {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return time.Time{}.UTC()
}

// GetStageName returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetStageName() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

// GetBranch returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetBranch() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

// GetBuildUrl returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetBuildUrl() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

// GetJobUrl returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetJobUrl() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

// GetCommit returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetCommit() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

// GetRepoUrl returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetRepoUrl() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "n/a"
}

// GetPullRequestConfig returns n/a for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) GetPullRequestConfig() PullRequestConfig {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return PullRequestConfig{
		Branch: "n/a",
		Base:   "n/a",
		Key:    "n/a",
	}
}

// IsPullRequest returns false for the unknownOrchestrator
func (u *UnknownOrchestratorConfigProvider) IsPullRequest() bool {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return false
}
