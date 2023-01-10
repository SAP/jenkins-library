package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/tms"
	"github.com/pkg/errors"
	"net/url"
	"strconv"
)

type tmsUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to tmsUploadUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// tmsUploadUtilsBundle and forward to the implementation of the dependency.
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

func setupCommunication(config tmsUploadOptions) (communicationInstance tms.CommunicationInterface) {
	client := &piperHttp.Client{}
	proxy := config.Proxy
	if proxy != "" {
		transportProxy, err := url.Parse(proxy)
		if err != nil {
			log.Entry().WithError(err).Fatalf("Failed to parse proxy string %v into a URL structure", proxy)
		}

		options := piperHttp.ClientOptions{TransportProxy: transportProxy}
		client.SetOptions(options)
		if GeneralConfig.Verbose {
			log.Entry().Infof("HTTP client instructed to use %v proxy", proxy)
		}
	}

	serviceKey, err := tms.UnmarshalServiceKey(config.TmsServiceKey)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to unmarshal TMS service key")
	}

	if GeneralConfig.Verbose {
		log.Entry().Info("Will be used for communication:")
		log.Entry().Infof("- client id: %v", serviceKey.Uaa.ClientId)
		log.Entry().Infof("- TMS URL: %v", serviceKey.Uri)
		log.Entry().Infof("- UAA URL: %v", serviceKey.Uaa.Url)
	}

	commuInstance, err := tms.NewCommunicationInstance(client, serviceKey.Uri, serviceKey.Uaa.Url, serviceKey.Uaa.ClientId, serviceKey.Uaa.ClientSecret, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to prepare client for talking with TMS")
	}
	return commuInstance
}

func uploadDescriptors(config tmsUploadOptions, communicationInstance tms.CommunicationInterface, utils tms.TmsUtils) error {
	description := tms.DEFAULT_TR_DESCRIPTION
	if config.CustomDescription != "" {
		description = config.CustomDescription
	}

	namedUser := config.NamedUser
	nodeName := config.NodeName
	mtaVersion := config.MtaVersion
	nodeNameExtDescriptorMapping := tms.JsonToMap(config.NodeExtDescriptorMapping)
	mtaPath := config.MtaPath

	if GeneralConfig.Verbose {
		log.Entry().Info("The step will use the following values:")
		log.Entry().Infof("- description: %v", description)

		if len(nodeNameExtDescriptorMapping) > 0 {
			log.Entry().Infof("- mapping between node names and MTA extension descriptor file paths: %v", nodeNameExtDescriptorMapping)
		}
		log.Entry().Infof("- MTA path: %v", mtaPath)
		log.Entry().Infof("- MTA version: %v", mtaVersion)
		if namedUser != "" {
			log.Entry().Infof("- named user: %v", namedUser)
		}
		log.Entry().Infof("- node name: %v", nodeName)
	}

	if len(nodeNameExtDescriptorMapping) > 0 {
		nodes, errGetNodes := communicationInstance.GetNodes()
		if errGetNodes != nil {
			log.SetErrorCategory(log.ErrorService)
			return fmt.Errorf("failed to get nodes: %w", errGetNodes)
		}

		mtaYamlMap, errGetMtaYamlAsMap := tms.GetYamlAsMap(utils, "mta.yaml")
		if errGetMtaYamlAsMap != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("failed to get mta.yaml as map: %w", errGetMtaYamlAsMap)
		}
		_, isIdParameterInMap := mtaYamlMap["ID"]
		_, isVersionParameterInMap := mtaYamlMap["version"]
		if !isIdParameterInMap || !isVersionParameterInMap {
			var errorMessage string
			if !isIdParameterInMap {
				errorMessage += "parameter 'ID' is not found in mta.yaml\n"
			}
			if !isVersionParameterInMap {
				errorMessage += "parameter 'version' is not found in mta.yaml\n"
			}
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.New(errorMessage)
		}

		// validate the whole mapping and then throw errors together, so that user can get them after a single pipeline run
		nodeIdExtDescriptorMapping, errGetNodeIdExtDescriptorMapping := tms.FormNodeIdExtDescriptorMappingWithValidation(utils, nodeNameExtDescriptorMapping, nodes, mtaYamlMap, mtaVersion)
		if errGetNodeIdExtDescriptorMapping != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errGetNodeIdExtDescriptorMapping
		}

		for nodeId, mtaExtDescriptorPath := range nodeIdExtDescriptorMapping {
			obtainedMtaExtDescriptor, errGetMtaExtDescriptor := communicationInstance.GetMtaExtDescriptor(nodeId, fmt.Sprintf("%v", mtaYamlMap["ID"]), mtaVersion)
			if errGetMtaExtDescriptor != nil {
				log.SetErrorCategory(log.ErrorService)
				return fmt.Errorf("failed to get MTA extension descriptor: %w", errGetMtaExtDescriptor)
			}

			if obtainedMtaExtDescriptor != (tms.MtaExtDescriptor{}) {
				_, errUpdateMtaExtDescriptor := communicationInstance.UpdateMtaExtDescriptor(nodeId, obtainedMtaExtDescriptor.Id, mtaExtDescriptorPath, mtaVersion, description, namedUser)
				if errUpdateMtaExtDescriptor != nil {
					log.SetErrorCategory(log.ErrorService)
					return fmt.Errorf("failed to update MTA extension descriptor: %w", errUpdateMtaExtDescriptor)
				}
			} else {
				_, errUploadMtaExtDescriptor := communicationInstance.UploadMtaExtDescriptorToNode(nodeId, mtaExtDescriptorPath, mtaVersion, description, namedUser)
				if errUploadMtaExtDescriptor != nil {
					log.SetErrorCategory(log.ErrorService)
					return fmt.Errorf("failed to upload MTA extension descriptor to node: %w", errUploadMtaExtDescriptor)
				}
			}
		}
	}
	return nil
}

func tmsUploadFile(config tmsUploadOptions, communicationInstance tms.CommunicationInterface, utils tms.TmsUtils) (string, error) {
	mtaPath := config.MtaPath
	exists, _ := utils.FileExists(mtaPath)
	if !exists {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", fmt.Errorf("mta file %s not found", mtaPath)
	}

	fileInfo, errUploadFile := communicationInstance.UploadFile(mtaPath, config.NamedUser)
	if errUploadFile != nil {
		log.SetErrorCategory(log.ErrorService)
		return "", fmt.Errorf("failed to upload file: %w", errUploadFile)
	}

	fileId := strconv.FormatInt(fileInfo.Id, 10)
	return fileId, nil
}

func tmsUpload(config tmsUploadOptions, telemetryData *telemetry.CustomData, influx *tmsUploadInflux) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTmsUtils()
	communicationInstance := setupCommunication(config)

	if err := runTmsUpload(config, communicationInstance, utils); err != nil {
		log.Entry().WithError(err).Fatal("Failed to run tmsUpload step")
	}
}

func runTmsUpload(config tmsUploadOptions, communicationInstance tms.CommunicationInterface, utils tms.TmsUtils) error {
	fileId, errUploadFile := tmsUploadFile(config, communicationInstance, utils)
	if errUploadFile != nil {
		return errUploadFile
	}

	errUploadDescriptors := uploadDescriptors(config, communicationInstance, utils)
	if errUploadDescriptors != nil {
		return errUploadDescriptors
	}

	description := tms.DEFAULT_TR_DESCRIPTION
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
