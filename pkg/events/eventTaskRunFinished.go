package events

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
)

func NewEventTaskRunFinished(eventTypePrefix, eventSource string, payload PayloadTaskRunFinished) ([]byte, error) {
	// Log the payload in human-readable form
	payloadJSON, _ := json.MarshalIndent(payload, "", "  ")
	log.Entry().Infof("Event payload:\n%s", string(payloadJSON))

	eventType := fmt.Sprintf("%seventTypeTaskRunFinished", eventTypePrefix)
	// create cloud event
	event := NewEvent(eventType, eventSource, "").Create(payload)

	// Log the event data in human-readable form
	var eventData map[string]interface{}
	event.cloudEvent.DataAs(&eventData)
	eventDataJSON, _ := json.MarshalIndent(eventData, "", "  ")
	log.Entry().Infof("Event data:\n%s", string(eventDataJSON))

	// convert event to bytes
	eventBytes, err := event.ToBytes()
	if err != nil {
		return []byte{}, err
	}
	return eventBytes, nil
}
