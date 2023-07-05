//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type batsExecuteTestsMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func (b batsExecuteTestsMockUtils) CloneRepo(URL string) error {
	if URL != "https://github.com/bats-core/bats-core.git" {
		return git.ErrRepositoryNotExists
	}
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
	t.Parallel()
	t.Run("success case", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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

	t.Run("failed to clone repo case", func(t *testing.T) {
		t.Parallel()
		config := &batsExecuteTestsOptions{
			OutputFormat: "junit",
			Repository:   "fail",
			TestPackage:  "piper-bats",
			TestPath:     "src/test",
		}

		mockUtils := newBatsExecuteTestsTestsUtils()
		err := runBatsExecuteTests(config, nil, &mockUtils)

		expectedError := fmt.Errorf("couldn't pull %s repository: %w", config.Repository, git.ErrRepositoryNotExists)
		assert.EqualError(t, err, expectedError.Error())
	})

	t.Run("failed to run bats case", func(t *testing.T) {
		t.Parallel()
		config := &batsExecuteTestsOptions{
			OutputFormat: "tap",
			Repository:   "https://github.com/bats-core/bats-core.git",
			TestPackage:  "piper-bats",
			TestPath:     "src/test",
		}

		mockUtils := batsExecuteTestsMockUtils{
			ExecMockRunner: &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"bats-core/bin/bats": errors.New("error case")}},
			FilesMock:      &mock.FilesMock{},
		}
		err := runBatsExecuteTests(config, nil, &mockUtils)
		assert.Contains(t, fmt.Sprint(err), "failed to run bats test")

		assert.False(t, mockUtils.HasFile("TEST-"+config.TestPackage+".tap"))
		assert.False(t, mockUtils.HasFile("TEST-"+config.TestPackage+".xml"))
	})
}
