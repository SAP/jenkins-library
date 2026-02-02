package events

import (
	"github.com/SAP/jenkins-library/pkg/log"
)

type eventClient interface {
	Publish(topic string, data []byte) error
}

const eventTopicTaskRunFinished = "pipelinetaskrun-finished"
const eventTypeTaskRunFinished = "pipelineTaskRunFinished"

func SendTaskRunFinished(eventSource, eventTypePrefix, eventTopicPrefix string, data EventPayload, client eventClient) error {
	eventType := eventTypePrefix + eventTypeTaskRunFinished
	eventTopic := eventTopicPrefix + eventTopicTaskRunFinished
	return Send(eventSource, eventType, eventTopic, data, client)
}

func Send(eventSource, eventType, eventTopic string, data EventPayload, client eventClient) error {
	// create cloud event
	event, err := NewEvent(eventType, eventSource, "").CreateWithJSONData(data.ToJSON())
	if err != nil {
		return err
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
