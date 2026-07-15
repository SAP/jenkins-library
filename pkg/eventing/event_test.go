//go:build unit

package eventing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	t.Run("creates valid CloudEvent with map data", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		event, err := newEvent("test.type", "test/source", data)
		assert.NoError(t, err)

		assert.Equal(t, "1.0", event.SpecVersion())
		assert.Equal(t, "test.type", event.Type())
		assert.Equal(t, "test/source", event.Source())
		assert.NotEmpty(t, event.ID())
		assert.NotEmpty(t, event.Time())

		var eventData map[string]any
		assert.NoError(t, event.DataAs(&eventData))
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
		event, err := newEvent("test.type", "test/source", payload)
		assert.NoError(t, err)

		var eventData map[string]any
		assert.NoError(t, event.DataAs(&eventData))
		assert.Equal(t, "build", eventData["taskName"])
		assert.Equal(t, "dev", eventData["stageName"])
		assert.Equal(t, "success", eventData["outcome"])
	})

	t.Run("handles nil data", func(t *testing.T) {
		event, err := newEvent("test.type", "test/source", nil)
		assert.NoError(t, err)
		assert.Equal(t, "test.type", event.Type())
	})
}

func TestNewEventFromJSON(t *testing.T) {
	t.Run("creates event from JSON data", func(t *testing.T) {
		event, err := NewEventFromJSON("test.type", "test/source", `{"key":"value"}`, "")
		assert.NoError(t, err)
		assert.Equal(t, "test.type", event.Type())

		var eventData map[string]any
		assert.NoError(t, event.DataAs(&eventData))
		assert.Equal(t, "value", eventData["key"])
	})

	t.Run("merges additional JSON data", func(t *testing.T) {
		event, err := NewEventFromJSON("test.type", "test/source",
			`{"key":"value"}`,
			`{"extra":"data"}`,
		)
		assert.NoError(t, err)

		var eventData map[string]any
		assert.NoError(t, event.DataAs(&eventData))
		assert.Equal(t, "value", eventData["key"])
		assert.Equal(t, "data", eventData["extra"])
	})

	t.Run("additional data overwrites existing keys", func(t *testing.T) {
		event, err := NewEventFromJSON("test.type", "test/source",
			`{"key":"original"}`,
			`{"key":"overwritten"}`,
		)
		assert.NoError(t, err)

		var eventData map[string]any
		assert.NoError(t, event.DataAs(&eventData))
		assert.Equal(t, "overwritten", eventData["key"])
	})

	t.Run("handles empty data strings", func(t *testing.T) {
		event, err := NewEventFromJSON("test.type", "test/source", "", "")
		assert.NoError(t, err)
		assert.Equal(t, "test.type", event.Type())
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
