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

func TestNewCustomTypeCDEventFromJSON(t *testing.T) {
	const validEventType = "dev.cdeventsx.piper-pipeline.runstarted.0.1.0"

	t.Run("creates valid CDEvent as CloudEvent", func(t *testing.T) {
		bytes, err := NewCustomTypeCDEventFromJSON(validEventType, "test/source", "my-run-id", `{"key":"value"}`, "")
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		assert.Equal(t, "1.0", result["specversion"])
		assert.Equal(t, validEventType, result["type"])
		assert.Equal(t, "test/source", result["source"])
		assert.NotEmpty(t, result["id"])

		data, ok := result["data"].(map[string]any)
		assert.True(t, ok)
		customData, ok := data["customData"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "value", customData["key"])
		assert.Equal(t, "application/json", data["customDataContentType"])
	})

	t.Run("subject.id is set to the provided subjectID", func(t *testing.T) {
		bytes, err := NewCustomTypeCDEventFromJSON(validEventType, "test/source", "pipeline-run-42", "", "")
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		data, ok := result["data"].(map[string]any)
		assert.True(t, ok)
		subject, ok := data["subject"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "pipeline-run-42", subject["id"])
	})

	t.Run("merges additional JSON data", func(t *testing.T) {
		bytes, err := NewCustomTypeCDEventFromJSON(validEventType, "test/source", "my-run-id",
			`{"key":"value"}`,
			`{"extra":"data"}`,
		)
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		data := result["data"].(map[string]any)
		customData := data["customData"].(map[string]any)
		assert.Equal(t, "value", customData["key"])
		assert.Equal(t, "data", customData["extra"])
	})

	t.Run("additional data overwrites existing keys", func(t *testing.T) {
		bytes, err := NewCustomTypeCDEventFromJSON(validEventType, "test/source", "my-run-id",
			`{"key":"original"}`,
			`{"key":"overwritten"}`,
		)
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		data := result["data"].(map[string]any)
		customData := data["customData"].(map[string]any)
		assert.Equal(t, "overwritten", customData["key"])
	})

	t.Run("no customData when both inputs are empty", func(t *testing.T) {
		bytes, err := NewCustomTypeCDEventFromJSON(validEventType, "test/source", "my-run-id", "", "")
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		data, ok := result["data"].(map[string]any)
		assert.True(t, ok)
		_, hasCustomData := data["customData"]
		assert.False(t, hasCustomData)
	})

	t.Run("only additionalEventData set produces customData", func(t *testing.T) {
		bytes, err := NewCustomTypeCDEventFromJSON(validEventType, "test/source", "my-run-id", "", `{"only":"additional"}`)
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)

		data, ok := result["data"].(map[string]any)
		assert.True(t, ok)
		customData, ok := data["customData"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "additional", customData["only"])
	})

	t.Run("returns error on invalid CDEvent type string", func(t *testing.T) {
		_, err := NewCustomTypeCDEventFromJSON("not-a-cdevent-type", "test/source", "my-run-id", "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid CDEvent type")
	})

	t.Run("returns error on invalid eventData JSON", func(t *testing.T) {
		_, err := NewCustomTypeCDEventFromJSON(validEventType, "test/source", "my-run-id", "not-json", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON")
	})

	t.Run("returns error on invalid additionalEventData JSON", func(t *testing.T) {
		_, err := NewCustomTypeCDEventFromJSON(validEventType, "test/source", "my-run-id", `{"key":"value"}`, "not-json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON")
	})
}
