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

}
