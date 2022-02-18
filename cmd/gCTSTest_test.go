package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type gCTSTestMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newGCTSTestTestsUtils() gCTSTestMockUtils {
	utils := gCTSTestMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunGCTSTest(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init

		utils := newGCTSTestTestsUtils()
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
