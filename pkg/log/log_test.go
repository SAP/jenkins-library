package log

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSecrets(t *testing.T) {
	t.Run("should log", func(t *testing.T) {
		secret := "password"

		outWriter := Entry().Logger.Out
		var buffer bytes.Buffer
		Entry().Logger.SetOutput(&buffer)
		defer func() { Entry().Logger.SetOutput(outWriter) }()

		Entry().Infof("My secret is %s.", secret)
		assert.Contains(t, buffer.String(), secret)

		buffer.Reset()
		RegisterSecret(secret)
		Entry().Infof("My secret is %s.", secret)
		assert.NotContains(t, buffer.String(), secret)
	})
}

func TestWriteLargeBuffer(t *testing.T) {
	t.Run("should log any buffer size without linebreaks", func(t *testing.T) {
		b := []byte{0}
		size := 131072
		b = bytes.Repeat(b, size)
		written, err := Writer().Write(b)
		assert.Equal(t, size, written)
		assert.NoError(t, err)
	})
}
