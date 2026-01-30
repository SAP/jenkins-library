package events

import (
	"github.com/SAP/jenkins-library/pkg/log"
)

type eventClient interface {
	Publish(topic string, data []byte) error
}

const eventTopicTaskRunFinished = "pipelinetaskrun-finished"
const eventTypeTaskRunFinished = "pipelineTaskRunFinished"

// Add uuidString parameter so callers can provide a stable identifier
func SendTaskRunFinished(eventSource, eventTypePrefix, eventTopicPrefix, data, additionalEventData, uuidString string, client eventClient) error {
	eventType := eventTypePrefix + eventTypeTaskRunFinished
	eventTopic := eventTopicPrefix + eventTopicTaskRunFinished
	return Send(eventSource, eventType, eventTopic, data, additionalEventData, uuidString, client)
}

func Send(eventSource, eventType, eventTopic, data, additionalEventData, uuidString string, client eventClient) error {
	// create cloud event with provided uuidString (falls back to random inside NewEvent if empty)
	event, err := NewEvent(eventType, eventSource, uuidString).CreateWithJSONData(data)
	if err != nil {
		return err
	}

	if err = event.AddToCloudEventData(additionalEventData); err != nil {
		log.Entry().Debugf("couldn't add additionalData to cloud event data: %s", err)
	}

	log.Entry().Debugf("event %+v", event)
	// log event payload
	var newEventData map[string]interface{}
	event.cloudEvent.DataAs(&newEventData)
	log.Entry().Debugf("event data %+v", newEventData)

	// publish cloud event via GCP Pub/Sub
	eventBytes, err := event.ToBytes()
	if err != nil {
		return err
	}
	if err = client.Publish(eventTopic, eventBytes); err != nil {
		return err
	}
	return nil
}
