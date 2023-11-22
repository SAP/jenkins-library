package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type imagePushToRegistryMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newImagePushToRegistryTestsUtils() imagePushToRegistryMockUtils {
	utils := imagePushToRegistryMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunImagePushToRegistry(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := imagePushToRegistryOptions{}

		utils := newImagePushToRegistryTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runImagePushToRegistry(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := imagePushToRegistryOptions{}

		utils := newImagePushToRegistryTestsUtils()

		// test
		err := runImagePushToRegistry(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
