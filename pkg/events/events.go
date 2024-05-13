package events

import (
	"encoding/json"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// type EventType string

type EventData struct {
	URL           string `json:"url"`
	CommitId      string `json:"commitId"`
	RepositoryURL string `json:"repositoryUrl"`
}

type Event struct {
	cloudEvent  cloudevents.Event
	eventType   string
	eventSource string
}

func NewEvent(eventType, eventSource string) Event {
	return Event{
		eventType:   eventType,
		eventSource: eventSource,
	}
}

func (e Event) CreateWithJSONData(data []byte, additionalData string, opts ...Option) Event {
	event := e.Create(data, opts...)
	err := event.AddToCloudEventData(additionalData)
	if err != nil {
		log.Entry().Debugf("couldn't add additionalData to cloud event data: %s", err)
	}
	return event
}

func (e Event) Create(data any, opts ...Option) Event {
	e.cloudEvent = cloudevents.NewEvent("1.0")
	// set default values
	e.cloudEvent.SetID(uuid.New().String())
	e.cloudEvent.SetType(e.eventType)
	e.cloudEvent.SetTime(time.Now())
	e.cloudEvent.SetSource(e.eventSource)
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

func (e *Event) AddToCloudEventData(additionalDataString string) error {
	if additionalDataString == "" {
		return nil
	}
	var additionalData map[string]interface{}
	json.Unmarshal([]byte(additionalDataString), &additionalData)

	var newEventData map[string]interface{}
	err := json.Unmarshal(e.cloudEvent.DataEncoded, &newEventData)
	if err != nil {
		errors.Wrap(err, "couldn't add additional data to cloud event")
	}

	for k, v := range additionalData {
		newEventData[k] = v
	}

	eventData, err := json.Marshal(newEventData)
	if err != nil {
		return errors.Wrap(err, "couldn't add additional data to cloud event")
	}
	e.cloudEvent.SetData("application/json", eventData)

	return nil
}
