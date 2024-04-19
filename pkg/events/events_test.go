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
		event := Event{}.Create(mock.Anything, nil)
		// asserts
		assert.Equal(t, mock.Anything, event.cloudEvent.Type())
		assert.Equal(t, "/default/sap.hyperspace.piper", event.cloudEvent.Source())
	})
}
