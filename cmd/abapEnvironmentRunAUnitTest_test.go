package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type abapEnvironmentRunAUnitTestMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newAbapEnvironmentRunAUnitTestTestsUtils() abapEnvironmentRunAUnitTestMockUtils {
	utils := abapEnvironmentRunAUnitTestMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunAbapEnvironmentRunAUnitTest(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentRunAUnitTestOptions{}

		utils := newAbapEnvironmentRunAUnitTestTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runAbapEnvironmentRunAUnitTest(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentRunAUnitTestOptions{}

		utils := newAbapEnvironmentRunAUnitTestTestsUtils()

		// test
		err := runAbapEnvironmentRunAUnitTest(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
