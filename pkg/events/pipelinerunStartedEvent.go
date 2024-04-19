package events

import (
	"strings"

	"github.com/SAP/jenkins-library/pkg/orchestrator"
)

const PipelinerunStartedEventType EventType = "PipelineRunStarted"

type PipelinerunStartedEvent struct {
	Event
	provider orchestrator.ConfigProvider
}

func (e PipelinerunStartedEvent) GetType() string {
	return strings.Join([]string{eventTypePrefix, "pipelinerunStarted"}, ".")
}

func (e PipelinerunStartedEvent) Create(opts ...Option) PipelinerunStartedEvent {
	e.Event = e.Event.Create(e.GetType(), PipelinerunStartedEventData{
		URL:           e.provider.BuildURL(),
		CommitId:      e.provider.CommitSHA(),
		RepositoryURL: e.provider.RepoURL(),
	}, opts...)
	return e
}

type PipelinerunStartedEventData struct {
	URL           string `json:"url"`
	CommitId      string `json:"commitId"`
	RepositoryURL string `json:"repositoryUrl"`
}

func NewPipelinerunStartedEvent(provider orchestrator.ConfigProvider) PipelinerunStartedEvent {
	return PipelinerunStartedEvent{
		provider: provider,
	}
}
