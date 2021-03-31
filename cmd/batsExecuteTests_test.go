package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type batsExecuteTestsMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func (b batsExecuteTestsMockUtils) CloneRepo(URL string) error {
	return nil
}

func newBatsExecuteTestsTestsUtils() batsExecuteTestsMockUtils {
	utils := batsExecuteTestsMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunBatsExecuteTests(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		config := &batsExecuteTestsOptions{
			OutputFormat: "junit",
			Repository:   "https://github.com/bats-core/bats-core.git",
			TestPackage:  "piper-bats",
			TestPath:     "src/test",
		}

		mockUtils := newBatsExecuteTestsTestsUtils()
		err := runBatsExecuteTests(config, nil, &mockUtils)
		assert.NoError(t, err)
		assert.True(t, mockUtils.HasFile("TEST-"+config.TestPackage+".tap"))
		assert.True(t, mockUtils.HasFile("TEST-"+config.TestPackage+".xml"))
	})

	t.Run("output tap case", func(t *testing.T) {
		config := &batsExecuteTestsOptions{
			OutputFormat: "tap",
			Repository:   "https://github.com/bats-core/bats-core.git",
			TestPackage:  "piper-bats",
			TestPath:     "src/test",
		}

		mockUtils := newBatsExecuteTestsTestsUtils()
		err := runBatsExecuteTests(config, nil, &mockUtils)
		assert.NoError(t, err)
		assert.True(t, mockUtils.HasFile("TEST-"+config.TestPackage+".tap"))
		assert.False(t, mockUtils.HasFile("TEST-"+config.TestPackage+".xml"))
	})

	t.Run("output format failed case", func(t *testing.T) {
		config := &batsExecuteTestsOptions{
			OutputFormat: "fail",
			Repository:   "https://github.com/bats-core/bats-core.git",
			TestPackage:  "piper-bats",
			TestPath:     "src/test",
		}

		mockUtils := newBatsExecuteTestsTestsUtils()
		err := runBatsExecuteTests(config, nil, &mockUtils)
		assert.EqualError(t, err, "output format 'fail' is incorrect. Possible drivers: tap, junit")
	})
}
