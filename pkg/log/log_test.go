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
	t.Run("should log large buffer without linebreaks via Writer()", func(t *testing.T) {
		b := []byte("a")
		size := 131072
		b = bytes.Repeat(b, size)
		written, err := Writer().Write(b)
		assert.Equal(t, size, written)
		assert.NoError(t, err)
	})
	t.Run("fails writing large buffer without linebreaks via Entry().Writer()", func(t *testing.T) {
		// NOTE: If this test starts failing, then the logrus issue has been fixed and
		// the work-around with Writer() can be removed.
		b := []byte("a")
		size := 131072
		b = bytes.Repeat(b, size)
		written, err := Entry().Writer().Write(b)
		assert.Error(t, err)
		assert.True(t, size != written)
	})
}
