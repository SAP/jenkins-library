package cmd

import (
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type imagePushToRegistryMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*dockerUtilsBundleMock
}

func newImagePushToRegistryTestsUtils() imagePushToRegistryUtils {
	utils := &imagePushToRegistryMockUtils{
		ExecMockRunner:        &mock.ExecMockRunner{},
		FilesMock:             &mock.FilesMock{},
		dockerUtilsBundleMock: &dockerUtilsBundleMock{},
	}
	return utils
}

type dockerUtilsBundleMock struct{}

func (d *dockerUtilsBundleMock) CreateDockerConfigJSON(registry, username, password, targetPath, configPath string, utils piperutils.FileUtils) (string, error) {
	return "", nil
}

func (d *dockerUtilsBundleMock) MergeDockerConfigJSON(sourcePath, targetPath string, utils piperutils.FileUtils) error {
	return nil
}

func (d *dockerUtilsBundleMock) LoadImage(src string) (v1.Image, error) {
	return nil, nil
}

func (d *dockerUtilsBundleMock) PushImage(im v1.Image, dest string) error {
	return nil
}

func (d *dockerUtilsBundleMock) CopyImage(src, dest string) error {
	return nil
}

func TestRunImagePushToRegistry(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := imagePushToRegistryOptions{}

		utils := newImagePushToRegistryTestsUtils()
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

		utils := newImagePushToRegistryTestsUtils()

		// test
		err := runImagePushToRegistry(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
