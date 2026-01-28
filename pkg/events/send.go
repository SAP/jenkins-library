package events

import (
	"github.com/SAP/jenkins-library/pkg/log"
)

type eventClient interface {
	Publish(topic string, data []byte) error
}

const eventTypeTaskRunFinished = "pipelineTaskRunFinished"
const eventTopicTaskRunFinished = "pipelinetaskrun-finished"

func SendTaskRunFinishedEvent(eventSource, eventTypePrefix, eventTopicPrefix, data, additionalEventData string, client eventClient) error {
	eventType := eventTypePrefix + eventTypeTaskRunFinished
	eventTopic := eventTopicPrefix + eventTopicTaskRunFinished
	return SendEvent(eventSource, eventType, eventTopic, data, additionalEventData, client)
}

func SendEvent(eventSource, eventType, eventTopic, data, additionalEventData string, client eventClient) error {
	// create cloud event
	event, err := NewEvent(eventType, eventSource, "").CreateWithJSONData(data)
	if err != nil {
		return err
	}
	// log.Entry().Infof("  with data %s", event)
	log.Entry().Debugf("event %+v", event)
	log.Entry().Debugf("event data %+v", event.cloudEvent.Data())

	err = event.AddToCloudEventData(additionalEventData)
	if err != nil {
		log.Entry().Debugf("couldn't add additionalData to cloud event data: %s", err)
	}

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
