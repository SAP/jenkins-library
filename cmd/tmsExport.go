package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/tms"
)

type tmsExportUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func tmsExport(exportConfig tmsExportOptions, telemetryData *telemetry.CustomData, influx *tmsExportInflux) {
	utils := tms.NewTmsUtils()
	config := convertExportOptions(exportConfig)
	communicationInstance := tms.SetupCommunication(config)

	err := runTmsExport(exportConfig, communicationInstance, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to run tmsExport")
	}
}

func runTmsExport(exportConfig tmsExportOptions, communicationInstance tms.CommunicationInterface, utils tms.TmsUtils) error {
	config := convertExportOptions(exportConfig)
	fileInfo, errUploadFile := tms.UploadFile(config, communicationInstance, utils)
	if errUploadFile != nil {
		return errUploadFile
	}

	errUploadDescriptors := tms.UploadDescriptors(config, communicationInstance, utils)
	if errUploadDescriptors != nil {
		return errUploadDescriptors
	}

	_, errExportFileToNode := communicationInstance.ExportFileToNode(fileInfo, config.NodeName, config.CustomDescription, config.NamedUser)
	if errExportFileToNode != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("failed to export file to node: %w", errExportFileToNode)
	}

	return nil
}

func convertExportOptions(exportConfig tmsExportOptions) tms.Options {
	var config tms.Options
	config.ServiceKey = exportConfig.ServiceKey
	if exportConfig.ServiceKey == "" && exportConfig.TmsServiceKey != "" {
		config.ServiceKey = exportConfig.TmsServiceKey
		log.Entry().Warn("DEPRECATION WARNING: The tmsServiceKey parameter has been deprecated, please use the serviceKey parameter instead.")
	}
	config.CustomDescription = exportConfig.CustomDescription
	if config.CustomDescription == "" {
		config.CustomDescription = tms.DEFAULT_TR_DESCRIPTION
	}
	config.NamedUser = exportConfig.NamedUser
	config.NodeName = exportConfig.NodeName
	config.MtaPath = exportConfig.MtaPath
	config.MtaVersion = exportConfig.MtaVersion
	config.NodeExtDescriptorMapping = exportConfig.NodeExtDescriptorMapping
	config.Proxy = exportConfig.Proxy
	config.Verbose = GeneralConfig.Verbose
	return config
}
