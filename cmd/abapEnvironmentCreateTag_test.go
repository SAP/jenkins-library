package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
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
		// init
		config := abapEnvironmentCreateTagOptions{}

		utils := newAbapEnvironmentCreateTagTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runAbapEnvironmentCreateTag(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentCreateTagOptions{}

		utils := newAbapEnvironmentCreateTagTestsUtils()

		// test
		err := runAbapEnvironmentCreateTag(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
