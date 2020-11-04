package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type hadolintExecuteScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newHadolintExecuteScanTestsUtils() hadolintExecuteScanMockUtils {
	utils := hadolintExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunHadolintExecuteScan(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		// init
		// config := hadolintExecuteScanOptions{}

		// utils := newHadolintExecuteScanTestsUtils()
		// utils.AddFile("file.txt", []byte("dummy content"))

		// test
		// err := runHadolintExecuteScan(&config, &telemetry.CustomData{})

		// assert
		// assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		// init
		// config := hadolintExecuteScanOptions{}

		// utils := newHadolintExecuteScanTestsUtils()

		// test
		// err := runHadolintExecuteScan(&config, &telemetry.CustomData{})

		// assert
		// assert.EqualError(t, err, "cannot run without important file")
	})
}
