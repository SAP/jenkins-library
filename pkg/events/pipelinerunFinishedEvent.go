package events

import (
	"strings"

	"github.com/SAP/jenkins-library/pkg/orchestrator"
)

const PipelinerunFinishedEventType EventType = "PipelineRunFinished"

type PipelinerunFinishedEvent struct {
	Event
	provider orchestrator.ConfigProvider
}

func (e PipelinerunFinishedEvent) GetType() string {
	return strings.Join([]string{eventTypePrefix, "pipelinerunFinished"}, ".")
}

func (e PipelinerunFinishedEvent) Create(opts ...Option) PipelinerunFinishedEvent {
	e.Event = e.Event.Create(e.GetType(), PipelinerunFinishedEventData{
		URL:           e.provider.BuildURL(),
		CommitId:      e.provider.CommitSHA(),
		RepositoryURL: e.provider.RepoURL(),
		Outcome:       e.provider.BuildStatus(),
	}, opts...)
	return e
}

type PipelinerunFinishedEventData struct {
	URL           string `json:"url"`
	CommitId      string `json:"commitId"`
	RepositoryURL string `json:"repositoryUrl"`
	Outcome       string `json:"outcome"`
}

func NewPipelinerunFinishedEvent(provider orchestrator.ConfigProvider) PipelinerunFinishedEvent {
	return PipelinerunFinishedEvent{
		provider: provider,
	}
}
