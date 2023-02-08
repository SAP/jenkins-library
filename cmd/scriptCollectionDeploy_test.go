package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type scriptCollectionDeployMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newScriptCollectionDeployTestsUtils() scriptCollectionDeployMockUtils {
	utils := scriptCollectionDeployMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunScriptCollectionDeploy(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := scriptCollectionDeployOptions{}

		utils := newScriptCollectionDeployTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runScriptCollectionDeploy(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := scriptCollectionDeployOptions{}

		utils := newScriptCollectionDeployTestsUtils()

		// test
		err := runScriptCollectionDeploy(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
