package log

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestSecrets(t *testing.T) {
	t.Run("should log", func(t *testing.T) {
		secret := "password"

		var buffer bytes.Buffer
		Entry().Logger.SetOutput(&buffer)
		Entry().Infof("My secret is %s.", secret)
		assert.Contains(t, buffer.String(), secret)

		buffer.Reset()
		RegisterSecret(secret)
		Entry().Infof("My secret is %s.", secret)
		assert.NotContains(t, buffer.String(), secret)
	})

	t.Run("should log message only", func(t *testing.T) {

		SetFormatter(true)

		message := "This is a log message"

		var buffer bytes.Buffer
		Entry().Logger.SetOutput(&buffer)
		Entry().Infof(message)
		assert.Equal(t, message, strings.TrimSpace(buffer.String()))
	})
}
