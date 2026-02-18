package events

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
)

func NewEventTaskRunFinished(eventTypePrefix, eventSource string, payload PayloadTaskRunFinished) ([]byte, error) {
	eventType := fmt.Sprintf("%seventTypeTaskRunFinished", eventTypePrefix)
	// create cloud event
	event := NewEvent(eventType, eventSource, "").Create(payload)
	log.Entry().Debugf("event %+v", event)
	// log event payload
	var newEventData map[string]any
	event.cloudEvent.DataAs(&newEventData)
	log.Entry().Debugf("event data %+v", newEventData)
	// convert event to bytes
	eventBytes, err := event.ToBytes()
	if err != nil {
		return []byte{}, err
	}
	return eventBytes, nil
}
