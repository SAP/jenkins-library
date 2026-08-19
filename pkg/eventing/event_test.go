//go:build unit

package eventing

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	t.Run("creates valid CloudEvent with map data", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		bytes, err := newEvent("test.type", "test/source", data)
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)
		assert.Equal(t, "1.0", result["specversion"])
		assert.Equal(t, "test.type", result["type"])
		assert.Equal(t, "test/source", result["source"])
		assert.NotEmpty(t, result["id"])
		assert.NotEmpty(t, result["time"])

		eventData := result["data"].(map[string]any)
		assert.Equal(t, "value", eventData["key"])
	})

	t.Run("creates valid CloudEvent with struct data", func(t *testing.T) {
		payload := struct {
			TaskName  string `json:"taskName"`
			StageName string `json:"stageName"`
			Outcome   string `json:"outcome"`
		}{
			TaskName:  "build",
			StageName: "dev",
			Outcome:   "success",
		}
		bytes, err := newEvent("test.type", "test/source", payload)
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		eventData := result["data"].(map[string]any)
		assert.Equal(t, "build", eventData["taskName"])
		assert.Equal(t, "dev", eventData["stageName"])
		assert.Equal(t, "success", eventData["outcome"])
	})

	t.Run("handles nil data", func(t *testing.T) {
		bytes, err := newEvent("test.type", "test/source", nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
	})
}

func TestNewEventFromJSON(t *testing.T) {
	t.Run("creates event from JSON data", func(t *testing.T) {
		bytes, err := NewEventFromJSON("test.type", "test/source", `{"key":"value"}`, "")
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)
		assert.Equal(t, "test.type", result["type"])

		eventData := result["data"].(map[string]any)
		assert.Equal(t, "value", eventData["key"])
	})

	t.Run("merges additional JSON data", func(t *testing.T) {
		bytes, err := NewEventFromJSON("test.type", "test/source",
			`{"key":"value"}`,
			`{"extra":"data"}`,
		)
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		eventData := result["data"].(map[string]any)
		assert.Equal(t, "value", eventData["key"])
		assert.Equal(t, "data", eventData["extra"])
	})

	t.Run("additional data overwrites existing keys", func(t *testing.T) {
		bytes, err := NewEventFromJSON("test.type", "test/source",
			`{"key":"original"}`,
			`{"key":"overwritten"}`,
		)
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		eventData := result["data"].(map[string]any)
		assert.Equal(t, "overwritten", eventData["key"])
	})

	t.Run("handles empty data strings", func(t *testing.T) {
		bytes, err := NewEventFromJSON("test.type", "test/source", "", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
	})

	t.Run("returns error on invalid JSON", func(t *testing.T) {
		_, err := NewEventFromJSON("test.type", "test/source", "not-json", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON")
	})

	t.Run("returns error on invalid additional JSON", func(t *testing.T) {
		_, err := NewEventFromJSON("test.type", "test/source", `{"key":"value"}`, "not-json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON")
	})
}
