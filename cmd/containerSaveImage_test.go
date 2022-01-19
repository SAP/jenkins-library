package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

type containerMock struct {
	filePath         string
	imageSource      string
	registryURL      string
	localPath        string
	includeLayers    bool
	downloadImageErr string
	imageSourceErr   string
	tarImageErr      string
}

func (c *containerMock) DownloadImageToPath(imageSource, filePath string) (pkgutil.Image, error) {
	c.imageSource = imageSource
	c.filePath = filePath
	if c.downloadImageErr != "" {
		return pkgutil.Image{}, fmt.Errorf(c.downloadImageErr)
	}
	return pkgutil.Image{}, nil
}

func (c *containerMock) GetImageSource() (string, error) {
	if c.imageSourceErr != "" {
		return "", fmt.Errorf(c.imageSourceErr)
	}
	return "imageSource", nil
}

func (c *containerMock) TarImage(writer io.Writer, image pkgutil.Image) error {
	if c.tarImageErr != "" {
		return fmt.Errorf(c.tarImageErr)
	}
	writer.Write([]byte("This is a test"))
	return nil
}

func TestRunContainerSaveImage(t *testing.T) {
	telemetryData := telemetry.CustomData{}

	t.Run("success case", func(t *testing.T) {
		config := containerSaveImageOptions{}
		tmpFolder, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("failed to create temp dir")
		}
		defer os.RemoveAll(tmpFolder)

		cacheFolder := filepath.Join(tmpFolder, "cache")

		config.FilePath = "testfile"

		dClient := containerMock{}

		err = runContainerSaveImage(&config, &telemetryData, cacheFolder, tmpFolder, &dClient)
		assert.NoError(t, err)

		assert.Equal(t, cacheFolder, dClient.filePath)
		assert.Equal(t, "imageSource", dClient.imageSource)

		content, err := ioutil.ReadFile(filepath.Join(tmpFolder, "testfile.tar"))
		assert.NoError(t, err)
		assert.Equal(t, "This is a test", string(content))
	})

	t.Run("failure - cache creation", func(t *testing.T) {
		config := containerSaveImageOptions{}
		dClient := containerMock{}
		err := runContainerSaveImage(&config, &telemetryData, "", "", &dClient)
		assert.Contains(t, fmt.Sprint(err), "failed to create cache: mkdir :")
	})

	t.Run("failure - get image source", func(t *testing.T) {
		config := containerSaveImageOptions{}
		tmpFolder, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("failed to create temp dir")
		}
		defer os.RemoveAll(tmpFolder)

		dClient := containerMock{imageSourceErr: "image source error"}
		err = runContainerSaveImage(&config, &telemetryData, filepath.Join(tmpFolder, "cache"), tmpFolder, &dClient)
		assert.EqualError(t, err, "failed to get docker image source: image source error")
	})

	t.Run("failure - download image", func(t *testing.T) {
		config := containerSaveImageOptions{}
		tmpFolder, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("failed to create temp dir")
		}
		defer os.RemoveAll(tmpFolder)

		dClient := containerMock{downloadImageErr: "download error"}
		err = runContainerSaveImage(&config, &telemetryData, filepath.Join(tmpFolder, "cache"), tmpFolder, &dClient)
		assert.EqualError(t, err, "failed to download docker image: download error")
	})

	t.Run("failure - tar image", func(t *testing.T) {
		config := containerSaveImageOptions{}
		tmpFolder, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("failed to create temp dir")
		}
		defer os.RemoveAll(tmpFolder)

		dClient := containerMock{tarImageErr: "tar error"}
		err = runContainerSaveImage(&config, &telemetryData, filepath.Join(tmpFolder, "cache"), tmpFolder, &dClient)
		assert.EqualError(t, err, "failed to tar container image: tar error")
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
