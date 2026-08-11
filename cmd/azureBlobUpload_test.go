//go:build unit
// +build unit

package cmd

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/stretchr/testify/assert"
)

type mockAzureContainerAPI func(blobName string) *blockblob.Client

func (m mockAzureContainerAPI) NewBlockBlobClient(blobName string) *blockblob.Client {
	return m(blobName)
}

func mockAzureContainerClient(t *testing.T, fail bool) azureContainerAPI {
	return mockAzureContainerAPI(func(blobName string) *blockblob.Client {
		t.Helper()
		if fail {
			return nil
		}
		return &blockblob.Client{}
	})
}

func uploadFuncMock(ctx context.Context, api *blockblob.Client, file *os.File, o *blockblob.UploadFileOptions) (blockblob.UploadFileResponse, error) {
	return blockblob.UploadFileResponse{}, nil
}

func TestRunAzureBlobUpload(t *testing.T) {
	t.Parallel()

	t.Run("positive tests", func(t *testing.T) {
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
	})

	t.Run("negative tests", func(t *testing.T) {
		t.Parallel()

		t.Run("error path", func(t *testing.T) {
			t.Parallel()
			// initialization
			config := azureBlobUploadOptions{
				FilePath: "nonExistingFilepath",
			}
			container := mockAzureContainerClient(t, false)
			// test
			err := executeUpload(&config, container, uploadFuncMock)
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
			uploadFuncFail := func(ctx context.Context, api *blockblob.Client, file *os.File, o *blockblob.UploadFileOptions) (blockblob.UploadFileResponse, error) {
				return blockblob.UploadFileResponse{}, errors.New("error uploading")
			}
			container := mockAzureContainerClient(t, false)
			// test
			err = executeUpload(&config, container, uploadFuncFail)
			// assert
			assert.EqualError(t, err, "There was an error during the upload of file '"+f.Name()+"': error uploading")
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
			assert.EqualError(t, err, "Azure credentials are not valid: Key: 'azureCredentials.SASToken' Error:Field validation for 'SASToken' failed on the 'required' tag")
		})

		t.Run("error JSONStruct", func(t *testing.T) {
			t.Parallel()
			// initialization
			config := azureBlobUploadOptions{
				JSONCredentialsAzure: `faulty json`,
			}
			// test
			_, err := setup(&config)
			// assert
			assert.EqualError(t, err, "Could not read JSONCredentialsAzure: invalid character 'u' in literal false (expecting 'l')")
		})
	})
}
