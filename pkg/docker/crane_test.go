package docker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/stretchr/testify/assert"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "HTTP/2 stream error NO_ERROR",
			err:      errors.New(`stream error: stream ID 19; NO_ERROR; received from peer`),
			expected: true,
		},
		{
			name:     "HTTP/2 stream error INTERNAL_ERROR",
			err:      errors.New(`stream error: stream ID 5; INTERNAL_ERROR`),
			expected: true,
		},
		{
			name:     "connection reset",
			err:      errors.New(`connection reset by peer`),
			expected: true,
		},
		{
			name:     "unexpected EOF",
			err:      errors.New(`unexpected EOF`),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      errors.New(`context deadline exceeded (Client.Timeout exceeded)`),
			expected: true,
		},
		{
			name:     "503 service unavailable",
			err:      fmt.Errorf("request failed: %w", &transport.Error{StatusCode: 503}),
			expected: true,
		},
		{
			name:     "502 bad gateway",
			err:      fmt.Errorf("request failed: %w", &transport.Error{StatusCode: 502}),
			expected: true,
		},
		{
			name:     "504 gateway timeout",
			err:      fmt.Errorf("request failed: %w", &transport.Error{StatusCode: 504}),
			expected: true,
		},
		{
			name:     "401 unauthorized - not retryable",
			err:      fmt.Errorf("request failed: %w", &transport.Error{StatusCode: 401}),
			expected: false,
		},
		{
			name:     "authentication error - not retryable",
			err:      errors.New(`UNAUTHORIZED: authentication required`),
			expected: false,
		},
		{
			name:     "not found error - not retryable",
			err:      errors.New(`NOT_FOUND: manifest unknown`),
			expected: false,
		},
		{
			name:     "permission denied - not retryable",
			err:      errors.New(`DENIED: requested access to the resource is denied`),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetryOperation(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		bundle := &craneUtilsBundle{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			BackoffFactor:  2.0,
		}

		attempts := 0
		err := bundle.retryOperation(context.Background(), "test", func() error {
			attempts++
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("retries on retryable error and succeeds", func(t *testing.T) {
		bundle := &craneUtilsBundle{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			BackoffFactor:  2.0,
		}

		attempts := 0
		err := bundle.retryOperation(context.Background(), "test", func() error {
			attempts++
			if attempts < 3 {
				return errors.New("stream error: stream ID 19; NO_ERROR; received from peer")
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("does not retry non-retryable error", func(t *testing.T) {
		bundle := &craneUtilsBundle{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			BackoffFactor:  2.0,
		}

		attempts := 0
		err := bundle.retryOperation(context.Background(), "test", func() error {
			attempts++
			return errors.New("UNAUTHORIZED: authentication required")
		})

		assert.Error(t, err)
		assert.Equal(t, 1, attempts)
		assert.Contains(t, err.Error(), "UNAUTHORIZED")
	})

	t.Run("fails after max retries", func(t *testing.T) {
		bundle := &craneUtilsBundle{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			BackoffFactor:  2.0,
		}

		attempts := 0
		err := bundle.retryOperation(context.Background(), "test", func() error {
			attempts++
			return errors.New("stream error: stream ID 19; NO_ERROR; received from peer")
		})

		assert.Error(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		bundle := &craneUtilsBundle{
			MaxRetries:     5,
			InitialBackoff: 100 * time.Millisecond,
			BackoffFactor:  2.0,
		}

		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := bundle.retryOperation(ctx, "test", func() error {
			attempts++
			return errors.New("stream error: retryable")
		})

		assert.Error(t, err)
		assert.True(t, attempts < 5, "should not complete all retries")
	})

	t.Run("uses default values when not configured", func(t *testing.T) {
		bundle := &craneUtilsBundle{}

		attempts := 0
		err := bundle.retryOperation(context.Background(), "test", func() error {
			attempts++
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})
}

func TestNewHTTPTransport(t *testing.T) {
	t.Run("HTTP/2 enabled by default", func(t *testing.T) {
		transport := newHTTPTransport(false)
		assert.NotNil(t, transport)
		assert.True(t, transport.ForceAttemptHTTP2)
	})

	t.Run("HTTP/2 disabled when requested", func(t *testing.T) {
		transport := newHTTPTransport(true)
		assert.NotNil(t, transport)
		assert.False(t, transport.ForceAttemptHTTP2)
	})

	t.Run("common settings", func(t *testing.T) {
		transport := newHTTPTransport(false)
		assert.Equal(t, 100, transport.MaxIdleConns)
		assert.Equal(t, 10, transport.MaxIdleConnsPerHost)
		assert.Equal(t, 90*time.Second, transport.IdleConnTimeout)
		assert.Equal(t, time.Duration(0), transport.ResponseHeaderTimeout)
		assert.NotNil(t, transport.TLSClientConfig)
	})
}
