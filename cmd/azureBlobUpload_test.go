package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stretchr/testify/assert"
)

type mockAzureContainerAPI func(blobName string) (*azblob.BlockBlobClient, error)

func (m mockAzureContainerAPI) NewBlockBlobClient(blobName string) (*azblob.BlockBlobClient, error) {
	return m(blobName)
}

func mockAzureContainerClient(t *testing.T, fail bool) AzureContainerAPI {
	return mockAzureContainerAPI(func(blobName string) (*azblob.BlockBlobClient, error) {
		t.Helper()
		if fail {
			return nil, fmt.Errorf("error containerClient")
		}
		return &azblob.BlockBlobClient{}, nil
	})
}

func uploadFuncMock(ctx context.Context, api *azblob.BlockBlobClient, file *os.File, o azblob.UploadOption) (*http.Response, error) {
	return &http.Response{}, nil
}

func TestRunAzureBlobUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// create temporary file
		f, err := os.CreateTemp("", "tmpfile-")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		defer os.Remove(f.Name())
		data := []byte("test test test")
		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}
		// initialization
		config := azureBlobUploadOptions{
			FilePath: f.Name(),
		}
		container := mockAzureContainerClient(t, false)
		// test
		err = executeUpload(&config, container, uploadFuncMock)
		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// initialization
		config := azureBlobUploadOptions{
			FilePath: "nonExistingFilepath",
		}
		container := mockAzureContainerClient
		// test
		err := executeUpload(&config, container(t, false), uploadFuncMock)
		// assert
		assert.IsType(t, &fs.PathError{}, errors.Unwrap(err))
	})

	t.Run("error containerClient", func(t *testing.T) {
		t.Parallel()
		// create temporary file
		f, err := os.CreateTemp("", "tmpfile-")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		defer os.Remove(f.Name())
		data := []byte("test test test")
		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}
		// initialization
		config := azureBlobUploadOptions{
			FilePath: f.Name(),
		}
		container := mockAzureContainerClient
		// test
		err = executeUpload(&config, container(t, true), uploadFuncMock)
		// assert
		assert.EqualError(t, err, "Could not instantiate Azure blockBlobClient from Azure Container Client: error containerClient")
	})

	t.Run("error credentials", func(t *testing.T) {
		t.Parallel()
		// initialization
		config := azureBlobUploadOptions{
			JSONCredentialsAzure: `{
				"account_name": "name",
				"container_name": "container"
			  }`,
			FilePath: "nonExistingFilepath",
		}
		// test
		_, err := setup(&config)
		// assert
		assert.EqualError(t, err, "Azure credentials are not valid: Key: 'AzureCredentials.SASToken' Error:Field validation for 'SASToken' failed on the 'required' tag")
	})
}
