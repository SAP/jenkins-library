//go:build unit

package gcs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_retryWithLogging(t *testing.T) {
	mLogger := &mockLogger{}

	t.Run("happy path, no retries", func(t *testing.T) {
		ctx := context.TODO()

		op := func(ctx context.Context) error {
			return nil
		}

		mLogger.On("Debugf", "Attempt %d/%d", mock.Anything, mock.Anything).Once().Return()

		err := retryWithLogging(ctx, mLogger, op, 1*time.Second, 3, 1)

		assert.NoError(t, err)
		mLogger.AssertExpectations(t)
	})

	t.Run("err on first attempt", func(t *testing.T) {
		ctx := context.TODO()

		attempt := 0
		task := func(ctx context.Context) error {
			if attempt == 1 {
				return nil
			}
			attempt++
			return errors.New("failed")
		}
		maxRetries := 3
		mLogger.On("Debugf", "Attempt %d/%d", []interface{}{1, maxRetries}).Return()
		mLogger.On("Debugf", "GCS client operation failed: %v", mock.Anything).Once().Return()
		mLogger.On("Debugf", "Attempt %d/%d", []interface{}{2, maxRetries}).Return()

		err := retryWithLogging(ctx, mLogger, task, 1*time.Second, maxRetries, 1)

		assert.NoError(t, err)
		mLogger.AssertExpectations(t)
	})

	t.Run("exhausting retries", func(t *testing.T) {
		ctx := context.TODO()

		attempt := 0
		op := func(ctx context.Context) error {
			if attempt >= maxRetries {
				return nil
			}
			attempt++
			return errors.New("failed")
		}

		maxRetries := 3

		mLogger.On("Debugf", "Attempt %d/%d", []interface{}{1, maxRetries}).Return()
		mLogger.On("Debugf", "Attempt %d/%d", []interface{}{2, maxRetries}).Return()
		mLogger.On("Debugf", "Attempt %d/%d", []interface{}{3, maxRetries}).Return()
		mLogger.On("Debugf", "GCS client operation failed: %v", mock.Anything).Return()

		err := retryWithLogging(ctx, mLogger, op, 1*time.Second, maxRetries, 1)

		assert.Error(t, err)
		assert.EqualError(t, err, "failed")
		mLogger.AssertExpectations(t)
	})
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)
}
