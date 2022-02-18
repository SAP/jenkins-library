package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type gctsExecuteABAPUnitTestsMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newGctsExecuteABAPUnitTestsTestsUtils() gctsExecuteABAPUnitTestsMockUtils {
	utils := gctsExecuteABAPUnitTestsMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunGctsExecuteABAPUnitTests(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newGctsExecuteABAPUnitTestsTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test

		// assert

	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init

		// test

		// assert

	})
}
