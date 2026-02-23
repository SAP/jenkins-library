//go:build unit

package cmd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPubsubClient struct {
	publishErr error
}

func (p *mockPubsubClient) Publish(_ string, _ []byte) error {
	return p.publishErr
}

func TestRunGcpPublishEvent(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		publisher := &mockPubsubClient{}
		cfg := &gcpPublishEventOptions{
			EventType:   "PipelineRunStarted",
			EventSource: "unittest",
			Topic:       "test-topic",
		}

		err := runGcpPublishEvent(publisher, cfg)
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		publisher := &mockPubsubClient{publishErr: errors.New("failed to send request")}
		cfg := &gcpPublishEventOptions{
			EventType:   "PipelineRunStarted",
			EventSource: "unittest",
			Topic:       "test-topic",
		}

		err := runGcpPublishEvent(publisher, cfg)
		assert.EqualError(t, err, "failed to publish event: failed to send request")
	})
}