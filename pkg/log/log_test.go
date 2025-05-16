//go:build unit

package log

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
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

	t.Run("should log url encoded", func(t *testing.T) {
		secret := "secret-token!0"
		encodedSecret := url.QueryEscape(secret)

		outWriter := Entry().Logger.Out
		var buffer bytes.Buffer
		Entry().Logger.SetOutput(&buffer)
		defer func() { Entry().Logger.SetOutput(outWriter) }()

		Entry().Infof("My secret is %s.", secret)
		assert.Contains(t, buffer.String(), secret)

		buffer.Reset()
		RegisterSecret(secret)
		Entry().Infof("My secret is %s.", encodedSecret)
		assert.NotContains(t, buffer.String(), encodedSecret)
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
}
