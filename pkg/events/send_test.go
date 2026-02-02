package events

import (
	"encoding/json"
	"errors"
	"testing"
)

type mockClient struct {
	publishErr error
	lastTopic  string
	lastData   []byte
}

func (m *mockClient) Publish(topic string, data []byte) error {
	m.lastTopic = topic
	m.lastData = data
	return m.publishErr
}

func TestSendTaskRunFinished(t *testing.T) {
	validPayload := map[string]string{"taskName": "my-step\"with\"quotes"}
	payloadBytes, _ := json.Marshal(validPayload)

	tests := []struct {
		name                string
		eventSource         string
		eventTypePrefix     string
		eventTopicPrefix    string
		data                string
		additionalEventData string
		publishErr          error
		wantErr             bool
		wantTopic           string
	}{
		{
			name:             "success with valid JSON payload and no additional data",
			eventSource:      "piper",
			eventTypePrefix:  "sap.",
			eventTopicPrefix: "sap.",
			data:             string(payloadBytes),
			wantErr:          false,
			wantTopic:        "sap.pipelinetaskrun-finished",
		},
		{
			name:                "success with valid JSON payload and valid additional data",
			eventSource:         "piper",
			eventTypePrefix:     "sap.",
			eventTopicPrefix:    "sap.",
			data:                string(payloadBytes),
			additionalEventData: `{"correlationId":"abc-123"}`,
			wantErr:             false,
			wantTopic:           "sap.pipelinetaskrun-finished",
		},
		{
			name:             "error on invalid payload JSON",
			eventSource:      "piper",
			eventTypePrefix:  "sap.",
			eventTopicPrefix: "sap.",
			data:             `{invalid`, // malformed JSON
			wantErr:          true,
		},
		{
			name:             "publish error is propagated",
			eventSource:      "piper",
			eventTypePrefix:  "sap.",
			eventTopicPrefix: "sap.",
			data:             string(payloadBytes),
			publishErr:       errors.New("pubsub failure"),
			wantErr:          true,
			wantTopic:        "sap.pipelinetaskrun-finished",
		},
		{
			name:                "invalid additionalEventData is ignored (no failure)",
			eventSource:         "piper",
			eventTypePrefix:     "sap.",
			eventTopicPrefix:    "sap.",
			data:                string(payloadBytes),
			additionalEventData: `{invalid`,
			wantErr:             false,
			wantTopic:           "sap.pipelinetaskrun-finished",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &mockClient{publishErr: tt.publishErr}
			payload := (&PayloadGeneric{JSONData: tt.data})
			payload.Merge(tt.additionalEventData)
			err := SendTaskRunFinished(tt.eventSource, tt.eventTypePrefix, tt.eventTopicPrefix, payload, mc)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SendTaskRunFinished() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantTopic != "" && mc.lastTopic != tt.wantTopic {
				t.Fatalf("Publish() topic = %q, want %q", mc.lastTopic, tt.wantTopic)
			}
			// When successful, ensure we produced a CloudEvent envelope
			if !tt.wantErr {
				if len(mc.lastData) == 0 {
					t.Fatalf("expected event bytes to be published")
				}
				// Minimal sanity check: published bytes should be valid JSON
				var v map[string]interface{}
				if jsonErr := json.Unmarshal(mc.lastData, &v); jsonErr != nil {
					t.Fatalf("published data is not valid JSON: %v", jsonErr)
				}
			}
		})
	}
}

func TestSend(t *testing.T) {
	validPayload := map[string]string{"taskName": "step"}
	payloadBytes, _ := json.Marshal(validPayload)

	t.Run("invalid payload JSON returns error", func(t *testing.T) {
		mc := &mockClient{}
		payload := (&PayloadGeneric{JSONData: `{invalid`})
		err := Send("piper", "sap.pipelineTaskRunFinished", "sap.pipelinetaskrun-finished", payload, mc)
		if err == nil {
			t.Fatalf("expected error for invalid JSON payload")
		}
	})

	t.Run("publish error returns error", func(t *testing.T) {
		mc := &mockClient{publishErr: errors.New("fail")}
		payload := (&PayloadGeneric{JSONData: string(payloadBytes)})
		err := Send("piper", "sap.pipelineTaskRunFinished", "sap.pipelinetaskrun-finished", payload, mc)
		if err == nil {
			t.Fatalf("expected error on publish failure")
		}
	})
}
