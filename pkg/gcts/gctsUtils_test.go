package gcts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGctsUtils(t *testing.T) {

	t.Run("parameters are correctly assigned to client options", func(t *testing.T) {
		username := "testUser"
		password := "testPassword"
		proxy := "https://proxy.example.com:8080"
		skipSSLVerification := true

		options, err := NewHttpClientOptions(username, password, proxy, skipSSLVerification)
		
		assert.NoError(t, err)
		assert.Equal(t, username, options.Username)
		assert.Equal(t, password, options.Password)
		assert.Equal(t, skipSSLVerification, options.TransportSkipVerification)
		assert.NotNil(t, options.TransportProxy)
		assert.Equal(t, proxy, options.TransportProxy.String())
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
