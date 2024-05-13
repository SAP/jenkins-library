package events

import (
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// init
		ec := cloudevents.EventContextV1{}
		opts := []Option{WithID(mock.Anything)}
		// test
		for _, applyOpt := range opts {
			applyOpt(&ec)
		}
		// asserts
		assert.Equal(t, mock.Anything, ec.GetID())
	})
}
