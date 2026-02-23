package eventing

import (
	"encoding/json"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// NewEvent creates a CloudEvent v1.0 with the given type, source, and data payload,
// and returns its JSON-serialized bytes.
func NewEvent(eventType, source string, data any) ([]byte, error) {
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
