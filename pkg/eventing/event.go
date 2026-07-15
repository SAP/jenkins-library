package eventing

import (
	"encoding/json"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// newEvent creates a CloudEvent v1.0 with the given type, source, and data payload.
func newEvent(eventType, source string, data any) (cloudevents.Event, error) {
	event := cloudevents.NewEvent("1.0")
	event.SetID(uuid.New().String())
	event.SetType(eventType)
	event.SetSource(source)
	event.SetTime(time.Now())

	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return event, fmt.Errorf("failed to set event data: %w", err)
	}

	return event, nil
}

// NewEventFromJSON creates a CloudEvent v1.0 from JSON string data, optionally merging
// additional JSON data into the event payload.
func NewEventFromJSON(eventType, source, jsonData, additionalJSON string) (cloudevents.Event, error) {
	var dataMap map[string]any
	if jsonData != "" {
		if err := json.Unmarshal([]byte(jsonData), &dataMap); err != nil {
			return cloudevents.Event{}, fmt.Errorf("eventData is invalid JSON: %w", err)
		}
	}

	if additionalJSON != "" {
		var additional map[string]any
		if err := json.Unmarshal([]byte(additionalJSON), &additional); err != nil {
			return cloudevents.Event{}, fmt.Errorf("additionalEventData is invalid JSON: %w", err)
		}
		if dataMap == nil {
			dataMap = make(map[string]any)
		}
		for k, v := range additional {
			dataMap[k] = v
		}
	}

	return newEvent(eventType, source, dataMap)
}
