package log

import (
	"bytes"
	"github.com/stretchr/testify/assert"
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
}
