//go:build unit

package eventing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskRunFinishedCDEvent(t *testing.T) {
	t.Run("creates valid CDEvent as CloudEvent", func(t *testing.T) {
		event, err := newTaskRunFinishedCDEvent("test/source", "myTask", "https://example.com/run/1", "success", "Build")
		assert.NoError(t, err)

		assert.Equal(t, "1.0", event.SpecVersion())
		assert.Contains(t, event.Type(), "taskrun")
		assert.Contains(t, event.Type(), "finished")
		assert.Equal(t, "test/source", event.Source())
		assert.NotEmpty(t, event.ID())
	})

	t.Run("includes stageName in customData", func(t *testing.T) {
		event, err := newTaskRunFinishedCDEvent("test/source", "myTask", "", "success", "Build")
		assert.NoError(t, err)

		var data map[string]any
		assert.NoError(t, event.DataAs(&data))
		customData, ok := data["customData"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "Build", customData["stageName"])
	})

	t.Run("no customData when stageName is empty", func(t *testing.T) {
		event, err := newTaskRunFinishedCDEvent("test/source", "myTask", "", "success", "")
		assert.NoError(t, err)

		var data map[string]any
		assert.NoError(t, event.DataAs(&data))
		_, hasCustomData := data["customData"]
		assert.False(t, hasCustomData)
	})

	t.Run("outcome failure", func(t *testing.T) {
		event, err := newTaskRunFinishedCDEvent("test/source", "myTask", "", "failure", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, event.ID())
	})
}

func TestNewPipelineRunStartedCDEvent(t *testing.T) {
	t.Run("creates valid CDEvent as CloudEvent", func(t *testing.T) {
		event, err := newPipelineRunStartedCDEvent("test/source", "myPipeline", "https://example.com/run/1")
		assert.NoError(t, err)

		assert.Equal(t, "1.0", event.SpecVersion())
		assert.Contains(t, event.Type(), "pipelinerun")
		assert.Contains(t, event.Type(), "started")
		assert.Equal(t, "test/source", event.Source())
		assert.NotEmpty(t, event.ID())
	})

	t.Run("empty pipeline URL", func(t *testing.T) {
		event, err := newPipelineRunStartedCDEvent("test/source", "myPipeline", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, event.ID())
	})
}
