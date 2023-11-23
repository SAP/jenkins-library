package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dockermock "github.com/SAP/jenkins-library/pkg/docker/mock"
	"github.com/SAP/jenkins-library/pkg/mock"
)

type imagePushToRegistryMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*dockermock.CraneMockUtils
	*dockerConfigUtilsBundle
}

func newImagePushToRegistryMockUtils(craneUtils *dockermock.CraneMockUtils) imagePushToRegistryUtils {
	utils := &imagePushToRegistryMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		CraneMockUtils: &dockermock.CraneMockUtils{},
	}

	return utils
}

func TestRunImagePushToRegistry(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := imagePushToRegistryOptions{
			SourceImage:            "source-image:latest",
			TargetImage:            "target-image",
			TargetRegistryURL:      "https://registry.test",
			TargetRegistryUser:     "user",
			TargetRegistryPassword: "password",
		}
		craneMockUtils := &dockermock.CraneMockUtils{}
		utils := newImagePushToRegistryMockUtils(craneMockUtils)
		// utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runImagePushToRegistry(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := imagePushToRegistryOptions{}
		craneMockUtils := &dockermock.CraneMockUtils{}
		utils := newImagePushToRegistryMockUtils(craneMockUtils)

		// test
		_ = runImagePushToRegistry(&config, nil, utils)

		// assert
		// assert.EqualError(t, err, "cannot run without important file")
	})
}
