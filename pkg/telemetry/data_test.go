package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataToMap(t *testing.T) {
	// init
	testData := Data{BaseData: BaseData{ActionName: "testAction"}, CustomData: CustomData{Custom2Label: "label", Custom2: "value"}}
	// test
	result := testData.toMap()
	// assert
	assert.Contains(t, result, "action_name")
	assert.Contains(t, result, "event_type")
	assert.Contains(t, result, "idsite")
	assert.Contains(t, result, "url")

	assert.Contains(t, result, "e_3")
	assert.Contains(t, result, "e_4")
	assert.Contains(t, result, "e_5")
	assert.Contains(t, result, "e_10")

	assert.Contains(t, result, "custom3")
	assert.Contains(t, result, "custom4")
	assert.Contains(t, result, "custom5")
	assert.Contains(t, result, "custom10")

	assert.Contains(t, result, "e_27")
	assert.Contains(t, result, "custom27")

	assert.Equal(t, 15, len(result))
}

func TestDataToPayload(t *testing.T) {
	t.Run("with single parameter", func(t *testing.T) {
		// init
		testData := Data{BaseData: BaseData{ActionName: "testAction"}}
		// test
		result := testData.toPayloadString()
		// assert
		assert.Contains(t, result, "action_name=testAction")
		assert.NotContains(t, result, "idsite=")
	})

	t.Run("with multiple parameters", func(t *testing.T) {
		// init
		testData := Data{BaseData: BaseData{ActionName: "testAction", SiteID: "gl8rkd6j211bw3j1fwb8rb4h0000gn"}}
		// test
		result := testData.toPayloadString()
		// assert
		assert.Contains(t, result, "&")
		assert.Contains(t, result, "action_name=testAction")
		assert.Contains(t, result, "idsite=gl8rkd6j211bw3j1fwb8rb4h0000gn")
	})

	t.Run("encoding", func(t *testing.T) {
		// init
		testData := Data{BaseData: BaseData{ActionName: "t€štÄçtïøñ"}}
		// test
		result := testData.toPayloadString()
		// assert
		assert.Contains(t, result, "t%E2%82%AC%C5%A1t%C3%84%C3%A7t%C3%AF%C3%B8%C3%B1")
		assert.NotContains(t, result, "t€štÄçtïøñ")
	})
}
