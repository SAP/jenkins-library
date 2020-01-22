package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataToMap(t *testing.T) {
	testData := Data{BaseData: BaseData{ActionName: "testAction"}}

	result := testData.toMap()

	assert.Contains(t, result, "actionName")
	assert.Contains(t, result, "idsite")
}

func TestDataToPayload(t *testing.T) {
	t.Run("with single parameter", func(t *testing.T) {
		testData := Data{BaseData: BaseData{ActionName: "testAction"}}

		result := testData.toPayloadString()

		assert.Contains(t, result, "actionName=testAction")
		assert.NotContains(t, result, "idsite=")
	})

	t.Run("with multiple parameters", func(t *testing.T) {
		testData := Data{BaseData: BaseData{ActionName: "testAction", SiteID: "gl8rkd6j211bw3j1fwb8rb4h0000gn"}}

		result := testData.toPayloadString()

		assert.Contains(t, result, "&")
		assert.Contains(t, result, "actionName=testAction")
		assert.Contains(t, result, "idsite=gl8rkd6j211bw3j1fwb8rb4h0000gn")
	})
}
