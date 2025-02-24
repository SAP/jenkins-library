//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestGCSIntegration ./integration/...

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/api/option"
)

func TestGCSIntegrationClient(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	testdataPath, err := filepath.Abs("testdata/TestGCSIntegration")
	assert.NoError(t, err)

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			AlwaysPullImage: true,
			Image:           "fsouza/fake-gcs-server:1.30.2",
			ExposedPorts:    []string{"4443/tcp"},
			WaitingFor:      wait.ForListeningPort("4443/tcp"),
			Cmd:             []string{"-scheme", "https", "-public-host", "localhost"},
			Mounts: testcontainers.Mounts(
				testcontainers.BindMount(testdataPath, "/data"),
			),
		},
		Started: true,
	}

	gcsContainer, err := testcontainers.GenericContainer(ctx, req)
	require.NoError(t, err)
	defer gcsContainer.Terminate(ctx)

	ip, err := gcsContainer.Host(ctx)
	require.NoError(t, err)
	port, err := gcsContainer.MappedPort(ctx, "4443")
	endpoint := fmt.Sprintf("https://%s:%s/storage/v1/", ip, port.Port())
	httpclient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	t.Run("Test list files - success", func(t *testing.T) {
		bucketID := "sample-bucket"
		gcsClient, err := gcs.NewClient("dummy", "", gcs.WithGCSClientOptions(option.WithEndpoint(endpoint), option.WithoutAuthentication(), option.WithHTTPClient(&httpclient)))
		assert.NoError(t, err)
		fileNames, err := gcsClient.ListFiles(bucketID)
		assert.NoError(t, err)
		assert.Equal(t, []string{"dir/test_file2.yaml", "test_file.txt"}, fileNames)
		err = gcsClient.Close()
		assert.NoError(t, err)
	})

	t.Run("Test list files in missing bucket", func(t *testing.T) {
		bucketID := "missing-bucket"
		gcsClient, err := gcs.NewClient("dummy", "", gcs.WithGCSClientOptions(option.WithEndpoint(endpoint), option.WithoutAuthentication(), option.WithHTTPClient(&httpclient)))
		defer gcsClient.Close()
		assert.NoError(t, err)
		_, err = gcsClient.ListFiles(bucketID)
		assert.Error(t, err, "bucket doesn't exist")
		err = gcsClient.Close()
		assert.NoError(t, err)
	})

	t.Run("Test upload & download files - success", func(t *testing.T) {
		bucketID := "upload-bucket"
		file1Reader, file1Writer := io.Pipe()
		file2Reader, file2Writer := io.Pipe()
		gcsClient, err := gcs.NewClient("dummy", "", gcs.WithOpenFileFunction(openFileMock), gcs.WithCreateFileFunction(getCreateFileMock(file1Writer, file2Writer)),
			gcs.WithGCSClientOptions(option.WithEndpoint(endpoint), option.WithoutAuthentication(), option.WithHTTPClient(&httpclient)))
		assert.NoError(t, err)
		err = gcsClient.UploadFile(bucketID, "file1", "test/file1")
		assert.NoError(t, err)
		err = gcsClient.UploadFile(bucketID, "folder/file2", "test/folder/file2")
		assert.NoError(t, err)
		fileNames, err := gcsClient.ListFiles(bucketID)
		assert.NoError(t, err)
		assert.Equal(t, []string{"placeholder", "test/file1", "test/folder/file2"}, fileNames)
		go gcsClient.DownloadFile(bucketID, "test/file1", "file1")
		fileContent, err := io.ReadAll(file1Reader)
		assert.NoError(t, err)
		assert.Equal(t, file1Content, string(fileContent))
		go gcsClient.DownloadFile(bucketID, "test/folder/file2", "file2")
		fileContent, err = io.ReadAll(file2Reader)
		assert.NoError(t, err)
		assert.Equal(t, file2Content, string(fileContent))

		err = gcsClient.Close()
		assert.NoError(t, err)
	})

	t.Run("Test upload missing file", func(t *testing.T) {
		bucketID := "upload-bucket"
		gcsClient, err := gcs.NewClient("dummy", "", gcs.WithOpenFileFunction(openFileMock),
			gcs.WithGCSClientOptions(option.WithEndpoint(endpoint), option.WithoutAuthentication(), option.WithHTTPClient(&httpclient)))
		assert.NoError(t, err)
		err = gcsClient.UploadFile(bucketID, "file3", "test/file3")
		assert.Contains(t, err.Error(), "could not open source file")
		err = gcsClient.Close()
		assert.NoError(t, err)
	})

	t.Run("Test download missing file", func(t *testing.T) {
		bucketID := "upload-bucket"
		gcsClient, err := gcs.NewClient("dummy", "", gcs.WithOpenFileFunction(openFileMock),
			gcs.WithGCSClientOptions(option.WithEndpoint(endpoint), option.WithoutAuthentication(), option.WithHTTPClient(&httpclient)))
		assert.NoError(t, err)
		err = gcsClient.DownloadFile(bucketID, "test/file3", "file3")
		assert.Contains(t, err.Error(), "could not open source file")
		err = gcsClient.Close()
		assert.NoError(t, err)
	})

	t.Run("Test download file - failed file creation", func(t *testing.T) {
		bucketID := "upload-bucket"
		_, file1Writer := io.Pipe()
		_, file2Writer := io.Pipe()
		gcsClient, err := gcs.NewClient("dummy", "", gcs.WithOpenFileFunction(openFileMock), gcs.WithCreateFileFunction(getCreateFileMock(file1Writer, file2Writer)),
			gcs.WithGCSClientOptions(option.WithEndpoint(endpoint), option.WithoutAuthentication(), option.WithHTTPClient(&httpclient)))
		assert.NoError(t, err)
		err = gcsClient.DownloadFile(bucketID, "placeholder", "file3")
		assert.Contains(t, err.Error(), "could not create target file")
		err = gcsClient.Close()
		assert.NoError(t, err)
	})
}

const (
	file1Content = `test file`
	file2Content = `
		foo : bar
		pleh : help
		stuff : {'foo': 'bar', 'bar': 'foo'}
	`
)

func openFileMock(name string) (io.ReadCloser, error) {
	var fileContent string
	switch name {
	case "file1":
		fileContent = file1Content
	case "folder/file2":
		fileContent = file2Content
	default:
		return nil, errors.New("open file faled")
	}
	return io.NopCloser(strings.NewReader(fileContent)), nil
}

func getCreateFileMock(file1Writer io.WriteCloser, file2Writer io.WriteCloser) func(name string) (io.WriteCloser, error) {
	return func(name string) (io.WriteCloser, error) {
		switch name {
		case "file1":
			return file1Writer, nil
		case "file2":
			return file2Writer, nil
		default:
			return nil, errors.New("could not create target file")
		}
	}
}
