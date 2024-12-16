//go:build unit

package cmd

import (
	"github.com/SAP/jenkins-library/pkg/gcp"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockGcpPublishEventUtilsBundle struct {
	config *gcpPublishEventOptions
}

func (g *mockGcpPublishEventUtilsBundle) GetConfig() *gcpPublishEventOptions {
	return g.config
}

func (g *mockGcpPublishEventUtilsBundle) NewPubsubClient(_, _, _, _, _ string) gcp.PubsubClient {
	return &mockPubsubClient{}
}

type mockPubsubClient struct {
}

func (p *mockPubsubClient) Publish(topic string, _ []byte) error {
	if topic == "goodTestCase" {
		return nil
	} else if topic == "badTestCase" {
		return errors.New("failed to send request")
	}
	return nil
}

func TestRunGcpPublishEvent(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		mock := &mockGcpPublishEventUtilsBundle{
			config: &gcpPublishEventOptions{
				EventType:   "PipelineRunStarted",
				EventSource: "unittest",
				Topic:       "goodTestCase",
			}}

		// test
		err := runGcpPublishEvent(mock)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		mock := &mockGcpPublishEventUtilsBundle{
			config: &gcpPublishEventOptions{
				EventType:   "PipelineRunStarted",
				EventSource: "unittest",
				Topic:       "badTestCase",
			}}

		// test
		err := runGcpPublishEvent(mock)

		// assert
		assert.EqualError(t, err, "failed to publish event: failed to send request")
	})
}
