package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/tms"
)

const DEFAULT_DESCRIPTION = "tmsUpload"

func tmsUpload(config tmsUploadOptions, telemetryData *telemetry.CustomData, influx *tmsUploadInflux) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTmsUtils(stepUpload)
	communicationInstance := communicationSetup(config)

	if err := runTmsUpload(config, communicationInstance, utils); err != nil {
		log.Entry().WithError(err).Fatal("Failed to run tmsUpload step")
	}
}

func runTmsUpload(config tmsUploadOptions, communicationInstance tms.CommunicationInterface, utils tmsUtils) error {
	fileId, errUploadFile := fileUpload(config, communicationInstance, utils)
	if errUploadFile != nil {
		return errUploadFile
	}

	errUploadDescriptors := uploadDescriptors(config, communicationInstance, utils)
	if errUploadDescriptors != nil {
		return errUploadDescriptors
	}

	description := DEFAULT_DESCRIPTION
	if config.CustomDescription != "" {
		description = config.CustomDescription
	}
	_, errUploadFileToNode := communicationInstance.UploadFileToNode(config.NodeName, fileId, description, config.NamedUser)
	if errUploadFileToNode != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("failed to upload file to node: %w", errUploadFileToNode)
	}

	return nil
}
