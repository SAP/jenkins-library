package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/tms"
)

type tmsUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newTmsUtils() tms.TmsUtils {
	utils := tmsUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func tmsUpload(uploadConfig tmsUploadOptions, telemetryData *telemetry.CustomData, influx *tmsUploadInflux) {
	utils := newTmsUtils()
	config := convertUploadOptions(uploadConfig)
	communicationInstance := tms.SetupCommunication(config)

	err := runTmsUpload(uploadConfig, communicationInstance, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to run tmsUpload step")
	}
}

func runTmsUpload(uploadConfig tmsUploadOptions, communicationInstance tms.CommunicationInterface, utils tms.TmsUtils) error {
	config := convertUploadOptions(uploadConfig)
	fileId, errUploadFile := tms.UploadFile(config, communicationInstance, utils)
	if errUploadFile != nil {
		return errUploadFile
	}

	errUploadDescriptors := tms.UploadDescriptors(config, communicationInstance, utils)
	if errUploadDescriptors != nil {
		return errUploadDescriptors
	}

	_, errUploadFileToNode := communicationInstance.UploadFileToNode(config.NodeName, fileId, config.CustomDescription, config.NamedUser)
	if errUploadFileToNode != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("failed to upload file to node: %w", errUploadFileToNode)
	}

	return nil
}

func convertUploadOptions(uploadConfig tmsUploadOptions) tms.Options {
	var config tms.Options
	config.TmsServiceKey = uploadConfig.TmsServiceKey
	config.CustomDescription = uploadConfig.CustomDescription
	if config.CustomDescription == "" {
		config.CustomDescription = tms.DEFAULT_TR_DESCRIPTION
	}
	config.NamedUser = uploadConfig.NamedUser
	config.NodeName = uploadConfig.NodeName
	config.MtaPath = uploadConfig.MtaPath
	config.MtaVersion = uploadConfig.MtaVersion
	config.NodeExtDescriptorMapping = uploadConfig.NodeExtDescriptorMapping
	config.Proxy = uploadConfig.Proxy
	config.StashContent = uploadConfig.StashContent
	config.Verbose = GeneralConfig.Verbose
	return config
}
