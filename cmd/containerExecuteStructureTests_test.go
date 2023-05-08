//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type containerStructureTestsMockUtils struct {
	shouldFail bool
	*mock.FilesMock
	*mock.ExecMockRunner
}

func (m *containerStructureTestsMockUtils) Glob(pattern string) (matches []string, err error) {
	switch pattern {
	case "**.yaml":
		return []string{"config1.yaml", "config2.yaml"}, nil
	case "empty":
		return []string{}, nil
	case "error":
		return nil, errors.New("failed to find fies")
	}

	return nil, nil
}

func newContainerStructureTestsMockUtils() containerStructureTestsMockUtils {
	utils := containerStructureTestsMockUtils{
		shouldFail:     false,
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

func TestRunContainerExecuteStructureTests(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		config := &containerExecuteStructureTestsOptions{
			PullImage:          true,
			TestConfiguration:  "**.yaml",
			TestDriver:         "docker",
			TestImage:          "reg/image:tag",
			TestReportFilePath: "report.json",
		}

		mockUtils := newContainerStructureTestsMockUtils()

		// test
		err := runContainerExecuteStructureTests(config, &mockUtils)
		// assert
		expectedParams := []string{
			"test",
			"--config", "config1.yaml",
			"--config", "config2.yaml",
			"--driver", "docker",
			"--pull",
			"--image", "reg/image:tag",
			"--test-report", "report.json",
		}

		assert.NoError(t, err)
		if assert.Equal(t, 1, len(mockUtils.Calls)) {
			assert.Equal(t, "./container-structure-test", mockUtils.Calls[0].Exec)
			assert.Equal(t, expectedParams, mockUtils.Calls[0].Params)
		}
	})

	t.Run("success case - without pulling image", func(t *testing.T) {
		config := &containerExecuteStructureTestsOptions{
			TestConfiguration:  "**.yaml",
			TestDriver:         "docker",
			TestImage:          "reg/image:tag",
			TestReportFilePath: "report.json",
		}

		mockUtils := newContainerStructureTestsMockUtils()

		// test
		err := runContainerExecuteStructureTests(config, &mockUtils)
		// assert
		expectedParams := []string{
			"test",
			"--config", "config1.yaml",
			"--config", "config2.yaml",
			"--driver", "docker",
			"--image", "reg/image:tag",
			"--test-report", "report.json",
		}

		assert.NoError(t, err)
		if assert.Equal(t, 1, len(mockUtils.Calls)) {
			assert.Equal(t, "./container-structure-test", mockUtils.Calls[0].Exec)
			assert.Equal(t, expectedParams, mockUtils.Calls[0].Params)
		}
	})

	t.Run("success case - verbose", func(t *testing.T) {
		GeneralConfig.Verbose = true
		config := &containerExecuteStructureTestsOptions{
			TestConfiguration:  "**.yaml",
			TestDriver:         "docker",
			TestImage:          "reg/image:tag",
			TestReportFilePath: "report.json",
		}

		mockUtils := newContainerStructureTestsMockUtils()

		// test
		err := runContainerExecuteStructureTests(config, &mockUtils)
		// assert
		expectedParams := []string{
			"test",
			"--config", "config1.yaml",
			"--config", "config2.yaml",
			"--driver", "docker",
			"--image", "reg/image:tag",
			"--test-report", "report.json",
			"--verbosity", "debug",
		}

		assert.NoError(t, err)
		if assert.Equal(t, 1, len(mockUtils.Calls)) {
			assert.Equal(t, "./container-structure-test", mockUtils.Calls[0].Exec)
			assert.Equal(t, expectedParams, mockUtils.Calls[0].Params)
		}
		GeneralConfig.Verbose = false
	})

	t.Run("success case - run on k8s", func(t *testing.T) {
		if err := os.Setenv("ON_K8S", "true"); err != nil {
			t.Error(err)
		}
		config := &containerExecuteStructureTestsOptions{
			TestConfiguration:  "**.yaml",
			TestImage:          "reg/image:tag",
			TestReportFilePath: "report.json",
		}

		mockUtils := newContainerStructureTestsMockUtils()

		// test
		err := runContainerExecuteStructureTests(config, &mockUtils)
		// assert
		expectedParams := []string{
			"test",
			"--config", "config1.yaml",
			"--config", "config2.yaml",
			"--driver", "tar",
			"--image", "reg/image:tag",
			"--test-report", "report.json",
		}

		assert.NoError(t, err)
		if assert.Equal(t, 1, len(mockUtils.Calls)) {
			assert.Equal(t, "./container-structure-test", mockUtils.Calls[0].Exec)
			assert.Equal(t, expectedParams, mockUtils.Calls[0].Params)
		}
		os.Unsetenv("ON_K8S")
	})

	t.Run("error case - execution failed", func(t *testing.T) {
		config := &containerExecuteStructureTestsOptions{
			PullImage:          true,
			TestConfiguration:  "**.yaml",
			TestDriver:         "docker",
			TestImage:          "reg/image:tag",
			TestReportFilePath: "report.json",
		}
		mockUtils := newContainerStructureTestsMockUtils()
		mockUtils.ExecMockRunner = &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{"container-structure-test": fmt.Errorf("container-structure-test run failed")},
		}

		// test
		err := runContainerExecuteStructureTests(config, &mockUtils)
		// assert
		assert.EqualError(t, err, "failed to run executable, command: '[./container-structure-test test --config config1.yaml --config config2.yaml --driver docker --pull --image reg/image:tag --test-report report.json]', error: container-structure-test run failed: container-structure-test run failed")
	})

	t.Run("error case - configuration is missing", func(t *testing.T) {
		config := &containerExecuteStructureTestsOptions{
			PullImage:          true,
			TestConfiguration:  "empty",
			TestDriver:         "docker",
			TestReportFilePath: "report.json",
		}
		mockUtils := newContainerStructureTestsMockUtils()

		// test
		err := runContainerExecuteStructureTests(config, &mockUtils)
		// assert
		assert.EqualError(t, err, "config files mustn't be missing")
	})

	t.Run("error case - failed to find config files", func(t *testing.T) {
		config := &containerExecuteStructureTestsOptions{
			PullImage:          true,
			TestConfiguration:  "error",
			TestDriver:         "docker",
			TestReportFilePath: "report.json",
		}
		mockUtils := newContainerStructureTestsMockUtils()

		// test
		err := runContainerExecuteStructureTests(config, &mockUtils)
		// assert
		assert.EqualError(t, err, "failed to find config files, error: failed to find fies: failed to find fies")
	})

	t.Run("error case - incorrect driver type", func(t *testing.T) {
		config := &containerExecuteStructureTestsOptions{
			PullImage:          true,
			TestConfiguration:  "**.yaml",
			TestDriver:         "wrongDriver",
			TestReportFilePath: "report.json",
		}
		mockUtils := newContainerStructureTestsMockUtils()

		// test
		err := runContainerExecuteStructureTests(config, &mockUtils)
		// assert
		assert.EqualError(t, err, "test driver wrongDriver is incorrect. Possible drivers: docker, tar")
	})
}
