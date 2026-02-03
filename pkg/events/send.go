package events

import (
	"github.com/SAP/jenkins-library/pkg/log"
)

const eventTopicTaskRunFinished = "pipelinetaskrun-finished"
const eventTypeTaskRunFinished = "pipelineTaskRunFinished"

type eventClient interface {
	Publish(topic string, data []byte) error
}

func Send(eventSource, eventType, eventTopic string, payload Payload, client eventClient) error {
	// create cloud event
	event, err := NewEvent(eventType, eventSource, "").CreateWithJSONData(payload.ToJSON())
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
