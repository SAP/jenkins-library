package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
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
		config := checkmarxoneExecuteScanOptions{}

		utils := newCheckmarxoneExecuteScanTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runCheckmarxoneExecuteScan(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := checkmarxoneExecuteScanOptions{}

		utils := newCheckmarxoneExecuteScanTestsUtils()

		// test
		err := runCheckmarxoneExecuteScan(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
