package eventing

import (
	"encoding/json"
	"fmt"

	cdevents "github.com/cdevents/sdk-go/pkg/api"
	cdeventsv04 "github.com/cdevents/sdk-go/pkg/api/v04"
)

// NewTaskRunFinishedCDEvent creates a CDEvents TaskRunFinished event and returns its JSON-serialized CloudEvent bytes.
func NewTaskRunFinishedCDEvent(source, taskName, pipelineURL, outcome string) ([]byte, error) {
	event, err := cdeventsv04.NewTaskRunFinishedEvent()
	if err != nil {
		return nil, fmt.Errorf("failed to create CDEvent: %w", err)
	}

	event.SetSource(source)
	event.SetSubjectId(taskName)
	event.SetSubjectSource(source)
	event.SetSubjectTaskName(taskName)
	event.SetSubjectUrl(pipelineURL)
	event.SetSubjectOutcome(outcome)

	ce, err := cdevents.AsCloudEvent(event)
	if err != nil {
		return nil, fmt.Errorf("failed to convert CDEvent to CloudEvent: %w", err)
	}

	bytes, err := json.Marshal(ce)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CloudEvent: %w", err)
	}
	return bytes, nil
}
