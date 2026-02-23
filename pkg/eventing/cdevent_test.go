//go:build unit

package eventing

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskRunFinishedCDEvent(t *testing.T) {
	t.Run("creates valid CDEvent as CloudEvent", func(t *testing.T) {
		bytes, err := NewTaskRunFinishedCDEvent("test/source", "myTask", "https://example.com/run/1", "success")
		assert.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		assert.Equal(t, "1.0", result["specversion"])
		assert.Contains(t, result["type"], "taskrun")
		assert.Contains(t, result["type"], "finished")
		assert.Equal(t, "test/source", result["source"])
		assert.NotEmpty(t, result["id"])
	})

	t.Run("outcome failure", func(t *testing.T) {
		bytes, err := NewTaskRunFinishedCDEvent("test/source", "myTask", "", "failure")
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
	})
}
