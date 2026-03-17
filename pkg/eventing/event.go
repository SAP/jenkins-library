package eventing

import (
	"encoding/json"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// newEvent creates a CloudEvent v1.0 with the given type, source, and data payload,
// and returns its JSON-serialized bytes.
func newEvent(eventType, source string, data any) ([]byte, error) {
	event := cloudevents.NewEvent("1.0")
	event.SetID(uuid.New().String())
	event.SetType(eventType)
	event.SetSource(source)
	event.SetTime(time.Now())

	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, fmt.Errorf("failed to set event data: %w", err)
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}
	return bytes, nil
}

// NewEventFromJSON creates a CloudEvent v1.0 from JSON string data, optionally merging
// additional JSON data into the event payload. Returns the serialized CloudEvent bytes.
func NewEventFromJSON(eventType, source, jsonData, additionalJSON string) ([]byte, error) {
	var dataMap map[string]any
	if jsonData != "" {
		if err := json.Unmarshal([]byte(jsonData), &dataMap); err != nil {
			return nil, fmt.Errorf("eventData is invalid JSON: %w", err)
		}
	}

	if additionalJSON != "" {
		var additional map[string]any
		if err := json.Unmarshal([]byte(additionalJSON), &additional); err != nil {
			return nil, fmt.Errorf("additionalEventData is invalid JSON: %w", err)
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
