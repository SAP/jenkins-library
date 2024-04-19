package events

import (
	"encoding/json"
	"testing"

	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/stretchr/testify/assert"
)

func TestPipelinerunFinishedEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// init
		// test
		event := NewPipelinerunFinishedEvent(&orchestrator.UnknownOrchestratorConfigProvider{}).Create()
		// asserts
		assert.Equal(t, "sap.hyperspace.pipelinerunFinished", event.cloudEvent.Type())
		assert.Equal(t, "/default/sap.hyperspace.piper", event.cloudEvent.Source())
		assert.Equal(t, "application/json", event.cloudEvent.DataContentType())

		var data map[string]interface{}
		err := json.Unmarshal(event.cloudEvent.DataEncoded, &data)
		assert.NoError(t, err)
		assert.Equal(t, "n/a", data["url"])
		assert.Equal(t, "n/a", data["repositoryUrl"])
		assert.Equal(t, "n/a", data["commitId"])
		assert.Equal(t, "FAILURE", data["outcome"])
	})
}
