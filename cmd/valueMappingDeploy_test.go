package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type valueMappingDeployMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newValueMappingDeployTestsUtils() valueMappingDeployMockUtils {
	utils := valueMappingDeployMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunValueMappingDeploy(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := valueMappingDeployOptions{}

		utils := newValueMappingDeployTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runValueMappingDeploy(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := valueMappingDeployOptions{}

		utils := newValueMappingDeployTestsUtils()

		// test
		err := runValueMappingDeploy(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
