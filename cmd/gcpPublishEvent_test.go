package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type gcpPublishEventMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newGcpPublishEventTestsUtils() gcpPublishEventMockUtils {
	utils := gcpPublishEventMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunGcpPublishEvent(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := gcpPublishEventOptions{}

		// test
		err := runGcpPublishEvent(&config, nil)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := gcpPublishEventOptions{}

		// test
		err := runGcpPublishEvent(&config, nil)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
