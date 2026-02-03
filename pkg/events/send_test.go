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
