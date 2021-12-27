package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type helmExecuteMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newHelmExecuteTestsUtils() helmExecuteMockUtils {
	utils := helmExecuteMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunHelmExecute(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := helmExecuteOptions{}

		utils := newHelmExecuteTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runHelmExecute(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := helmExecuteOptions{}

		utils := newHelmExecuteTestsUtils()

		// test
		err := runHelmExecute(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
