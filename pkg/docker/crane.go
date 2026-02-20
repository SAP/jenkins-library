package docker

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/SAP/jenkins-library/pkg/log"
)

// Retry configuration for transient network errors
const (
	defaultMaxRetries     = 3
	defaultInitialBackoff = 5 * time.Second
	defaultBackoffFactor  = 2.0
)

type CraneUtilsBundle struct {
	MaxRetries     int
	InitialBackoff time.Duration
	BackoffFactor  float64
}

// newHTTPTransport creates an HTTP transport optimized for large file transfers
// with settings to mitigate HTTP/2 stream errors
func newHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 0,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
}

// isRetryableError checks if the error is transient and should be retried
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// HTTP/2 stream errors
	if strings.Contains(errMsg, "stream error") {
		return true
	}
	// Connection reset errors
	if strings.Contains(errMsg, "connection reset") {
		return true
	}
	// EOF during transfer
	if strings.Contains(errMsg, "unexpected EOF") {
		return true
	}
	// Timeout errors
	if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "Timeout") {
		return true
	}
	// Generic network errors
	if strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection refused") {
		return true
	}
	// Server temporarily unavailable
	if strings.Contains(errMsg, "503") || strings.Contains(errMsg, "502") || strings.Contains(errMsg, "504") {
		return true
	}
	return false
}

// retryOperation executes an operation with exponential backoff retry logic
func (c *CraneUtilsBundle) retryOperation(ctx context.Context, operation string, fn func() error) error {
	maxRetries := c.MaxRetries
	if maxRetries <= 0 {
		maxRetries = defaultMaxRetries
	}
	backoff := c.InitialBackoff
	if backoff <= 0 {
		backoff = defaultInitialBackoff
	}
	factor := c.BackoffFactor
	if factor <= 0 {
		factor = defaultBackoffFactor
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if !isRetryableError(lastErr) {
			log.Entry().Debugf("%s: non-retryable error: %v", operation, lastErr)
			return lastErr
		}

		if attempt >= maxRetries {
			log.Entry().Warnf("%s: all %d attempts failed, last error: %v", operation, maxRetries, lastErr)
			return lastErr
		}

		log.Entry().Warnf("%s: attempt %d/%d failed with retryable error: %v, retrying in %v...",
			operation, attempt, maxRetries, lastErr, backoff)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		backoff = time.Duration(float64(backoff) * factor)
	}
	return lastErr
}

// getCraneOptions returns common crane options with custom transport
func getCraneOptions(ctx context.Context, platform *v1.Platform) []crane.Option {
	opts := []crane.Option{
		crane.WithContext(ctx),
		crane.WithTransport(newHTTPTransport()),
	}
	if platform != nil {
		opts = append(opts, crane.WithPlatform(platform))
	}
	return opts
}

func (c *CraneUtilsBundle) CopyImage(ctx context.Context, src, dest, platform string) error {
	p, err := parsePlatform(platform)
	if err != nil {
		return err
	}
	return c.retryOperation(ctx, "CopyImage", func() error {
		return crane.Copy(src, dest, getCraneOptions(ctx, p)...)
	})
}

func (c *CraneUtilsBundle) PushImage(ctx context.Context, im v1.Image, dest, platform string) error {
	p, err := parsePlatform(platform)
	if err != nil {
		return err
	}
	return c.retryOperation(ctx, "PushImage", func() error {
		return crane.Push(im, dest, getCraneOptions(ctx, p)...)
	})
}

func (c *CraneUtilsBundle) LoadImage(ctx context.Context, src string) (v1.Image, error) {
	var img v1.Image
	err := c.retryOperation(ctx, "LoadImage", func() error {
		var loadErr error
		img, loadErr = crane.Load(src, crane.WithContext(ctx))
		return loadErr
	})
	return img, err
}

// parsePlatform is a wrapper for v1.ParsePlatform. It is necessary because
// v1.ParsePlatform returns an empty struct when the platform is equal to an empty string,
// whereas we expect 'nil'
func parsePlatform(p string) (*v1.Platform, error) {
	if p == "" {
		return nil, nil
	}
	return v1.ParsePlatform(p)
}
