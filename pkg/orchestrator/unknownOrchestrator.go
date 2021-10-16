package orchestrator

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"time"
)

type UnknownOrchestratorConfigProvider struct{}

func (j *UnknownOrchestratorConfigProvider) InitOrchestratorProvider(settings *OrchestratorSettings) {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
}

func (a *UnknownOrchestratorConfigProvider) OrchestratorVersion() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "N/A"
}

func (a *UnknownOrchestratorConfigProvider) OrchestratorType() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "N/A"
}

func (a *UnknownOrchestratorConfigProvider) GetLog() ([]byte, error) {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return nil, nil
}

func (a *UnknownOrchestratorConfigProvider) GetPipelineStartTime() time.Time {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	timestamp, _ := time.Parse(time.UnixDate, "Wed Feb 25 11:06:39 PST 1970")
	return timestamp
}
func (g *UnknownOrchestratorConfigProvider) GetStageName() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "N/A"
}

func (g *UnknownOrchestratorConfigProvider) GetBranch() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "N/A"
}

func (g *UnknownOrchestratorConfigProvider) GetBuildUrl() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "N/A"
}

func (g *UnknownOrchestratorConfigProvider) GetJobUrl() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "N/A"
}

func (g *UnknownOrchestratorConfigProvider) GetCommit() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "N/A"
}

func (g *UnknownOrchestratorConfigProvider) GetRepoUrl() string {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return "N/A"
}

func (g *UnknownOrchestratorConfigProvider) GetPullRequestConfig() PullRequestConfig {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return PullRequestConfig{
		Branch: "N/A",
		Base:   "N/A",
		Key:    "N/A",
	}
}

func (g *UnknownOrchestratorConfigProvider) IsPullRequest() bool {
	log.Entry().Warning("Unknown orchestrator - returning default values.")
	return false
}
