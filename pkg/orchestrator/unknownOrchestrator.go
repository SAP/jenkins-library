package orchestrator

import (
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

type UnknownOrchestratorConfigProvider struct{}

const unknownOrchestratorWarning = "Unknown orchestrator - returning default values."

func newUnknownOrchestratorConfigProvider() *UnknownOrchestratorConfigProvider {
	return &UnknownOrchestratorConfigProvider{}
}

func (u *UnknownOrchestratorConfigProvider) Configure(_ *Options) error {
	log.Entry().Warning(unknownOrchestratorWarning)
	return nil
}

func (u *UnknownOrchestratorConfigProvider) OrchestratorVersion() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) BuildStatus() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "FAILURE"
}

func (u *UnknownOrchestratorConfigProvider) ChangeSets() []ChangeSet {
	log.Entry().Info(unknownOrchestratorWarning)
	return []ChangeSet{}
}

func (u *UnknownOrchestratorConfigProvider) BuildReason() string {
	log.Entry().Info(unknownOrchestratorWarning)
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) BuildID() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) JobName() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) OrchestratorType() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "Unknown"
}

func (u *UnknownOrchestratorConfigProvider) FullLogs() ([]byte, error) {
	log.Entry().Warning(unknownOrchestratorWarning)
	return []byte{}, nil
}

func (u *UnknownOrchestratorConfigProvider) PipelineStartTime() time.Time {
	log.Entry().Warning(unknownOrchestratorWarning)
	return time.Time{}.UTC()
}

func (u *UnknownOrchestratorConfigProvider) StageName() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) Branch() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) GitReference() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) BuildURL() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) JobURL() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) CommitSHA() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) RepoURL() string {
	log.Entry().Warning(unknownOrchestratorWarning)
	return "n/a"
}

func (u *UnknownOrchestratorConfigProvider) PullRequestConfig() PullRequestConfig {
	log.Entry().Warning(unknownOrchestratorWarning)
	return PullRequestConfig{
		Branch: "n/a",
		Base:   "n/a",
		Key:    "n/a",
	}
}

func (u *UnknownOrchestratorConfigProvider) IsPullRequest() bool {
	log.Entry().Warning(unknownOrchestratorWarning)
	return false
}
