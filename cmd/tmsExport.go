package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/tms"
)

const DEFAULT_DESCRIPTION_EXPORT = "tmsExport"

type tmsExportUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The tmsExportUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type tmsExportUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to tmsExportUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// tmsExportUtilsBundle and forward to the implementation of the dependency.
}

func newTmsExportUtils() tmsExportUtils {
	utils := tmsExportUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func tmsExport(config tmsExportOptions, telemetryData *telemetry.CustomData, influx *tmsExportInflux) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTmsUtils()
	var uploadConfig tmsUploadOptions
	uploadConfig.TmsServiceKey = config.TmsServiceKey
	uploadConfig.CustomDescription = config.CustomDescription
	uploadConfig.NamedUser = config.NamedUser
	uploadConfig.NodeName = config.NodeName
	uploadConfig.MtaPath = config.MtaPath
	uploadConfig.MtaVersion = config.MtaVersion
	uploadConfig.NodeExtDescriptorMapping = config.NodeExtDescriptorMapping
	uploadConfig.Proxy = config.Proxy
	uploadConfig.StashContent = config.StashContent

	communicationInstance := setupCommunication(uploadConfig)

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runTmsExport(uploadConfig, communicationInstance, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to run tmsExport")
	}
}

func runTmsExport(config tmsUploadOptions, communicationInstance tms.CommunicationInterface, utils tmsUtils) error {
	fileId, errUploadFile := fileUpload(config, communicationInstance, utils)
	if errUploadFile != nil {
		return errUploadFile
	}

	errUploadDescriptors := uploadDescriptors(config, communicationInstance, utils)
	if errUploadDescriptors != nil {
		return errUploadDescriptors
	}

	description := DEFAULT_DESCRIPTION_EXPORT
	if config.CustomDescription != "" {
		description = config.CustomDescription
	}
	_, errExportFileToNode := communicationInstance.ExportFileToNode(config.NodeName, fileId, description, config.NamedUser)
	if errExportFileToNode != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("failed to export file to node: %w", errExportFileToNode)
	}

	return nil
}
