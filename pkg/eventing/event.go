package eventing

import (
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"uuid"
)

type cloudEvent struct {
	SpecVersion     string `json:"specversion"`
	ID              string `json:"id"`
	Type            string `json:"type"`
	Source          string `json:"source"`
	Time            string `json:"time"`
	DataContentType string `json:"datacontenttype,omitempty"`
	Data            any    `json:"data,omitempty"`
}

// newEvent creates a CloudEvent v1.0 with the given type, source, and data payload,
// and returns its JSON-serialized bytes.
func newEvent(eventType, source string, data any) ([]byte, error) {
	event := cloudEvent{
		SpecVersion: "1.0",
		ID:          uuid.New().String(),
		Type:        eventType,
		Source:      source,
		Time:        time.Now().UTC().Format(time.RFC3339Nano),
	}
	if data != nil {
		event.DataContentType = "application/json"
		event.Data = data
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
		maps.Copy(dataMap, additional)
	}

	return newEvent(eventType, source, dataMap)
}
