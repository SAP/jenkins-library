package events

import (
	"encoding/json"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/pkg/errors"
)

const eventTypePrefix = "sap.hyperspace"

type SAPEventType string

const PipelineRunStartedEventType SAPEventType = "PipelineRunStarted"
const PipelineRunFinishedEventType SAPEventType = "PipelineRunFinished"

type SAPEvent interface {
	GetType() string
}

type CommonEventData struct {
	URL           string `json:"url"`
	CommitId      string `json:"commitId"`
	RepositoryURL string `json:"repositoryUrl"`
}

type PipelineRunStartedEventData struct {
	URL           string `json:"url"`
	CommitId      string `json:"commitId"`
	RepositoryURL string `json:"repositoryUrl"`
}

func (e PipelineRunStartedEventData) GetType() string {
	return strings.Join([]string{eventTypePrefix, "pipelinerunStarted"}, ".")
}

type PipelineRunFinishedEventData struct {
	URL           string `json:"url"`
	CommitId      string `json:"commitId"`
	RepositoryURL string `json:"repositoryUrl"`
	Outcome       string `json:"outcome"`
}

func (e PipelineRunFinishedEventData) GetType() string {
	return strings.Join([]string{eventTypePrefix, "pipelinerunFinished"}, ".")
}

func ToCloudEvent(context cloudevents.EventContextV1, sapEvent SAPEvent) cloudevents.Event {
	cloudEvent := cloudevents.NewEvent()
	cloudEvent.Context = &context

	cloudEvent.SetData("application/json", sapEvent)
	cloudEvent.SetType(sapEvent.GetType())

	return cloudEvent
}

func ToByteArray(context cloudevents.EventContextV1, sapEvent SAPEvent) ([]byte, error) {
	event := ToCloudEvent(context, sapEvent)
	data, err := json.Marshal(event)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal event data")
	}
	return data, nil
}
