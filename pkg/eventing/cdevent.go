package eventing

import (
	"encoding/json"
	"fmt"

	cdevents "github.com/cdevents/sdk-go/pkg/api"
	cdeventsv04 "github.com/cdevents/sdk-go/pkg/api/v04"
)

// newPipelineRunStartedCDEvent creates a CDEvents PipelineRunStarted event and returns its JSON-serialized CloudEvent bytes.
// Custom data fields (commitID, repositoryURL, pipelineRunMode, cumulusInformation) can be added via SetCustomData by the consuming team.
func newPipelineRunStartedCDEvent(source, pipelineName, pipelineURL string) ([]byte, error) {
	event, err := cdeventsv04.NewPipelineRunStartedEvent()
	if err != nil {
		return nil, fmt.Errorf("failed to create CDEvent: %w", err)
	}

	event.SetSource(source)
	event.SetSubjectId(pipelineName)
	event.SetSubjectSource(source)
	event.SetSubjectPipelineName(pipelineName)
	event.SetSubjectUrl(pipelineURL)

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

// newTaskRunFinishedCDEvent creates a CDEvents TaskRunFinished event and returns its JSON-serialized CloudEvent bytes.
func newTaskRunFinishedCDEvent(source, taskName, pipelineURL, outcome, stageName string) ([]byte, error) {
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

	if stageName != "" {
		customData := map[string]string{"stageName": stageName}
		if err = event.SetCustomData("application/json", customData); err != nil {
			return nil, fmt.Errorf("failed to set custom data: %w", err)
		}
	}

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
