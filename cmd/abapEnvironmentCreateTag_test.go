package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type abapEnvironmentCreateTagMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newAbapEnvironmentCreateTagTestsUtils() abapEnvironmentCreateTagMockUtils {
	utils := abapEnvironmentCreateTagMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunAbapEnvironmentCreateTag(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

	})

}
