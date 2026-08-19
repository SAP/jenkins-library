//go:build unit

package eventing

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskRunFinishedCDEvent(t *testing.T) {
	t.Run("creates valid CDEvent as CloudEvent", func(t *testing.T) {
		bytes, err := newTaskRunFinishedCDEvent("test/source", "myTask", "https://example.com/run/1", "success", "Build")
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		assert.Equal(t, "1.0", result["specversion"])
		assert.Contains(t, result["type"], "taskrun")
		assert.Contains(t, result["type"], "finished")
		assert.Equal(t, "test/source", result["source"])
		assert.NotEmpty(t, result["id"])
	})

	t.Run("includes stageName in customData", func(t *testing.T) {
		bytes, err := newTaskRunFinishedCDEvent("test/source", "myTask", "", "success", "Build")
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		data, ok := result["data"].(map[string]any)
		assert.True(t, ok)
		customData, ok := data["customData"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "Build", customData["stageName"])
	})

	t.Run("no customData when stageName is empty", func(t *testing.T) {
		bytes, err := newTaskRunFinishedCDEvent("test/source", "myTask", "", "success", "")
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		data, ok := result["data"].(map[string]any)
		assert.True(t, ok)
		_, hasCustomData := data["customData"]
		assert.False(t, hasCustomData)
	})

	t.Run("outcome failure", func(t *testing.T) {
		bytes, err := newTaskRunFinishedCDEvent("test/source", "myTask", "", "failure", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
	})
}

func TestNewPipelineRunStartedCDEvent(t *testing.T) {
	t.Run("creates valid CDEvent as CloudEvent", func(t *testing.T) {
		bytes, err := newPipelineRunStartedCDEvent("test/source", "myPipeline", "https://example.com/run/1")
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		assert.Equal(t, "1.0", result["specversion"])
		assert.Contains(t, result["type"], "pipelinerun")
		assert.Contains(t, result["type"], "started")
		assert.Equal(t, "test/source", result["source"])
		assert.NotEmpty(t, result["id"])
	})

	t.Run("empty pipeline URL", func(t *testing.T) {
		bytes, err := newPipelineRunStartedCDEvent("test/source", "myPipeline", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
	})
}
