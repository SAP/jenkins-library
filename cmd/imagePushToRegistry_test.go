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

func newImagePushToRegistryMockUtils(dockerUtilsBundleMock *dockerUtilsBundleMock) imagePushToRegistryUtils {
	utils := &imagePushToRegistryMockUtils{
		ExecMockRunner:        &mock.ExecMockRunner{},
		FilesMock:             &mock.FilesMock{},
		dockerUtilsBundleMock: dockerUtilsBundleMock,
	}
	return utils
}

type dockerUtilsBundleMock struct {
	errCreateConfig, errMergeConfig, errLoadImage, errPushImage, errCopyImage error
}

func (d *dockerUtilsBundleMock) CreateDockerConfigJSON(registry, username, password, targetPath, configPath string, utils piperutils.FileUtils) (string, error) {
	return "", d.errCreateConfig
}

func (d *dockerUtilsBundleMock) MergeDockerConfigJSON(sourcePath, targetPath string, utils piperutils.FileUtils) error {
	return d.errMergeConfig
}

func (d *dockerUtilsBundleMock) LoadImage(src string) (v1.Image, error) {
	return nil, d.errLoadImage
}

func (d *dockerUtilsBundleMock) PushImage(im v1.Image, dest string) error {
	return d.errPushImage
}

func (d *dockerUtilsBundleMock) CopyImage(src, dest string) error {
	return d.errCopyImage
}

func TestRunImagePushToRegistry(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := imagePushToRegistryOptions{}
		dockerMockUtils := &dockerUtilsBundleMock{}
		utils := newImagePushToRegistryMockUtils(dockerMockUtils)
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
		dockerMockUtils := &dockerUtilsBundleMock{}
		utils := newImagePushToRegistryMockUtils(dockerMockUtils)

		// test
		err := runImagePushToRegistry(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
