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
		bytes, err := NewEvent("test.type", "test/source", data)
		assert.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)
		assert.Equal(t, "1.0", result["specversion"])
		assert.Equal(t, "test.type", result["type"])
		assert.Equal(t, "test/source", result["source"])
		assert.NotEmpty(t, result["id"])
		assert.NotEmpty(t, result["time"])

		eventData := result["data"].(map[string]interface{})
		assert.Equal(t, "value", eventData["key"])
	})

	t.Run("creates valid CloudEvent with struct data", func(t *testing.T) {
		payload := TaskRunFinishedPayload{
			TaskName:  "build",
			StageName: "dev",
			Outcome:   "success",
		}
		bytes, err := NewEvent("test.type", "test/source", payload)
		assert.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		eventData := result["data"].(map[string]interface{})
		assert.Equal(t, "build", eventData["taskName"])
		assert.Equal(t, "dev", eventData["stageName"])
		assert.Equal(t, "success", eventData["outcome"])
	})

	t.Run("handles nil data", func(t *testing.T) {
		bytes, err := NewEvent("test.type", "test/source", nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
	})
}