package cmd

import (
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

func (g *mockGcpPublishEventUtilsBundle) GetOIDCTokenByValidation(roleID string) (string, error) {
	return "testOIDCtoken123", nil
}

func (g *mockGcpPublishEventUtilsBundle) GetFederatedToken(projectNumber, pool, provider, token string) (string, error) {
	return "testFederatedToken123", nil
}

func (g *mockGcpPublishEventUtilsBundle) Publish(projectNumber string, topic string, token string, data []byte) error {
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
