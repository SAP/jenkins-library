//go:build unit

package whitesource

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockUntilProjectIsUpdated(t *testing.T) {
	t.Parallel()

	nowString := "2010-05-30 00:15:00 +0100"
	now, err := time.Parse(DateTimeLayout, nowString)
	require.NoError(t, err)
	options := pollOptions{
		scanTime:         now,
		maxAge:           2 * time.Second,
		timeBetweenPolls: 1 * time.Second,
		maxWaitTime:      1 * time.Second,
	}

	t.Run("already new enough", func(t *testing.T) {
		// init
		lastUpdatedDate := "2010-05-30 00:15:01 +0100"
		systemMock := NewSystemMock(lastUpdatedDate)
		// test
		err = blockUntilProjectIsUpdated(systemMock.Projects[0].Token, systemMock, options)
		// assert
		assert.NoError(t, err)
	})
	t.Run("timeout while polling", func(t *testing.T) {
		// init
		lastUpdatedDate := "2010-05-30 00:07:00 +0100"
		systemMock := NewSystemMock(lastUpdatedDate)
		// test
		err = blockUntilProjectIsUpdated(systemMock.Projects[0].Token, systemMock, options)
		// assert
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "timeout while waiting")
		}
	})
	t.Run("timeout while polling, no update time", func(t *testing.T) {
		// init
		systemMock := NewSystemMock("")
		// test
		err = blockUntilProjectIsUpdated(systemMock.Projects[0].Token, systemMock, options)
		// assert
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "timeout while waiting")
		}
	})
}
