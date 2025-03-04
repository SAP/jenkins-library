package gcs

import (
	"context"
	"time"
)

// Retry configuration
const (
	maxRetries      = 5
	initialBackoff  = 5 * time.Second
	maxRetryPeriod  = 22 * time.Second
	retryMultiplier = 2
)

type debugLogger interface {
	Debugf(format string, args ...interface{})
}

func retryWithLogging(
	ctx context.Context,
	log debugLogger,
	taskFn func(ctx context.Context) error,
	initialBackoff time.Duration,
	maxRetries int,
	retryMult float64,
) error {
	backoff := initialBackoff
	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Debugf("Attempt %d/%d", attempt, maxRetries)
		if err = taskFn(ctx); err == nil {
			return nil
		}

		log.Debugf("GCS client operation failed: %v", err)
		if attempt >= maxRetries {
			return err
		}

		time.Sleep(backoff)
		backoff *= time.Duration(retryMult)
	}
	return nil
}
