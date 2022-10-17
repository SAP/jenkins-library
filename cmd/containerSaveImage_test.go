package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/fake"
)

func TestRunContainerSaveImage(t *testing.T) {
	telemetryData := telemetry.CustomData{}

	t.Run("success case", func(t *testing.T) {
		config := containerSaveImageOptions{}
		config.FilePath = "testfile.tar"

		dClient := mock.DownloadMock{}
		files := mock.FilesMock{}

		cacheFolder, err := files.TempDir("", "containerSaveImage-")
		assert.NoError(t, err)

		dClient.Stub = func(imgRef string, dest string) (v1.Image, error) {
			files.AddFile(dest, []byte("This is a test"))
			return &fake.FakeImage{}, nil
		}

		filePath, err := runContainerSaveImage(&config, &telemetryData, cacheFolder, cacheFolder, &dClient, &files)
		assert.NoError(t, err)

		content, err := files.FileRead(filepath.Join(cacheFolder, "testfile.tar"))
		assert.NoError(t, err)
		assert.Equal(t, "This is a test", string(content))

		assert.Contains(t, filePath, "testfile.tar")
	})

	t.Run("failure - download image", func(t *testing.T) {
		config := containerSaveImageOptions{}
		tmpFolder := t.TempDir()

		dClient := mock.DownloadMock{ReturnError: "download error"}
		files := mock.FilesMock{}
		_, err := runContainerSaveImage(&config, &telemetryData, filepath.Join(tmpFolder, "cache"), tmpFolder, &dClient, &files)
		assert.EqualError(t, err, "failed to download docker image: download error")
	})
}

func TestFilenameFromContainer(t *testing.T) {

	tt := []struct {
		rootPath       string
		containerImage string
		expected       string
	}{
		{rootPath: "", containerImage: "image:tag", expected: "image_tag.tar"},
		{rootPath: "root/path", containerImage: "image:tag", expected: filepath.Join("root/path", "image_tag.tar")},
		{rootPath: "", containerImage: "my.registry.com:55555/path/to/my/image:tag", expected: "my_registry_com_55555_path_to_my_image_tag.tar"},
		{rootPath: "root/path", containerImage: "my.registry.com:55555/path/to/my/image:tag", expected: filepath.Join("root/path", "my_registry_com_55555_path_to_my_image_tag.tar")},
	}

	for _, test := range tt {
		assert.Equal(t, test.expected, filenameFromContainer(test.rootPath, test.containerImage))
	}

}

func TestCorrectContainerDockerConfigEnvVar(t *testing.T) {
	t.Run("with credentials", func(t *testing.T) {
		// init
		utilsMock := mock.FilesMock{}
		utilsMock.CurrentDir = "/tmp/test"

		dockerConfigFile := "myConfig/docker.json"
		utilsMock.AddFile(dockerConfigFile, []byte("{}"))

		resetValue := os.Getenv("DOCKER_CONFIG")
		os.Setenv("DOCKER_CONFIG", "")
		defer os.Setenv("DOCKER_CONFIG", resetValue)

		// test
		err := correctContainerDockerConfigEnvVar(&containerSaveImageOptions{DockerConfigJSON: dockerConfigFile}, &utilsMock)
		// assert
		assert.NoError(t, err)
		assert.NotNil(t, os.Getenv("DOCKER_CONFIG"))
	})
	t.Run("with added credentials", func(t *testing.T) {
		// init
		utilsMock := mock.FilesMock{}
		utilsMock.CurrentDir = "/tmp/test"

		dockerConfigFile := "myConfig/docker.json"
		utilsMock.AddFile(dockerConfigFile, []byte("{}"))

		resetValue := os.Getenv("DOCKER_CONFIG")
		os.Setenv("DOCKER_CONFIG", "")
		defer os.Setenv("DOCKER_CONFIG", resetValue)

		// test
		err := correctContainerDockerConfigEnvVar(&containerSaveImageOptions{DockerConfigJSON: dockerConfigFile, ContainerRegistryURL: "https://test.registry", ContainerRegistryUser: "testuser", ContainerRegistryPassword: "testPassword"}, &utilsMock)
		// assert
		assert.NoError(t, err)
		assert.NotNil(t, os.Getenv("DOCKER_CONFIG"))
		absoluteFilePath, _ := utilsMock.Abs(fmt.Sprintf("%s/%s", os.Getenv("DOCKER_CONFIG"), "config.json"))
		content, _ := utilsMock.FileRead(absoluteFilePath)
		assert.Contains(t, string(content), "https://test.registry")
	})
	t.Run("without credentials", func(t *testing.T) {
		// init
		utilsMock := mock.FilesMock{}
		resetValue := os.Getenv("DOCKER_CONFIG")
		os.Setenv("DOCKER_CONFIG", "")
		defer os.Setenv("DOCKER_CONFIG", resetValue)
		// test
		err := correctContainerDockerConfigEnvVar(&containerSaveImageOptions{}, &utilsMock)
		// assert
		assert.NoError(t, err)
		assert.NotNil(t, os.Getenv("DOCKER_CONFIG"))
	})
}
