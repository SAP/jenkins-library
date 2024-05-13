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

	t.Run("AddToCloudEventData", func(t *testing.T) {
		// init
		additionalData := `{"additionalKey": "additionalValue"}`
		// test
		event := NewEvent(mock.Anything, mock.Anything).CreateWithJSONData([]byte(`{"mockKey": "mockValue"}`), additionalData)
		// asserts
		assert.Equal(t, mock.Anything, event.cloudEvent.Type())
		assert.Equal(t, mock.Anything, event.cloudEvent.Source())
		assert.Contains(t, string(event.cloudEvent.Data()), `"additionalKey":"additionalValue"`)
	})

}
