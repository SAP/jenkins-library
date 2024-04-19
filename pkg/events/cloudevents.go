package events

import (
	"encoding/json"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const eventTypePrefix = "sap.hyperspace"

type SAPEventType string

const PipelineRunStartedEventType SAPEventType = "PipelineRunStarted"
const PipelineRunFinishedEventType SAPEventType = "PipelineRunFinished"

type SAPEvent interface {
	GetType() string
}

// type CommonEventData struct {
// 	URL           string `json:"url"`
// 	CommitId      string `json:"commitId"`
// 	RepositoryURL string `json:"repositoryUrl"`
// }

func ToCloudEvent(sapEvent SAPEvent, opts ...Option) cloudevents.Event {
	cloudEvent := cloudevents.NewEvent("1.0")
	// set default values
	cloudEvent.SetID(uuid.New().String())
	cloudEvent.SetType(sapEvent.GetType())
	cloudEvent.SetTime(time.Now())
	cloudEvent.SetSource("/default/sap.hyperspace.piper")
	cloudEvent.SetData("application/json", sapEvent)

	for _, applyOpt := range opts {
		applyOpt(cloudEvent.Context.AsV1())
	}

	return cloudEvent
}

func ToByteArray(sapEvent SAPEvent, opts ...Option) ([]byte, error) {
	event := ToCloudEvent(sapEvent, opts...)
	data, err := json.Marshal(event)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal event data")
	}
	return data, nil
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
