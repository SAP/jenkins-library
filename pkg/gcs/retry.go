package gcs

import (
	"context"
	"strings"
	"time"
)

// Retry configuration
const (
	maxRetries      = 5
	initialBackoff  = 5 * time.Second
	retryMultiplier = 2
)

const noRetryError = "is under active Temporary hold and cannot be deleted, overwritten or archived until hold is removed"

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

		if strings.Contains(err.Error(), noRetryError) {
			return err
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
