package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type checkmarxoneExecuteScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newCheckmarxoneExecuteScanTestsUtils() checkmarxoneExecuteScanMockUtils {
	utils := checkmarxoneExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunCheckmarxoneExecuteScan(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := checkmarxOneExecuteScanOptions{}

		// test
		err := RunCheckmarxOneExecuteScan(config, nil)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := checkmarxOneExecuteScanOptions{}

		// test
		err := RunCheckmarxOneExecuteScan(config, nil)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
