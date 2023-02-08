package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type messageMappingDeployMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newMessageMappingDeployTestsUtils() messageMappingDeployMockUtils {
	utils := messageMappingDeployMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunMessageMappingDeploy(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := messageMappingDeployOptions{}

		utils := newMessageMappingDeployTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runMessageMappingDeploy(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := messageMappingDeployOptions{}

		utils := newMessageMappingDeployTestsUtils()

		// test
		err := runMessageMappingDeploy(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
