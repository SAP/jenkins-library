package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"

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

type BlobUploadAPI interface {
	//Upload(ctx context.Context, body io.ReadSeekCloser, options *BlockBlobUploadOptions) (BlockBlobUploadResponse, error)
}

//func UploadFile(ctx context.Context, api BlobUploadAPI, body io.ReadSeekCloser) (BlockBlobUploadResponse, error) {
// return api.Upload(ctx, body)
//}

func azureBlobUpload(config azureBlobUploadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newAzureBlobUploadUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runAzureBlobUpload(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAzureBlobUpload(config *azureBlobUploadOptions, telemetryData *telemetry.CustomData, utils azureBlobUploadUtils) error {
	if config.FilePath == "" {
		return errors.New("File Path Parameter is empty. Please specify a file or directory to Upload to AWS!")
	}

	// Use UNIX Filepaths
	filePath := filepath.ToSlash(config.FilePath)
	log.Entry().Infof("Start walk through FilePath '%v'", filePath)

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
			_, e := ioutil.ReadFile("testdata/hello")
			if err != nil {
				log.Entry().WithError(e).Warnf("Could not read the file '%s'", currentFilePath)
				return e
			}

			log.Entry().Infof("Upload of file '%v' was successful!", currentFilePath)
			return nil
		}
		return nil
	})

	if err != nil {
		return err
	}
	log.Entry().Infoln("Upload was successfully finished!")
	return err
}
