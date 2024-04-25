//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// type gcpPublishEventMockUtils struct {
// 	*mock.ExecMockRunner
// 	*mock.FilesMock
// }

// func newGcpPublishEventTestsUtils() gcpPublishEventMockUtils {
// 	utils := gcpPublishEventMockUtils{
// 		ExecMockRunner: &mock.ExecMockRunner{},
// 		FilesMock:      &mock.FilesMock{},
// 	}
// 	return utils
// }

type mockGcpPublishEventUtilsBundle struct {
	config *gcpPublishEventOptions
}

func (g *mockGcpPublishEventUtilsBundle) GetConfig() *gcpPublishEventOptions {
	return g.config
}

func (g *mockGcpPublishEventUtilsBundle) GetOIDCTokenByValidation(roleID string) (string, error) {
	return "testOIDCtoken123", nil
}

func TestRunGcpPublishEvent(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		mock := &mockGcpPublishEventUtilsBundle{}

		// test
		err := runGcpPublishEvent(mock)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		mock := &mockGcpPublishEventUtilsBundle{}

		// test
		err := runGcpPublishEvent(mock)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
