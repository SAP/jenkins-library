package events

import (
	"bytes"
	"encoding/json"
	"time"

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
	uuidData    string
}

func NewEvent(eventType, eventSource, uuidString string) Event {
	return Event{
		eventType:   eventType,
		eventSource: eventSource,
		uuidData:    uuidString,
	}
}

func (e Event) CreateWithJSONData(data string, opts ...Option) (Event, error) {
	// passing a string to e.cloudEvent.SetData will result in the string being marshalled, ending up with double escape characters
	// therefore pass a map instead
	var dataMap map[string]interface{}
	if data != "" {
		err := json.Unmarshal([]byte(data), &dataMap)
		if err != nil {
			return e, errors.Wrap(err, "eventData is an invalid JSON")
		}
	}
	return e.Create(dataMap, opts...), nil
}

func (e Event) Create(data any, opts ...Option) Event {
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

	for _, applyOpt := range opts {
		applyOpt(e.cloudEvent.Context.AsV1())
	}
	return e
}

func GetUUID(pipelineIdentifier string) string {
	return uuid.NewMD5(uuid.NameSpaceOID, []byte(pipelineIdentifier)).String()
}

func (e Event) ToBytes() ([]byte, error) {
	data, err := json.Marshal(e.cloudEvent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal event data")
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
		errors.Wrap(err, "couldn't add additional data to cloud event")
	}

	var newEventData map[string]interface{}
	err = json.Unmarshal(e.cloudEvent.DataEncoded, &newEventData)
	if err != nil {
		errors.Wrap(err, "couldn't add additional data to cloud event")
	}

	for k, v := range additionalData {
		newEventData[k] = v
	}

	e.cloudEvent.SetData("application/json", newEventData)
	return nil
}

// SafeDataFromKV builds a valid JSON object from a single key/value using encoding/json.
func SafeDataFromKV(key, value string) (string, error) {
	payload := map[string]string{key: value}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal event payload")
	}
	return string(b), nil
}

// SafeDataFromTaskName builds the standard payload containing taskName.
func SafeDataFromTaskName(taskName string) (string, error) {
	return SafeDataFromKV("taskName", taskName)
}
