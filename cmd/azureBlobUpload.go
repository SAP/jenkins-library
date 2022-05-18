package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

// AzureContainerAPI is used to mock Azure containerClients in unit tests
type AzureContainerAPI interface {
	NewBlockBlobClient(blobName string) azblob.BlockBlobClient
}

// NewBlockBlobClient creates a blockBlobClient from a containerClient
func NewBlockBlobClient(blobName string, api AzureContainerAPI) azblob.BlockBlobClient {
	return api.NewBlockBlobClient(blobName)
}

// UploadFile uploads a file to an Azure Blob Storage
// The function is uses the UploadFileToBlockBlob function from the Azure SDK
// We introduce this 'wrapper' for mocking reasons
func UploadFile(ctx context.Context, api *azblob.BlockBlobClient, file *os.File, o azblob.HighLevelUploadToBlockBlobOption) (*http.Response, error) {
	return api.UploadFileToBlockBlob(ctx, file, o)
}

// Struct to store Azure credentials from specified JSON string
type azureCredentials struct {
	SASToken    string `json:"sas_token"`
	AccountName string `json:"account_name"`
	Container   string `json:"container_name"`
}

func azureBlobUpload(config azureBlobUploadOptions, telemetryData *telemetry.CustomData) {
	// Prepare Credentials
	log.Entry().Infoln("Start reading Azure Credentials")
	var obj azureCredentials

	err := json.Unmarshal([]byte(config.JSONCredentialsAzure), &obj)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Could not read JSONCredentialsAzure")
	}

	// Initialize Azure Service Client
	sasURL := fmt.Sprintf("https://%s.blob.core.windows.net/?%s", obj.AccountName, obj.SASToken)
	serviceClient, err := azblob.NewServiceClientWithNoCredential(sasURL, nil)
	if err != nil {
		log.Entry().WithError(err).Fatal("Could not instantiate Azure Service Client!")
	}

	// Get a containerClient from ServiceClient
	containerClient := serviceClient.NewContainerClient(obj.Container)

	err = runAzureBlobUpload(&config, containerClient, UploadFile)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAzureBlobUpload(config *azureBlobUploadOptions, containerClient AzureContainerAPI, Upload func(ctx context.Context, api *azblob.BlockBlobClient, file *os.File, o azblob.HighLevelUploadToBlockBlobOption) (*http.Response, error)) error {

	log.Entry().Infof("Starting walk through file path '%v'", config.FilePath)

	// All Blob Operations operate with context.Context, in our case the clients do not expire
	ctx := context.Background()

	//iterate through directories
	err := filepath.Walk(config.FilePath, func(currentFilePath string, f os.FileInfo, err error) error {
		// Handle Failure to prevent panic (e.g. in case of an invalid filepath)
		if err != nil {
			log.Entry().WithError(err).Warnf("Failed to access path: '%v'", currentFilePath)
			return err
		}
		// Skip directories, only upload files
		if !f.IsDir() {
			log.Entry().Infof("Current target path is: '%v'", currentFilePath)

			// Read Data from File
			data, e := os.Open(currentFilePath)
			if e != nil {
				log.Entry().WithError(e).Warnf("Could not read the file '%s'", currentFilePath)
				return e
			}
			defer data.Close()

			// Create a filepath in UNIX format so that the BlockBlobClient automatically detects directories
			key := filepath.ToSlash(currentFilePath)

			// Get a blockBlobClient from containerClient
			blockBlobClient := NewBlockBlobClient(key, containerClient)

			// Upload File
			log.Entry().Infof("Start upload of file '%v'", currentFilePath)
			var blockOptions azblob.HighLevelUploadToBlockBlobOption
			_, e = Upload(ctx, &blockBlobClient, data, blockOptions)
			if e != nil {
				log.Entry().WithError(e).Warnf("There was an error during the upload of file '%v'", currentFilePath)
				return e
			}

			log.Entry().Infof("Upload of file '%v' was successful!", currentFilePath)
			return e
		}
		return nil
	})

	if err != nil {
		return err
	}
	log.Entry().Infoln("Upload has successfully finished!")
	return err
}
