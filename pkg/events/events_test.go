package events

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestEventCreation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// init
		// test
		event := NewEvent(mock.Anything, mock.Anything, "").Create(nil)
		// asserts
		assert.Equal(t, mock.Anything, event.cloudEvent.Type())
		assert.Equal(t, mock.Anything, event.cloudEvent.Source())
	})

	t.Run("CreateWithJSONData - no additional data", func(t *testing.T) {
		// init
		testData := `{"testKey":"testValue"}`
		// test
		event, err := NewEvent(mock.Anything, mock.Anything, "").CreateWithJSONData(testData)
		// asserts
		assert.NoError(t, err)
		assert.Equal(t, string(event.cloudEvent.Data()), testData)
	})

	t.Run("CreateWithJSONData + AddToCloudEventData", func(t *testing.T) {
		// init
		testData := `{"testKey": "testValue"}`
		additionalData := `{"additionalKey": "additionalValue"}`
		// test
		event, err := NewEvent(mock.Anything, mock.Anything, "").CreateWithJSONData(testData)
		event.AddToCloudEventData(additionalData)
		// asserts
		assert.NoError(t, err)
		assert.Equal(
			t,
			string(event.cloudEvent.Data()),
			`{"additionalKey":"additionalValue","testKey":"testValue"}`,
		)
	})
}

func TestGetUUID(t *testing.T) {
	pipelineIdentifier := "pipelineIdentifier"
	uuid := GetUUID(pipelineIdentifier)

	if uuid == "" {
		t.Fatalf("expected a UUID but got none")
	}

	uuid2 := GetUUID(pipelineIdentifier)
	if uuid != uuid2 {
		t.Fatalf("expected the same UUID but got different ones")
	}
}

func TestSkipEscapeForHTML(t *testing.T) {
	event := cloudevents.NewEvent()
	event.SetSource("test/source")
	event.SetType("test.type")
	event.SetID("fixed-id-1234")

	event.SetData(cloudevents.ApplicationJSON, map[string]string{
		"message": "Hello & welcome",
	})

	eventWrapper := Event{
		cloudEvent: event,
	}
	result, err := eventWrapper.ToBytesWithoutEscapeHTML()

	got := string(result)

	expected := `{
	  "specversion": "1.0",
	"type": "test.type",
	"source": "test/source",
	"id": "fixed-id-1234",
	"datacontenttype": "application/json",
	"data": {
			"message": "Hello & welcome"
		}
	}
	`
	assert.NoError(t, err)
	assert.JSONEq(
		t,
		expected,
		got,
	)
}

func TestSafeDataFromKV(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"simple", "taskName", "step"},
		{"with quotes", "taskName", `my"step'name`},
		{"unicode", "taskName", "ステップ"},
		{"special chars", "taskName", `}{:"',\n`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := SafeDataFromKV(tt.key, tt.value)
			if err != nil {
				t.Fatalf("SafeDataFromKV error: %v", err)
			}
			var obj map[string]string
			if err := json.Unmarshal([]byte(s), &obj); err != nil {
				t.Fatalf("payload is not valid JSON: %v\npayload: %s", err, s)
			}
			if obj[tt.key] != tt.value {
				t.Fatalf("value mismatch: got %q want %q", obj[tt.key], tt.value)
			}
		})
	}
}

func TestSafeDataFromTaskName(t *testing.T) {
	s, err := SafeDataFromTaskName(`my"step`)
	if err != nil {
		t.Fatalf("SafeDataFromTaskName error: %v")
	}
	var obj map[string]string
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		t.Fatalf("payload is not valid JSON: %v\npayload: %s", err, s)
	}
	if obj["taskName"] != `my"step` {
		t.Fatalf("value mismatch: got %q", obj["taskName"])
	}
}
