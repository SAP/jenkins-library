package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type Event struct {
	cloudEvent  cloudevents.Event
	eventType   string
	eventSource string
	uuidData    string
}

func NewEvent(eventType, eventSource string, uuidString string) Event {
	return Event{
		eventType:   eventType,
		eventSource: eventSource,
		uuidData:    uuidString,
	}
}

func (e Event) CreateWithJSONData(data string) (Event, error) {
	// passing a string to e.cloudEvent.SetData will result in the string being marshalled, ending up with double escape characters
	// therefore pass a map instead
	var dataMap map[string]interface{}
	if data != "" {
		if err := json.Unmarshal([]byte(data), &dataMap); err != nil {
			return e, fmt.Errorf("eventData is an invalid JSON: %w", err)
		}
	}
	return e.Create(dataMap), nil
}

func (e Event) Create(data any) Event {
	e.cloudEvent = cloudevents.NewEvent("1.0")

	if e.uuidData != "" {
		e.cloudEvent.SetID(GetUUID(e.uuidData))
	} else {
		e.cloudEvent.SetID(uuid.New().String())
	}

	// set default values
	e.cloudEvent.SetType(e.eventType)
	e.cloudEvent.SetTime(time.Now())
	e.cloudEvent.SetSource(e.eventSource)
	e.cloudEvent.SetData("application/json", data)

	return e
}

func GetUUID(pipelineIdentifier string) string {
	return uuid.NewMD5(uuid.NameSpaceOID, []byte(pipelineIdentifier)).String()
}

func (e Event) ToBytes() ([]byte, error) {
	data, err := json.Marshal(e.cloudEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}
	return data, nil
}

func (e *Event) ToBytesWithoutEscapeHTML() ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // disable escaping
	if err := encoder.Encode(e.cloudEvent); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *Event) AddToCloudEventData(additionalDataString string) error {
	if additionalDataString == "" {
		return nil
	}

	var additionalData map[string]interface{}
	err := json.Unmarshal([]byte(additionalDataString), &additionalData)
	if err != nil {
		return fmt.Errorf("couldn't add additional data to cloud event: %w", err)
	}

	var newEventData map[string]interface{}
	err = json.Unmarshal(e.cloudEvent.DataEncoded, &newEventData)
	if err != nil {
		return fmt.Errorf("couldn't add additional data to cloud event: %w", err)
	}

	for k, v := range additionalData {
		newEventData[k] = v
	}

	e.cloudEvent.SetData("application/json", newEventData)
	return nil
}
