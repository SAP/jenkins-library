package events

import (
	"encoding/json"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const eventTypePrefix = "sap.hyperspace"

type EventType string

type Event struct {
	cloudEvent cloudevents.Event
}

func (e Event) Create(eventType string, data any, opts ...Option) Event {
	e.cloudEvent = cloudevents.NewEvent("1.0")
	// set default values
	e.cloudEvent.SetID(uuid.New().String())
	e.cloudEvent.SetType(eventType)
	e.cloudEvent.SetTime(time.Now())
	e.cloudEvent.SetSource("/default/sap.hyperspace.piper")
	e.cloudEvent.SetData("application/json", data)

	for _, applyOpt := range opts {
		applyOpt(e.cloudEvent.Context.AsV1())
	}

	return e
}

func (e Event) ToBytes() ([]byte, error) {
	data, err := json.Marshal(e.cloudEvent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal event data")
	}
	return data, nil
}
