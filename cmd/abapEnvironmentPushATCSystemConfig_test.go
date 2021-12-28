package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type abapEnvironmentPushATCSystemConfigMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newAbapEnvironmentPushATCSystemConfigTestsUtils() abapEnvironmentPushATCSystemConfigMockUtils {
	utils := abapEnvironmentPushATCSystemConfigMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunAbapEnvironmentPushATCSystemConfig(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentPushATCSystemConfigOptions{}

		utils := newAbapEnvironmentPushATCSystemConfigTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runAbapEnvironmentPushATCSystemConfig(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentPushATCSystemConfigOptions{}

		utils := newAbapEnvironmentPushATCSystemConfigTestsUtils()

		// test
		err := runAbapEnvironmentPushATCSystemConfig(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
