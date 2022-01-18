package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"

	"github.com/SAP/jenkins-library/pkg/command"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/tms"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

type uaa struct {
	Url          string `json:"url"`
	ClientId     string `json:"clientid"`
	ClientSecret string `json:"clientsecret"`
}

type serviceKey struct {
	Uaa uaa    `json:"uaa"`
	Uri string `json:"uri"`
}

type tmsUploadUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
	FileRead(path string) ([]byte, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The tmsUploadUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type tmsUploadUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to tmsUploadUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// tmsUploadUtilsBundle and forward to the implementation of the dependency.
}

func newTmsUploadUtils() tmsUploadUtils {
	utils := tmsUploadUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func tmsUpload(config tmsUploadOptions, telemetryData *telemetry.CustomData, influx *tmsUploadInflux) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTmsUploadUtils()
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

	serviceKey, err := unmarshalServiceKey(config.TmsServiceKey)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to unmarshal TMS service key")
	}

	if GeneralConfig.Verbose {
		log.Entry().Info("Will be used for communication:")
		log.Entry().Infof("- client id: %v", serviceKey.Uaa.ClientId)
		log.Entry().Infof("- TMS URL: %v", serviceKey.Uri)
		log.Entry().Infof("- UAA URL: %v", serviceKey.Uaa.Url)
	}

	communicationInstance, err := tms.NewCommunicationInstance(client, serviceKey.Uri, serviceKey.Uaa.Url, serviceKey.Uaa.ClientId, serviceKey.Uaa.ClientSecret, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to prepare client for talking with TMS")
	}

	if err := runTmsUpload(config, communicationInstance, utils); err != nil {
		log.Entry().WithError(err).Fatal("Failed to run tmsUpload step")
	}
}

func runTmsUpload(config tmsUploadOptions, communicationInstance tms.CommunicationInterface, utils tmsUploadUtils) error {
	mtaPath := config.MtaPath
	exists, _ := utils.FileExists(mtaPath)
	if !exists {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("mta file %s not found", mtaPath)
	}

	description := config.CustomDescription
	namedUser := config.NamedUser
	nodeName := config.NodeName
	mtaVersion := config.MtaVersion
	nodeNameExtDescriptorMapping := config.NodeExtDescriptorMapping

	if GeneralConfig.Verbose {
		log.Entry().Info("The step will use the following values:")
		log.Entry().Infof("- description: %v", description)

		if len(nodeNameExtDescriptorMapping) != 0 {
			log.Entry().Infof("- mapping between node names and MTA extension descriptor file paths: %v", nodeNameExtDescriptorMapping)
		}
		log.Entry().Infof("- MTA path: %v", mtaPath)
		log.Entry().Infof("- MTA version: %v", mtaVersion)
		if namedUser != "" {
			log.Entry().Infof("- named user: %v", namedUser)
		}
		log.Entry().Infof("- node name: %v", nodeName)
	}

	if len(nodeNameExtDescriptorMapping) != 0 {
		nodes, errGetNodes := communicationInstance.GetNodes()
		if errGetNodes != nil {
			log.SetErrorCategory(log.ErrorService)
			return fmt.Errorf("failed to get nodes: %w", errGetNodes)
		}

		mtaYamlMap, errGetMtaYamlAsMap := getYamlAsMap(utils, "mta.yaml")
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
		nodeIdExtDescriptorMapping, errGetNodeIdExtDescriptorMapping := formNodeIdExtDescriptorMappingWithValidation(utils, nodeNameExtDescriptorMapping, nodes, mtaYamlMap, mtaVersion)
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

			idOfMtaExtDescriptor := obtainedMtaExtDescriptor.Id
			if idOfMtaExtDescriptor != int64(0) {
				_, errUpdateMtaExtDescriptor := communicationInstance.UpdateMtaExtDescriptor(nodeId, idOfMtaExtDescriptor, mtaExtDescriptorPath, mtaVersion, description, namedUser)
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

	fileInfo, errUploadFile := communicationInstance.UploadFile(mtaPath, namedUser)
	if errUploadFile != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("failed to upload file: %w", errUploadFile)
	}

	_, errUploadFileToNode := communicationInstance.UploadFileToNode(nodeName, strconv.FormatInt(fileInfo.Id, 10), description, namedUser)
	if errUploadFileToNode != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("failed to upload file to node: %w", errUploadFileToNode)
	}

	return nil
}

func formNodeIdExtDescriptorMappingWithValidation(utils tmsUploadUtils, nodeNameExtDescriptorMapping map[string]interface{}, nodes []tms.Node, mtaYamlMap map[string]interface{}, mtaVersion string) (map[int64]string, error) {
	var wrongMtaIdExtDescriptors []string
	var wrongExtDescriptorPaths []string
	var wrongNodeNames []string
	var errorMessage string

	nodeIdExtDescriptorMapping := make(map[int64]string)
	for nodeName, mappedValue := range nodeNameExtDescriptorMapping {
		mappedValueString := fmt.Sprintf("%v", mappedValue)
		exists, _ := utils.FileExists(mappedValueString)
		if exists {
			extDescriptorMap, errGetYamlAsMap := getYamlAsMap(utils, mappedValueString)
			if errGetYamlAsMap == nil {
				if fmt.Sprintf("%v", mtaYamlMap["ID"]) != fmt.Sprintf("%v", extDescriptorMap["extends"]) {
					wrongMtaIdExtDescriptors = append(wrongMtaIdExtDescriptors, mappedValueString)
				}
			} else {
				wrappedErr := errors.Wrapf(errGetYamlAsMap, "tried to parse %v as yaml, but got an error", mappedValueString)
				errorMessage += fmt.Sprintf("%v\n", wrappedErr)
			}
		} else {
			wrongExtDescriptorPaths = append(wrongExtDescriptorPaths, mappedValueString)
		}

		isNodeFound := false
		for _, node := range nodes {
			if node.Name == nodeName {
				nodeIdExtDescriptorMapping[node.Id] = mappedValueString
				isNodeFound = true
				break
			}
		}
		if !isNodeFound {
			wrongNodeNames = append(wrongNodeNames, nodeName)
		}
	}

	if mtaVersion != "*" && mtaVersion != mtaYamlMap["version"] {
		errorMessage += "parameter 'mtaVersion' does not match the MTA version in mta.yaml\n"
	}

	if len(wrongMtaIdExtDescriptors) > 0 || len(wrongExtDescriptorPaths) > 0 || len(wrongNodeNames) > 0 {
		if len(wrongMtaIdExtDescriptors) > 0 {
			sort.Strings(wrongMtaIdExtDescriptors)
			errorMessage += fmt.Sprintf("parameter 'extends' in MTA extension descriptor files %v is not the same as MTA ID or is missing at all\n", wrongMtaIdExtDescriptors)
		}
		if len(wrongExtDescriptorPaths) > 0 {
			sort.Strings(wrongExtDescriptorPaths)
			errorMessage += fmt.Sprintf("MTA extension descriptor files %v do not exist\n", wrongExtDescriptorPaths)
		}
		if len(wrongNodeNames) > 0 {
			sort.Strings(wrongNodeNames)
			errorMessage += fmt.Sprintf("nodes %v do not exist. Please check node names provided in 'nodeExtDescriptorMapping' parameter or create these nodes\n", wrongNodeNames)
		}
	}

	if errorMessage == "" {
		return nodeIdExtDescriptorMapping, nil
	} else {
		return nil, errors.New(errorMessage)
	}
}

func getYamlAsMap(utils tmsUploadUtils, yamlPath string) (map[string]interface{}, error) {
	var result map[string]interface{}
	bytes, err := utils.FileRead(yamlPath)
	if err != nil {
		return result, err
	}
	err = yaml.Unmarshal(bytes, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func unmarshalServiceKey(serviceKeyJson string) (serviceKey serviceKey, err error) {
	err = json.Unmarshal([]byte(serviceKeyJson), &serviceKey)
	if err != nil {
		return
	}
	return
}
