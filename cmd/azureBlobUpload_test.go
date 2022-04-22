package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type azureBlobUploadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newAzureBlobUploadTestsUtils() azureBlobUploadMockUtils {
	utils := azureBlobUploadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

type mockAzureContainerAPI func(blobName string) (*azblob.BlockBlobClient, error)

func (m mockAzureContainerAPI) NewBlockBlobClient(blobName string) (*azblob.BlockBlobClient, error) {
	return m(blobName)
}

func TestRunAzureBlobUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// initialization

		config := azureBlobUploadOptions{
			FilePath: filepath.Join("testdata", t.Name()+"_test.txt"),
		}
		utils := newAzureBlobUploadTestsUtils()
		container := mockAzureContainerClient

		// test
		err := runAzureBlobUpload(&config, nil, utils, container(t), UploadMock)
		// assert
		assert.NoError(t, err)
	})

	t.Run("no path", func(t *testing.T) {
		t.Parallel()
		//init
		config := azureBlobUploadOptions{}
		utils := newAzureBlobUploadTestsUtils()
		container := mockAzureContainerClient

		err := runAzureBlobUpload(&config, nil, utils, container(t), UploadMock)
		assert.EqualError(t, err, "File Path Parameter is empty. Please specify a file or directory to Upload to Azure!")
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// initialization
		config := azureBlobUploadOptions{
			FilePath: "nonExistingFilepath",
		}
		container := mockAzureContainerClient
		utils := newAzureBlobUploadTestsUtils()
		// test
		err := runAzureBlobUpload(&config, nil, utils, container(t), UploadMock)
		// assert
		_, ok := err.(*fs.PathError)
		assert.True(t, ok)
	})

	t.Run("error blobName", func(t *testing.T) {
		t.Parallel()
		// initialization
		config := azureBlobUploadOptions{
			FilePath: filepath.Join("testdata", t.Name()+"_test.txt"),
		}
		container := mockAzureContainerClient
		utils := newAzureBlobUploadTestsUtils()
		// test
		err := runAzureBlobUpload(&config, nil, utils, container(t), UploadMock)
		// assert
		assert.EqualError(t, err, "invalid blobName")
	})
}

func mockAzureContainerClient(t *testing.T) AzureContainerAPI {
	return mockAzureContainerAPI(func(blobName string) (*azblob.BlockBlobClient, error) {
		t.Helper()
		if blobName == "" {
			return nil, fmt.Errorf("expect blobName not to be empty")
		}
		if blobName == "testdata/TestRunAzureBlobUpload/error_blobName_test.txt" {
			return nil, fmt.Errorf("invalid blobName")
		}
		return &azblob.BlockBlobClient{}, nil
	})
}

func UploadMock(ctx context.Context, api *azblob.BlockBlobClient, file *os.File, o azblob.UploadOption) (*http.Response, error) {
	return &http.Response{}, nil
}
