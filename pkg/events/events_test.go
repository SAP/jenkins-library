package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEventCreation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// init
		// test
		event := NewEvent(mock.Anything, mock.Anything).Create(nil)
		// asserts
		assert.Equal(t, mock.Anything, event.cloudEvent.Type())
		assert.Equal(t, mock.Anything, event.cloudEvent.Source())
	})

	t.Run("CreateWithJSONData - no additional data", func(t *testing.T) {
		// init
		testData := `{"testKey":"testValue"}`
		// test
		event, err := NewEvent(mock.Anything, mock.Anything).CreateWithJSONData(testData)
		// asserts
		assert.NoError(t, err)
		assert.Equal(t, string(event.cloudEvent.Data()), testData)
	})

	t.Run("CreateWithJSONData + AddToCloudEventData", func(t *testing.T) {
		// init
		testData := `{"testKey": "testValue"}`
		additionalData := `{"additionalKey": "additionalValue"}`
		// test
		event, err := NewEvent(mock.Anything, mock.Anything).CreateWithJSONData(testData)
		event.AddToCloudEventData(additionalData)
		// asserts
		assert.NoError(t, err)
		assert.Equal(t, string(event.cloudEvent.Data()), `{"additionalKey":"additionalValue","testKey":"testValue"}`)
	})
}
