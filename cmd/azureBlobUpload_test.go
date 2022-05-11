package cmd

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stretchr/testify/assert"
)

type mockAzureContainerAPI func(blobName string) azblob.BlockBlobClient

func (m mockAzureContainerAPI) NewBlockBlobClient(blobName string) azblob.BlockBlobClient {
	return m(blobName)
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
		container := mockAzureContainerClient
		// test
		err = runAzureBlobUpload(&config, container(t), UploadMock)
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
		err := runAzureBlobUpload(&config, container(t), UploadMock)
		// assert
		_, ok := err.(*fs.PathError)
		assert.True(t, ok)
	})
}

func mockAzureContainerClient(t *testing.T) AzureContainerAPI {
	return mockAzureContainerAPI(func(blobName string) azblob.BlockBlobClient {
		t.Helper()
		return azblob.BlockBlobClient{}
	})
}

func UploadMock(ctx context.Context, api *azblob.BlockBlobClient, file *os.File, o azblob.HighLevelUploadToBlockBlobOption) (*http.Response, error) {
	return &http.Response{}, nil
}
