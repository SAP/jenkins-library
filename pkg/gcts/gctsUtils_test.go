package gcts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsUtils(t *testing.T) {

	t.Run("valid proxy is used in config", func(t *testing.T) {
		options, err := NewHttpClientOptions("password", "userName", "https://example.org/my-proxy", false)
		assert.NoError(t, err)
		assert.Equal(t, options.TransportProxy.String(), "https://example.org/my-proxy")
	})

	t.Run("no transport proxy set when proxy is not defined", func(t *testing.T) {
		options, err := NewHttpClientOptions("password", "userName", "", false)
		assert.NoError(t, err)
		assert.Nil(t, options.TransportProxy)
	})

	t.Run("error raised when transport proxy url is invalid", func(t *testing.T) {
		_, err := NewHttpClientOptions("password", "userName", "invalid\n url", false)
		assert.ErrorContains(t, err, "parsing proxy-url failed")
	})
}
