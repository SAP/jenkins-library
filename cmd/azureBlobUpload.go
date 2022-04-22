package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type azureBlobUploadUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The azureBlobUploadUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type azureBlobUploadUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to azureBlobUploadUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// azureBlobUploadUtilsBundle and forward to the implementation of the dependency.
}

func newAzureBlobUploadUtils() azureBlobUploadUtils {
	utils := azureBlobUploadUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

//interface used to mock Azure containerClients in unit tests
type AzureContainerAPI interface {
	NewBlockBlobClient(blobName string) (*azblob.BlockBlobClient, error)
}

func NewBlockBlobClient(blobName string, api AzureContainerAPI) (*azblob.BlockBlobClient, error) {
	return api.NewBlockBlobClient(blobName)
}

func UploadFile(ctx context.Context, api *azblob.BlockBlobClient, file *os.File, o azblob.UploadOption) (*http.Response, error) {
	return api.UploadFile(ctx, file, o)
}

type AzureCredentials struct {
	SAS_Token    string `json:"sas_token"`
	Account_Name string `json:"account_name"`
	Container    string `json:"container_name"`
	Azure_Region string `json:"region"`
}

func azureBlobUpload(config azureBlobUploadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newAzureBlobUploadUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	//Prepare Credentials
	if config.JSONCredentialsAzure == "" {
		log.Entry().Fatal("Azure Credentials were not set!")
	}

	log.Entry().Infoln("Start reading Azure Credentials")
	var obj AzureCredentials

	err := json.Unmarshal([]byte(config.JSONCredentialsAzure), &obj)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Could not read JSONCredentialsAzure")
	}

	//Initialize Azure Service Client
	sasURL := fmt.Sprintf("https://%s.blob.core.windows.net/?%s", obj.Account_Name, obj.SAS_Token)
	serviceClient, err := azblob.NewServiceClientWithNoCredential(sasURL, nil)
	if err != nil {
		log.Entry().WithError(err).Fatal("Could not instantiate Azure Service Client!")
	}

	//Get a containerClient from ServiceClient
	containerClient, err := serviceClient.NewContainerClient(obj.Container)
	if err != nil {
		log.Entry().WithError(err).Fatal("Could not instantiate Azure Container Client from Azure Service Client!")
	}

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err = runAzureBlobUpload(&config, telemetryData, utils, containerClient, UploadFile)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAzureBlobUpload(config *azureBlobUploadOptions, telemetryData *telemetry.CustomData, utils azureBlobUploadUtils, containerClient AzureContainerAPI,
	UploadMock func(ctx context.Context, api *azblob.BlockBlobClient, file *os.File, o azblob.UploadOption) (*http.Response, error)) error {

	if config.FilePath == "" {
		return errors.New("File Path Parameter is empty. Please specify a file or directory to Upload to Azure!")
	}

	// Use UNIX Filepaths
	filePath := filepath.ToSlash(config.FilePath)
	log.Entry().Infof("Start walk through FilePath '%v'", filePath)

	// All Blob Operations operate with context.Context, in our case the clients do not expire
	ctx := context.Background()

	//iterate through directories
	err := filepath.Walk(filePath, func(currentFilePath string, f os.FileInfo, err error) error {
		// Handle Failure to prevent panic (e.g. in case of an invalid filepath)
		if err != nil {
			log.Entry().WithError(err).Warnf("Prevent panic by handling failure accessing the path '%v'", currentFilePath)
			return err
		}
		//skip directories, only upload files
		if !f.IsDir() {
			log.Entry().Infof("Current target path is: '%v'", currentFilePath)

			//Read Data from File
			data, e := os.Open(filepath.ToSlash(currentFilePath))
			if e != nil {
				log.Entry().WithError(e).Warnf("Could not read the file '%s'", currentFilePath)
				return e
			}
			defer data.Close()

			key := filepath.ToSlash(currentFilePath)

			//Get a blockBlobClient from containerClient
			blockBlobClient, e := NewBlockBlobClient(key, containerClient)
			if e != nil {
				log.Entry().WithError(e).Warnf("Could not instantiate Azure blockBlobClient from Azure Container Client!")
				return e
			}

			//Upload File
			log.Entry().Infof("Start upload of file '%v'", currentFilePath)
			var blockOptions azblob.UploadOption
			_, e = UploadMock(ctx, blockBlobClient, data, blockOptions)
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
	log.Entry().Infoln("Upload was successfully finished!")
	return err
}
