package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type npmExecuteTestsMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newNpmExecuteTestsTestsUtils() npmExecuteTestsMockUtils {
	utils := npmExecuteTestsMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunNpmExecuteTests(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := npmExecuteTestsOptions{}

		utils := newNpmExecuteTestsTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runNpmExecuteTests(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := npmExecuteTestsOptions{}

		utils := newNpmExecuteTestsTestsUtils()

		// test
		err := runNpmExecuteTests(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
