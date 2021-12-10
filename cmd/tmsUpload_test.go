package cmd

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/tms"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const NODE_ID = 777
const ID_OF_MTA_EXT_DESCRIPTOR = 456
const FILE_ID = 333
const NODE_NAME = "TEST_NODE"
const MTA_PATH_LOCAL = "example.mtar"
const MTA_NAME = "example.mtar"
const MTA_ID = "com.sap.tms.upload.test"
const MTA_EXT_ID = "com.sap.tms.upload.test_ext"
const MTA_YAML_PATH_LOCAL = "mta.yaml"
const MTA_YAML_PATH = "./testdata/TestRunTmsUpload/valid/mta.yaml"
const INVALID_MTA_YAML_PATH = "./testdata/TestRunTmsUpload/invalid/mta.yaml"
const MTA_EXT_DESCRIPTOR_PATH_LOCAL = "test.mtaext"
const INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL = "wrong_content.mtaext"
const INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL_2 = "wrong_extends_parameter.mtaext"
const MTA_EXT_DESCRIPTOR_PATH = "./testdata/TestRunTmsUpload/valid/test.mtaext"
const INVALID_MTA_EXT_DESCRIPTOR_PATH = "./testdata/TestRunTmsUpload/invalid/wrong_content.mtaext"
const INVALID_MTA_EXT_DESCRIPTOR_PATH_2 = "./testdata/TestRunTmsUpload/invalid/wrong_extends_parameter.mtaext"
const CUSTOM_DESCRIPTION = "This is a test description"
const NAMED_USER = "techUser"
const MTA_VERSION = "1.0.0"
const WRONG_MTA_VERSION = "3.2.1"
const LAST_CHANGED_AT = "2021-11-16T13:06:05.711Z"
const GIT_COMMIT_ID = "7f654c6d1cabaf6902337fad3865b6f8de876c52"
const INVALID_INPUT_MSG = "Invalid input parameter(s) when getting MTA extension descriptor"

type tmsUploadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTmsUploadTestsUtils() tmsUploadMockUtils {
	utils := tmsUploadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

type gitMock struct {
	commitID string
}

type communicationInstanceMock struct {
	getNodesResponse                      []tms.Node
	getMtaExtDescriptorResponse           tms.MtaExtDescriptor
	updateMtaExtDescriptorResponse        tms.MtaExtDescriptor
	uploadMtaExtDescriptorToNodeResponse  tms.MtaExtDescriptor
	uploadFileResponse                    tms.FileInfo
	uploadFileToNodeResponse              tms.NodeUploadResponseEntity
	isErrorOnGetNodes                     bool
	isErrorOnGetMtaExtDescriptor          bool
	isErrorOnUpdateMtaExtDescriptor       bool
	isErrorOnUploadMtaExtDescriptorToNode bool
	isErrorOnUploadFile                   bool
	isErrorOnUploadFileToNode             bool
	isExpectingParametersFromEnvironment  bool
}

func (cim *communicationInstanceMock) GetNodes() ([]tms.Node, error) {
	if cim.isErrorOnGetNodes {
		var nodes []tms.Node
		return nodes, errors.New("Something went wrong on getting nodes")
	} else {
		return cim.getNodesResponse, nil
	}
}

func (cim *communicationInstanceMock) GetMtaExtDescriptor(nodeId int64, mtaId, mtaVersion string) (tms.MtaExtDescriptor, error) {
	var mtaExtDescriptor tms.MtaExtDescriptor

	// check for input parameters
	// this type of check might not be sufficient/correct, if there are more than one calls of this client method during the single test execution
	if cim.isExpectingParametersFromEnvironment && mtaVersion != "*" || !cim.isExpectingParametersFromEnvironment && mtaVersion != MTA_VERSION {
		return mtaExtDescriptor, errors.New(INVALID_INPUT_MSG)
	}
	if nodeId != NODE_ID || mtaId != MTA_ID {
		return mtaExtDescriptor, errors.New(INVALID_INPUT_MSG)
	}

	if cim.isErrorOnGetMtaExtDescriptor {
		return mtaExtDescriptor, errors.New("Something went wrong on getting MTA extension descriptor")
	} else {
		return cim.getMtaExtDescriptorResponse, nil
	}
}

func (cim *communicationInstanceMock) UpdateMtaExtDescriptor(nodeId, idOfMtaExtDescriptor int64, file, mtaVersion, description, namedUser string) (tms.MtaExtDescriptor, error) {
	var mtaExtDescriptor tms.MtaExtDescriptor

	// check for input parameters
	// this type of check might not be sufficient/correct, if there are more than one calls of this client method during the single test execution
	if cim.isExpectingParametersFromEnvironment && (mtaVersion != "*" || description != fmt.Sprintf("Git commit id: %v", GIT_COMMIT_ID)) || !cim.isExpectingParametersFromEnvironment && (mtaVersion != MTA_VERSION || description != CUSTOM_DESCRIPTION) {
		return mtaExtDescriptor, errors.New(INVALID_INPUT_MSG)
	}
	if nodeId != NODE_ID || idOfMtaExtDescriptor != ID_OF_MTA_EXT_DESCRIPTOR || file != MTA_EXT_DESCRIPTOR_PATH_LOCAL || namedUser != NAMED_USER {
		return mtaExtDescriptor, errors.New(INVALID_INPUT_MSG)
	}

	if cim.isErrorOnUpdateMtaExtDescriptor {
		return mtaExtDescriptor, errors.New("Something went wrong on updating MTA extension descriptor")
	} else {
		return cim.updateMtaExtDescriptorResponse, nil
	}
}

func (cim *communicationInstanceMock) UploadMtaExtDescriptorToNode(nodeId int64, file, mtaVersion, description, namedUser string) (tms.MtaExtDescriptor, error) {
	var mtaExtDescriptor tms.MtaExtDescriptor

	// check for input parameters
	// this type of check might not be sufficient/correct, if there are more than one calls of this client method during the single test execution
	if cim.isExpectingParametersFromEnvironment && (mtaVersion != "*" || description != fmt.Sprintf("Git commit id: %v", GIT_COMMIT_ID)) || !cim.isExpectingParametersFromEnvironment && (mtaVersion != MTA_VERSION || description != CUSTOM_DESCRIPTION) {
		return mtaExtDescriptor, errors.New(INVALID_INPUT_MSG)
	}
	if nodeId != NODE_ID || file != MTA_EXT_DESCRIPTOR_PATH_LOCAL || namedUser != NAMED_USER {
		return mtaExtDescriptor, errors.New(INVALID_INPUT_MSG)
	}

	if cim.isErrorOnUploadMtaExtDescriptorToNode {
		return mtaExtDescriptor, errors.New("Something went wrong on uploading MTA extension descriptor to node")
	} else {
		return cim.uploadMtaExtDescriptorToNodeResponse, nil
	}
}

func (cim *communicationInstanceMock) UploadFile(file, namedUser string) (tms.FileInfo, error) {
	var fileInfo tms.FileInfo

	// check for input parameters
	if file != MTA_PATH_LOCAL || namedUser != NAMED_USER {
		return fileInfo, errors.New(INVALID_INPUT_MSG)
	}

	if cim.isErrorOnUploadFile {
		return fileInfo, errors.New("Something went wrong on uploading file")
	} else {
		return cim.uploadFileResponse, nil
	}
}

func (cim *communicationInstanceMock) UploadFileToNode(nodeName, fileId, description, namedUser string) (tms.NodeUploadResponseEntity, error) {
	var nodeUploadResponseEntity tms.NodeUploadResponseEntity

	// check for input parameters
	if cim.isExpectingParametersFromEnvironment && description != fmt.Sprintf("Git commit id: %v", GIT_COMMIT_ID) || !cim.isExpectingParametersFromEnvironment && description != CUSTOM_DESCRIPTION {
		return nodeUploadResponseEntity, errors.New(INVALID_INPUT_MSG)
	}
	if nodeName != NODE_NAME || fileId != strconv.FormatInt(FILE_ID, 10) || namedUser != NAMED_USER {
		return nodeUploadResponseEntity, errors.New(INVALID_INPUT_MSG)
	}

	if cim.isErrorOnUploadFileToNode {
		return nodeUploadResponseEntity, errors.New("Something went wrong on uploading file to node")
	} else {
		return cim.uploadFileToNodeResponse, nil
	}
}

func TestRunTmsUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path: 1. get nodes 2. get MTA ext descriptor -> nothing obtained 3. upload MTA ext descriptor to node 4. upload file 5. upload file to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		fileInfo := tms.FileInfo{Id: FILE_ID, Name: MTA_NAME}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, uploadFileResponse: fileInfo}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.NoError(t, err)
	})

	t.Run("happy path: 1. get nodes 2. get MTA ext descriptor 3. update the MTA ext descriptor 4. upload file 5. upload file to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		mtaExtDescriptor := tms.MtaExtDescriptor{Id: ID_OF_MTA_EXT_DESCRIPTOR, Description: "Some existing description", MtaId: MTA_ID, MtaExtId: MTA_EXT_ID, MtaVersion: MTA_VERSION, LastChangedAt: LAST_CHANGED_AT}
		fileInfo := tms.FileInfo{Id: FILE_ID, Name: MTA_NAME}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, getMtaExtDescriptorResponse: mtaExtDescriptor, uploadFileResponse: fileInfo}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.NoError(t, err)
	})

	t.Run("happy path: no MtaPath, CustomDescription, MtaVersion provided in configuration", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		mtaExtDescriptor := tms.MtaExtDescriptor{Id: ID_OF_MTA_EXT_DESCRIPTOR, Description: "Some existing description", MtaId: MTA_ID, MtaExtId: MTA_EXT_ID, MtaVersion: MTA_VERSION, LastChangedAt: LAST_CHANGED_AT}
		fileInfo := tms.FileInfo{Id: FILE_ID, Name: MTA_NAME}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, getMtaExtDescriptorResponse: mtaExtDescriptor, uploadFileResponse: fileInfo, isExpectingParametersFromEnvironment: true}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		gitStruct := gitMock{commitID: GIT_COMMIT_ID}
		commonPipelineEnvironment := tmsUploadCommonPipelineEnvironment{mtarFilePath: MTA_PATH_LOCAL, git: gitStruct}

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{NamedUser: NAMED_USER, NodeName: NODE_NAME, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, &commonPipelineEnvironment)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path: MTA file does not exist", func(t *testing.T) {
		t.Parallel()

		// init
		communicationInstance := communicationInstanceMock{}
		utils := newTmsUploadTestsUtils()

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, fmt.Sprintf("mta file %s not found", MTA_PATH_LOCAL))
	})

	t.Run("error path: error while getting nodes", func(t *testing.T) {
		t.Parallel()

		// init
		communicationInstance := communicationInstanceMock{isErrorOnGetNodes: true}
		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to get nodes: Something went wrong on getting nodes")
	})

	t.Run("error path: cannot read mta.yaml (the file is missing)", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes}
		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to get mta.yaml as map: could not read 'mta.yaml'")
	})

	t.Run("error path: cannot unmarshal mta.yaml (the file does not represent a yaml)", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes}
		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(INVALID_MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to get mta.yaml as map: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type map[string]interface {}")
	})

	t.Run("error path: errors on validating the mapping between node names and MTA extension descriptor paths", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes}
		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		invalidMtaExtDescriptorBytes, _ := os.ReadFile(INVALID_MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL, invalidMtaExtDescriptorBytes)

		invalidMtaExtDescriptorBytes2, _ := os.ReadFile(INVALID_MTA_EXT_DESCRIPTOR_PATH_2)
		utils.AddFile(INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL_2, invalidMtaExtDescriptorBytes2)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL, "UNEXISTING_NODE": "unexisting.mtaext", "ONE_MORE_UNEXISTING_NODE": INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL, "ONE_MORE_UNEXISTING_NODE_2": INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL_2}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: WRONG_MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		var expectedErrorMessage string
		expectedErrorMessage += "tried to parse wrong_content.mtaext as yaml, but got an error: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type map[string]interface {}\n"
		expectedErrorMessage += "parameter 'mtaVersion' does not match the MTA version in mta.yaml\n"
		expectedErrorMessage += "parameter 'extends' in MTA extension descriptor files [wrong_extends_parameter.mtaext] is not the same as MTA ID\n"
		expectedErrorMessage += "MTA extension descriptor files [unexisting.mtaext] do not exist\n"
		expectedErrorMessage += "nodes [ONE_MORE_UNEXISTING_NODE ONE_MORE_UNEXISTING_NODE_2 UNEXISTING_NODE] do not exist. Please check node names provided in 'nodeExtDescriptorMapping' parameter or create these nodes\n"
		assert.EqualError(t, err, expectedErrorMessage)
	})

	t.Run("error path: error while getting MTA extension descriptor", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, isErrorOnGetMtaExtDescriptor: true}
		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to get MTA extension descriptor: Something went wrong on getting MTA extension descriptor")
	})

	t.Run("error path: error while updating MTA extension descriptor", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		mtaExtDescriptor := tms.MtaExtDescriptor{Id: ID_OF_MTA_EXT_DESCRIPTOR, Description: "Some existing description", MtaId: MTA_ID, MtaExtId: MTA_EXT_ID, MtaVersion: MTA_VERSION, LastChangedAt: LAST_CHANGED_AT}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, getMtaExtDescriptorResponse: mtaExtDescriptor, isErrorOnUpdateMtaExtDescriptor: true}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to update MTA extension descriptor: Something went wrong on updating MTA extension descriptor")
	})

	t.Run("error path: error while uploading MTA extension descriptor to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, isErrorOnUploadMtaExtDescriptorToNode: true}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to upload MTA extension descriptor to node: Something went wrong on uploading MTA extension descriptor to node")
	})

	t.Run("error path: error while uploading file", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, isErrorOnUploadFile: true}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to upload file: Something went wrong on uploading file")
	})

	t.Run("error path: error while uploading file to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		fileInfo := tms.FileInfo{Id: FILE_ID, Name: MTA_NAME}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, uploadFileResponse: fileInfo, isErrorOnUploadFileToNode: true}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to upload file to node: Something went wrong on uploading file to node")
	})
}
