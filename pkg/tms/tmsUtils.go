package tms

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"

	"errors"

	"github.com/SAP/jenkins-library/pkg/command"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
)

type TmsUtils interface {
	command.ExecRunner
	FileExists(filename string) (bool, error)
	FileRead(path string) ([]byte, error)
}

type uaa struct {
	Url          string `json:"url"`
	ClientId     string `json:"clientid"`
	ClientSecret string `json:"clientsecret"`
}

type serviceKey struct {
	Uaa           uaa           `json:"uaa"`
	Uri           string        `json:"uri"`
	CALMEndpoints cALMEndpoints `json:"endpoints"`
}

type cALMEndpoints *struct {
	API string `json:"Api"`
}

type CommunicationInstance struct {
	tmsUrl       string
	uaaUrl       string
	clientId     string
	clientSecret string
	httpClient   piperHttp.Uploader
	logger       *logrus.Entry
	isVerbose    bool
}

type Node struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type nodes struct {
	Nodes []Node `json:"nodes"`
}

type MtaExtDescriptor struct {
	Id            int64  `json:"id"`
	Description   string `json:"description"`
	MtaId         string `json:"mtaId"`
	MtaExtId      string `json:"mtaExtId"`
	MtaVersion    string `json:"mtaVersion"`
	LastChangedAt string `json:"lastChangedAt"`
}

type mtaExtDescriptors struct {
	MtaExtDescriptors []MtaExtDescriptor `json:"mtaExtDescriptors"`
}

type FileInfo struct {
	Id   int64  `json:"fileId"`
	Name string `json:"fileName"`
}

type NodeUploadResponseEntity struct {
	TransportRequestId          int64        `json:"transportRequestId"`
	TransportRequestDescription string       `json:"transportRequestDescription"`
	QueueEntries                []QueueEntry `json:"queueEntries"`
}

type QueueEntry struct {
	Id       int64  `json:"queueId"`
	NodeId   int64  `json:"nodeId"`
	NodeName string `json:"nodeName"`
}

type NodeUploadRequestEntity struct {
	ContentType string  `json:"contentType"`
	StorageType string  `json:"storageType"`
	NodeName    string  `json:"nodeName"`
	Description string  `json:"description"`
	NamedUser   string  `json:"namedUser"`
	Entries     []Entry `json:"entries"`
}

type Entry struct {
	Uri string `json:"uri"`
}

type CommunicationInterface interface {
	GetNodes() ([]Node, error)
	GetMtaExtDescriptor(nodeId int64, mtaId, mtaVersion string) (MtaExtDescriptor, error)
	UpdateMtaExtDescriptor(nodeId, idOfMtaExtDescriptor int64, file, mtaVersion, description, namedUser string) (MtaExtDescriptor, error)
	UploadMtaExtDescriptorToNode(nodeId int64, file, mtaVersion, description, namedUser string) (MtaExtDescriptor, error)
	UploadFile(file, namedUser string) (FileInfo, error)
	UploadFileToNode(fileInfo FileInfo, nodeName, description, namedUser string) (NodeUploadResponseEntity, error)
	ExportFileToNode(fileInfo FileInfo, nodeName, description, namedUser string) (NodeUploadResponseEntity, error)
}

type Options struct {
	ServiceKey               string
	CustomDescription        string
	NamedUser                string
	NodeName                 string
	MtaPath                  string
	MtaVersion               string
	NodeExtDescriptorMapping map[string]interface{}
	Proxy                    string
	StashContent             []string
	Verbose                  bool
}

type tmsUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

const DEFAULT_TR_DESCRIPTION = "Created by Piper"
const CALM_REROUTING_ENDPOINT_TO_CTMS = "/imp-cdm-transport-management-api/v1"

func NewTmsUtils() TmsUtils {
	utils := tmsUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func unmarshalServiceKey(serviceKeyJson string) (serviceKey serviceKey, err error) {
	err = json.Unmarshal([]byte(serviceKeyJson), &serviceKey)
	if err != nil {
		return
	}
	if len(serviceKey.Uri) == 0 {
		if serviceKey.CALMEndpoints != nil && len(serviceKey.CALMEndpoints.API) > 0 {
			serviceKey.Uri = serviceKey.CALMEndpoints.API + CALM_REROUTING_ENDPOINT_TO_CTMS
		} else {
			err = fmt.Errorf("neither uri nor endpoints.Api is set in service key json string")
			return
		}
	}
	return
}

func FormNodeIdExtDescriptorMappingWithValidation(utils TmsUtils, nodeNameExtDescriptorMapping map[string]interface{}, nodes []Node, mtaYamlMap map[string]interface{}, mtaVersion string) (map[int64]string, error) {
	var wrongMtaIdExtDescriptors []string
	var wrongExtDescriptorPaths []string
	var wrongNodeNames []string
	var errorMessage string

	nodeIdExtDescriptorMapping := make(map[int64]string)
	for nodeName, mappedValue := range nodeNameExtDescriptorMapping {
		mappedValueString := fmt.Sprintf("%v", mappedValue)
		exists, _ := utils.FileExists(mappedValueString)
		if exists {
			extDescriptorMap, errGetYamlAsMap := GetYamlAsMap(utils, mappedValueString)
			if errGetYamlAsMap == nil {
				if fmt.Sprintf("%v", mtaYamlMap["ID"]) != fmt.Sprintf("%v", extDescriptorMap["extends"]) {
					wrongMtaIdExtDescriptors = append(wrongMtaIdExtDescriptors, mappedValueString)
				}
			} else {
				wrappedErr := fmt.Errorf("tried to parse %v as yaml, but got an error: %w", mappedValueString, errGetYamlAsMap)
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

func GetYamlAsMap(utils TmsUtils, yamlPath string) (map[string]interface{}, error) {
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

func SetupCommunication(config Options) (communicationInstance CommunicationInterface) {
	client := &piperHttp.Client{}
	proxy := config.Proxy
	options := piperHttp.ClientOptions{}
	if proxy != "" {
		transportProxy, err := url.Parse(proxy)
		if err != nil {
			log.Entry().WithError(err).Fatalf("Failed to parse proxy string %v into a URL structure", proxy)
		}

		options = piperHttp.ClientOptions{TransportProxy: transportProxy}
		client.SetOptions(options)
		if config.Verbose {
			log.Entry().Infof("HTTP client instructed to use %v proxy", proxy)
		}
	}

	serviceKey, err := unmarshalServiceKey(config.ServiceKey)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to unmarshal service key")
	}
	log.RegisterSecret(serviceKey.Uaa.ClientSecret)

	if config.Verbose {
		log.Entry().Info("Will be used for communication:")
		log.Entry().Infof("- client id: %v", serviceKey.Uaa.ClientId)
		log.Entry().Infof("- TMS URL: %v", serviceKey.Uri)
		log.Entry().Infof("- UAA URL: %v", serviceKey.Uaa.Url)
	}

	commuInstance, err := NewCommunicationInstance(client, serviceKey.Uri, serviceKey.Uaa.Url, serviceKey.Uaa.ClientId, serviceKey.Uaa.ClientSecret, config.Verbose, options)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to prepare client for talking with TMS")
	}
	return commuInstance
}

func UploadDescriptors(config Options, communicationInstance CommunicationInterface, utils TmsUtils) error {
	description := config.CustomDescription
	namedUser := config.NamedUser
	nodeName := config.NodeName
	mtaVersion := config.MtaVersion
	nodeNameExtDescriptorMapping := config.NodeExtDescriptorMapping
	mtaPath := config.MtaPath

	if config.Verbose {
		log.Entry().Info("The step will use the following values:")
		log.Entry().Infof("- description: %v", config.CustomDescription)

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

		mtaYamlMap, errGetMtaYamlAsMap := GetYamlAsMap(utils, "mta.yaml")
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
		nodeIdExtDescriptorMapping, errGetNodeIdExtDescriptorMapping := FormNodeIdExtDescriptorMappingWithValidation(utils, nodeNameExtDescriptorMapping, nodes, mtaYamlMap, mtaVersion)
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

			if obtainedMtaExtDescriptor != (MtaExtDescriptor{}) {
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

func UploadFile(config Options, communicationInstance CommunicationInterface, utils TmsUtils) (FileInfo, error) {
	var fileInfo FileInfo

	mtaPath := config.MtaPath
	exists, _ := utils.FileExists(mtaPath)
	if !exists {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fileInfo, fmt.Errorf("mta file %s not found", mtaPath)
	}

	fileInfo, errUploadFile := communicationInstance.UploadFile(mtaPath, config.NamedUser)
	if errUploadFile != nil {
		log.SetErrorCategory(log.ErrorService)
		return fileInfo, fmt.Errorf("failed to upload file: %w", errUploadFile)
	}

	return fileInfo, nil
}
