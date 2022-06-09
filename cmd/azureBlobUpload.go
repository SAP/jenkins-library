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
	"github.com/go-playground/validator/v10"
)

// AzureContainerAPI is used to mock Azure containerClients in unit tests
type azureContainerAPI interface {
	NewBlockBlobClient(blobName string) (*azblob.BlockBlobClient, error)
}

// newBlockBlobClient creates a blockBlobClient from a containerClient
func newBlockBlobClient(blobName string, api azureContainerAPI) (*azblob.BlockBlobClient, error) {
	return api.NewBlockBlobClient(blobName)
}

// uploadFileFunc uploads a file to an Azure Blob Storage
// The function uses the UploadFile function from the Azure SDK
// We introduce this 'wrapper' for mocking reasons
func uploadFileFunc(ctx context.Context, blobClient *azblob.BlockBlobClient, file *os.File, o azblob.UploadOption) (*http.Response, error) {
	return blobClient.UploadFile(ctx, file, o)
}

// Struct to store Azure credentials from specified JSON string
type azureCredentials struct {
	SASToken    string `json:"sas_token" validate:"required"`
	AccountName string `json:"account_name" validate:"required"`
	Container   string `json:"container_name" validate:"required"`
}

func azureBlobUpload(config azureBlobUploadOptions, telemetryData *telemetry.CustomData) {
	err := runAzureBlobUpload(&config)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAzureBlobUpload(config *azureBlobUploadOptions) error {
	containerClient, err := setup(config)
	if err != nil {
		return err
	}
	return executeUpload(config, containerClient, uploadFileFunc)
}

func setup(config *azureBlobUploadOptions) (*azblob.ContainerClient, error) {
	// Read credentials from JSON String
	log.Entry().Infoln("Start reading Azure Credentials")
	var creds azureCredentials

	err := json.Unmarshal([]byte(config.JSONCredentialsAzure), &creds)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return nil, fmt.Errorf("Could not read JSONCredentialsAzure: %w", err)
	}

	// Validate credentials (check for nil values in struct)
	if err = validate(&creds); err != nil {
		return nil, fmt.Errorf("Azure credentials are not valid: %w", err)
	}

	// Initialize Azure Service Client
	sasURL := fmt.Sprintf("https://%s.blob.core.windows.net/?%s", creds.AccountName, creds.SASToken)
	serviceClient, err := azblob.NewServiceClientWithNoCredential(sasURL, nil)
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return nil, fmt.Errorf("Could not instantiate Azure Service Client: %w", err)
	}

	// Get a containerClient from ServiceClient
	containerClient, err := serviceClient.NewContainerClient(creds.Container)
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return nil, fmt.Errorf("Could not instantiate Azure Container Client from Azure Service Client: %w", err)
	}
	return containerClient, nil
}

// Validate validates the Azure credentials (checks for empty fields in struct)
func validate(creds *azureCredentials) error {
	validate := validator.New()
	if err := validate.Struct(creds); err != nil {
		return err
	}
	return nil
}

func executeUpload(config *azureBlobUploadOptions, containerClient azureContainerAPI, uploadFunc func(ctx context.Context, api *azblob.BlockBlobClient, file *os.File, o azblob.UploadOption) (*http.Response, error)) error {
	log.Entry().Infof("Starting walk through FilePath '%v'", config.FilePath)

	// All Blob Operations operate with context.Context, in our case the clients do not expire
	ctx := context.Background()

	// Iterate through directories
	err := filepath.Walk(config.FilePath, func(currentFilePath string, f os.FileInfo, err error) error {
		// Handle Failure to prevent panic (e.g. in case of an invalid filepath)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("Failed to access path: %w", err)
		}
		// Skip directories, only upload files
		if !f.IsDir() {
			log.Entry().Infof("Current target path is: '%v'", currentFilePath)

			//Read Data from File
			data, e := os.Open(currentFilePath)
			if e != nil {
				log.SetErrorCategory(log.ErrorInfrastructure)
				return fmt.Errorf("Could not read the file '%s': %w", currentFilePath, e)
			}
			defer data.Close()

			// Create a filepath in UNIX format so that the BlockBlobClient automatically detects directories
			key := filepath.ToSlash(currentFilePath)

			// Get a blockBlobClient from containerClient
			blockBlobClient, e := newBlockBlobClient(key, containerClient)
			if e != nil {
				log.SetErrorCategory(log.ErrorService)
				return fmt.Errorf("Could not instantiate Azure blockBlobClient from Azure Container Client: %w", e)
			}

			// Upload File
			log.Entry().Infof("Start upload of file '%v'", currentFilePath)
			_, e = uploadFunc(ctx, blockBlobClient, data, azblob.UploadOption{})
			if e != nil {
				log.SetErrorCategory(log.ErrorService)
				return fmt.Errorf("There was an error during the upload of file '%v': %w", currentFilePath, e)
			}

			log.Entry().Infof("Upload of file '%v' was successful!", currentFilePath)
			return e
		}
		return nil
	})

	if err == nil {
		log.Entry().Infoln("Upload has successfully finished!")
	}
	return err
}
