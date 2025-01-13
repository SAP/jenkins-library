package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type onapsisExecuteScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newOnapsisExecuteScanTestsUtils() onapsisExecuteScanMockUtils {
	utils := onapsisExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunOnapsisExecuteScan(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := onapsisExecuteScanOptions{}

		utils := newOnapsisExecuteScanTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runOnapsisExecuteScan(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := onapsisExecuteScanOptions{}

		utils := newOnapsisExecuteScanTestsUtils()

		// test
		err := runOnapsisExecuteScan(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
