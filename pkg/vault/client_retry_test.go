//go:build unit

package vault

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vaultAPI "github.com/hashicorp/vault/api"
)

func newRetryCheckFunc(t *testing.T) func(ctx context.Context, resp *http.Response, err error) (bool, error) {
	t.Helper()
	apiClient, err := vaultAPI.NewClient(vaultAPI.DefaultConfig())
	require.NoError(t, err)
	applyApiClientRetryConfiguration(apiClient)
	return apiClient.CheckRetry()
}

func TestCheckRetryOriginalErrorIsLogged(t *testing.T) {
	_, hook := logtest.NewNullLogger()
	log.RegisterHook(hook)
	t.Cleanup(func() { hook.Reset() })

	checkRetry := newRetryCheckFunc(t)

	originalErr := errors.New("connection reset by peer")
	resp := &http.Response{Status: "503 Service Unavailable", StatusCode: 503, Body: http.NoBody}

	retry, err := checkRetry(context.Background(), resp, originalErr)

	assert.True(t, retry)
	assert.NoError(t, err)

	// The original error must appear in the Info log message, not be shadowed
	// by the retryPolicyErr returned from vaultAPI.DefaultRetryPolicy.
	var found bool
	for _, entry := range hook.Entries {
		if entry.Level == logrus.InfoLevel && entry.Message != "" {
			if strings.Contains(entry.Message, originalErr.Error()) {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "expected original error %q to be logged; log entries: %v", originalErr, logEntryMessages(hook.Entries))
}

func TestCheckRetryNoResponseLogsOriginalError(t *testing.T) {
	_, hook := logtest.NewNullLogger()
	log.RegisterHook(hook)
	t.Cleanup(func() { hook.Reset() })

	checkRetry := newRetryCheckFunc(t)

	originalErr := errors.New("dial tcp: connection refused")
	retry, err := checkRetry(context.Background(), nil, originalErr)

	assert.True(t, retry)
	assert.NoError(t, err)

	var found bool
	for _, entry := range hook.Entries {
		if entry.Level == logrus.InfoLevel && strings.Contains(entry.Message, originalErr.Error()) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected original error %q to be logged; log entries: %v", originalErr, logEntryMessages(hook.Entries))
}

func TestCheckRetryEOFTriggersRetry(t *testing.T) {
	_, hook := logtest.NewNullLogger()
	log.RegisterHook(hook)
	t.Cleanup(func() { hook.Reset() })

	checkRetry := newRetryCheckFunc(t)

	eofErr := errors.New("unexpected EOF")
	retry, err := checkRetry(context.Background(), nil, eofErr)

	assert.True(t, retry)
	assert.NoError(t, err)
}

func TestCheckRetrySuccessDoesNotRetry(t *testing.T) {
	_, hook := logtest.NewNullLogger()
	log.RegisterHook(hook)
	t.Cleanup(func() { hook.Reset() })

	checkRetry := newRetryCheckFunc(t)

	resp := &http.Response{Status: "200 OK", StatusCode: 200, Body: http.NoBody}
	retry, err := checkRetry(context.Background(), resp, nil)

	assert.False(t, retry)
	assert.NoError(t, err)

	for _, entry := range hook.Entries {
		if entry.Level == logrus.InfoLevel {
			assert.NotContains(t, entry.Message, "Retrying vault request", "unexpected retry log on 200 response")
		}
	}
}

func logEntryMessages(entries []logrus.Entry) []string {
	msgs := make([]string, len(entries))
	for i, e := range entries {
		msgs[i] = fmt.Sprintf("[%s] %s", e.Level, e.Message)
	}
	return msgs
}
