// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/api/option"
)

func Test_gcsClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testdataPath, err := filepath.Abs("testdata/TestGCSIntegration")
	assert.NoError(t, err)

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			AlwaysPullImage: true,
			Image:           "fsouza/fake-gcs-server:1.29.2",
			ExposedPorts:    []string{"4443/tcp"},
			WaitingFor:      wait.ForListeningPort("4443/tcp"),
			Cmd:             []string{"-scheme", "http"},
			BindMounts: map[string]string{
				testdataPath: "/data",
			},
		},
		Started: true,
	}

	gcsContainer, err := testcontainers.GenericContainer(ctx, req)
	assert.NoError(t, err)
	defer gcsContainer.Terminate(ctx)

	ip, err := gcsContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := gcsContainer.MappedPort(ctx, "4443")
	endpoint := fmt.Sprintf("http://%s:%s/storage/v1/", ip, port.Port())

	t.Run("Test list files - success", func(t *testing.T) {
		bucketID := "sample-bucket"
		gcsClient, err := gcs.NewClient([]gcs.EnvVar{}, nil, nil, bucketID, "", option.WithEndpoint(endpoint), option.WithoutAuthentication())
		assert.NoError(t, err)
		fileNames, err := gcsClient.ListFiles()
		assert.NoError(t, err)
		assert.Equal(t, []string{"dir/test_file2.yaml", "test_file.txt"}, fileNames)
		err = gcsClient.Close()
		assert.NoError(t, err)
	})

	t.Run("Test list files in missing bucket", func(t *testing.T) {
		bucketID := "missing-bucket"
		gcsClient, err := gcs.NewClient([]gcs.EnvVar{}, nil, nil, bucketID, "", option.WithEndpoint(endpoint), option.WithoutAuthentication())
		defer gcsClient.Close()
		assert.NoError(t, err)
		_, err = gcsClient.ListFiles()
		assert.Error(t, err, "bucket doesn't exist")
		err = gcsClient.Close()
		assert.NoError(t, err)
	})

	t.Run("Test upload files - success", func(t *testing.T) {
		bucketID := "upload-bucket"
		targetFolder := "test/"
		gcsClient, err := gcs.NewClient([]gcs.EnvVar{}, openFileMock, nil, bucketID, targetFolder, option.WithEndpoint(endpoint), option.WithoutAuthentication())
		assert.NoError(t, err)
		err = gcsClient.UploadFile("file1")
		assert.NoError(t, err)
		err = gcsClient.UploadFile("folder/file2")
		assert.NoError(t, err)
		fileNames, err := gcsClient.ListFiles()
		assert.NoError(t, err)
		assert.Equal(t, []string{"placeholder", "test/file1", "test/folder/file2"}, fileNames)

		err = gcsClient.Close()
		assert.NoError(t, err)
	})

	t.Run("Test upload missing file", func(t *testing.T) {
		bucketID := "upload-bucket"
		targetFolder := "test/"
		gcsClient, err := gcs.NewClient([]gcs.EnvVar{}, openFileMock, nil, bucketID, targetFolder, option.WithEndpoint(endpoint), option.WithoutAuthentication())
		assert.NoError(t, err)
		err = gcsClient.UploadFile("file3")
		assert.Error(t, err, "could not open source file")
		err = gcsClient.Close()
		assert.NoError(t, err)
	})

	// TODO: Test DownloadFile
}

func openFileMock(name string) (io.ReadCloser, error) {
	var fileContent string
	switch name {
	case "file1":
		fileContent = `test file`
	case "folder/file2":
		fileContent = `
		foo : bar
		pleh : help
		stuff : {'foo': 'bar', 'bar': 'foo'}
		`
	default:
		return nil, errors.New("open file faled")
	}
	return ioutil.NopCloser(strings.NewReader(fileContent)), nil
}
