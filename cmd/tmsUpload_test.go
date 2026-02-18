//go:build unit

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"errors"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/tms"
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
const INVALID_MTA_YAML_PATH = "./testdata/TestRunTmsUpload/invalid/mta_not_a_yaml.yaml"
const INVALID_MTA_YAML_PATH_2 = "./testdata/TestRunTmsUpload/invalid/mta_no_id_and_version_parameters.yaml"
const MTA_EXT_DESCRIPTOR_PATH_LOCAL = "test.mtaext"
const INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL = "wrong_content.mtaext"
const INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL_2 = "wrong_extends_parameter.mtaext"
const INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL_3 = "missing_extends_parameter.mtaext"
const MTA_EXT_DESCRIPTOR_PATH = "./testdata/TestRunTmsUpload/valid/test.mtaext"
const INVALID_MTA_EXT_DESCRIPTOR_PATH = "./testdata/TestRunTmsUpload/invalid/wrong_content.mtaext"
const INVALID_MTA_EXT_DESCRIPTOR_PATH_2 = "./testdata/TestRunTmsUpload/invalid/wrong_extends_parameter.mtaext"
const INVALID_MTA_EXT_DESCRIPTOR_PATH_3 = "./testdata/TestRunTmsUpload/invalid/missing_extends_parameter.mtaext"
const CUSTOM_DESCRIPTION = "This is a test description"
const NAMED_USER = "techUser"
const MTA_VERSION = "1.0.0"
const WRONG_MTA_VERSION = "3.2.1"
const LAST_CHANGED_AT = "2021-11-16T13:06:05.711Z"
const INVALID_INPUT_MSG = "Invalid input parameter(s) when getting MTA extension descriptor"

type tmsMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTmsTestsUtils() tmsMockUtils {
	utils := tmsMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

type communicationInstanceMock struct {
	getNodesResponse                      []tms.Node
	getMtaExtDescriptorResponse           tms.MtaExtDescriptor
	updateMtaExtDescriptorResponse        tms.MtaExtDescriptor
	uploadMtaExtDescriptorToNodeResponse  tms.MtaExtDescriptor
	uploadFileResponse                    tms.FileInfo
	uploadFileToNodeResponse              tms.NodeUploadResponseEntity
	exportFileToNodeResponse              tms.NodeUploadResponseEntity
	isErrorOnGetNodes                     bool
	isErrorOnGetMtaExtDescriptor          bool
	isErrorOnUpdateMtaExtDescriptor       bool
	isErrorOnUploadMtaExtDescriptorToNode bool
	isErrorOnUploadFile                   bool
	isErrorOnUploadFileToNode             bool
	isErrorOnExportFileToNode             bool
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
	if mtaVersion != MTA_VERSION || nodeId != NODE_ID || mtaId != MTA_ID {
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
	if mtaVersion != MTA_VERSION || description != CUSTOM_DESCRIPTION || nodeId != NODE_ID || idOfMtaExtDescriptor != ID_OF_MTA_EXT_DESCRIPTOR || file != MTA_EXT_DESCRIPTOR_PATH_LOCAL || namedUser != NAMED_USER {
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
	if mtaVersion != MTA_VERSION || description != CUSTOM_DESCRIPTION || nodeId != NODE_ID || file != MTA_EXT_DESCRIPTOR_PATH_LOCAL || namedUser != NAMED_USER {
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
	if file != MTA_PATH_LOCAL || namedUser != NAMED_USER {
		return fileInfo, errors.New(INVALID_INPUT_MSG)
	}

	if cim.isErrorOnUploadFile {
		return fileInfo, errors.New("Something went wrong on uploading file")
	} else {
		return cim.uploadFileResponse, nil
	}
}

func (cim *communicationInstanceMock) UploadFileToNode(fileInfo tms.FileInfo, nodeName, description, namedUser string) (tms.NodeUploadResponseEntity, error) {
	fileId := strconv.FormatInt(fileInfo.Id, 10)
	var nodeUploadResponseEntity tms.NodeUploadResponseEntity
	if description != CUSTOM_DESCRIPTION || nodeName != NODE_NAME || fileId != strconv.FormatInt(FILE_ID, 10) || namedUser != NAMED_USER {
		return nodeUploadResponseEntity, errors.New(INVALID_INPUT_MSG)
	}

	if cim.isErrorOnUploadFileToNode {
		return nodeUploadResponseEntity, errors.New("Something went wrong on uploading file to node")
	} else {
		return cim.uploadFileToNodeResponse, nil
	}
}

func mapToJson(m map[string]interface{}) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func TestRunTmsUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path: 1. get nodes 2. get MTA ext descriptor -> nothing obtained 3. upload MTA ext descriptor to node 4. upload file 5. upload file to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		fileInfo := tms.FileInfo{Id: FILE_ID, Name: MTA_NAME}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, uploadFileResponse: fileInfo}

		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("happy path: no mapping between node names and MTA extension descriptors is provided -> only upload file and upload file to node calls will be executed", func(t *testing.T) {
		t.Parallel()

		// init
		fileInfo := tms.FileInfo{Id: FILE_ID, Name: MTA_NAME}
		communicationInstance := communicationInstanceMock{uploadFileResponse: fileInfo}

		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

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

		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path: MTA file does not exist", func(t *testing.T) {
		t.Parallel()

		// init
		communicationInstance := communicationInstanceMock{}
		utils := newTmsTestsUtils()

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, fmt.Sprintf("mta file %s not found", MTA_PATH_LOCAL))
	})

	t.Run("error path: error while getting nodes", func(t *testing.T) {
		t.Parallel()

		// init
		communicationInstance := communicationInstanceMock{isErrorOnGetNodes: true}
		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, "failed to get nodes: Something went wrong on getting nodes")
	})

	t.Run("error path: cannot read mta.yaml (the file is missing)", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes}
		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, "failed to get mta.yaml as map: could not read 'mta.yaml'")
	})

	t.Run("error path: cannot unmarshal mta.yaml (the file does not represent a yaml)", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes}
		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(INVALID_MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, "failed to get mta.yaml as map: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type map[string]interface {}")
	})

	t.Run("error path: no 'ID' and 'version' parameters found in mta.yaml", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes}
		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(INVALID_MTA_YAML_PATH_2)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		var expectedErrorMessage string
		expectedErrorMessage += "parameter 'ID' is not found in mta.yaml\n"
		expectedErrorMessage += "parameter 'version' is not found in mta.yaml\n"

		assert.EqualError(t, err, expectedErrorMessage)
	})

	t.Run("error path: errors on validating the mapping between node names and MTA extension descriptor paths", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes}
		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		invalidMtaExtDescriptorBytes, _ := os.ReadFile(INVALID_MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL, invalidMtaExtDescriptorBytes)

		invalidMtaExtDescriptorBytes2, _ := os.ReadFile(INVALID_MTA_EXT_DESCRIPTOR_PATH_2)
		utils.AddFile(INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL_2, invalidMtaExtDescriptorBytes2)

		invalidMtaExtDescriptorBytes3, _ := os.ReadFile(INVALID_MTA_EXT_DESCRIPTOR_PATH_3)
		utils.AddFile(INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL_3, invalidMtaExtDescriptorBytes3)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL, "UNEXISTING_NODE": "unexisting.mtaext", "ONE_MORE_UNEXISTING_NODE": INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL, "ONE_MORE_UNEXISTING_NODE_2": INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL_2, "ONE_MORE_UNEXISTING_NODE_3": INVALID_MTA_EXT_DESCRIPTOR_PATH_LOCAL_3}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: WRONG_MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		var expectedErrorMessage string
		expectedErrorMessage += "tried to parse wrong_content.mtaext as yaml, but got an error: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type map[string]interface {}\n"
		expectedErrorMessage += "parameter 'mtaVersion' does not match the MTA version in mta.yaml\n"
		expectedErrorMessage += "parameter 'extends' in MTA extension descriptor files [missing_extends_parameter.mtaext wrong_extends_parameter.mtaext] is not the same as MTA ID or is missing at all\n"
		expectedErrorMessage += "MTA extension descriptor files [unexisting.mtaext] do not exist\n"
		expectedErrorMessage += "nodes [ONE_MORE_UNEXISTING_NODE ONE_MORE_UNEXISTING_NODE_2 ONE_MORE_UNEXISTING_NODE_3 UNEXISTING_NODE] do not exist. Please check node names provided in 'nodeExtDescriptorMapping' parameter or create these nodes\n"
		assert.EqualError(t, err, expectedErrorMessage)
	})

	t.Run("error path: error while getting MTA extension descriptor", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, isErrorOnGetMtaExtDescriptor: true}
		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, "failed to get MTA extension descriptor: Something went wrong on getting MTA extension descriptor")
	})

	t.Run("error path: error while updating MTA extension descriptor", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		mtaExtDescriptor := tms.MtaExtDescriptor{Id: ID_OF_MTA_EXT_DESCRIPTOR, Description: "Some existing description", MtaId: MTA_ID, MtaExtId: MTA_EXT_ID, MtaVersion: MTA_VERSION, LastChangedAt: LAST_CHANGED_AT}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, getMtaExtDescriptorResponse: mtaExtDescriptor, isErrorOnUpdateMtaExtDescriptor: true}

		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, "failed to update MTA extension descriptor: Something went wrong on updating MTA extension descriptor")
	})

	t.Run("error path: error while uploading MTA extension descriptor to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, isErrorOnUploadMtaExtDescriptorToNode: true}

		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, "failed to upload MTA extension descriptor to node: Something went wrong on uploading MTA extension descriptor to node")
	})

	t.Run("error path: error while uploading file", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, isErrorOnUploadFile: true}

		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, "failed to upload file: Something went wrong on uploading file")
	})

	t.Run("error path: error while uploading file to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		fileInfo := tms.FileInfo{Id: FILE_ID, Name: MTA_NAME}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, uploadFileResponse: fileInfo, isErrorOnUploadFileToNode: true}

		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, "failed to upload file to node: Something went wrong on uploading file to node")
	})
}

func Test_convertUploadOptions(t *testing.T) {
	t.Parallel()
	mockServiceKey := `no real serviceKey json necessary for these tests`

	t.Run("Use of new serviceKey parameter works", func(t *testing.T) {
		t.Parallel()

		// init
		config := tmsUploadOptions{ServiceKey: mockServiceKey}
		wantOptions := tms.Options{ServiceKey: mockServiceKey, CustomDescription: "Created by Piper"}

		// test
		gotOptions := convertUploadOptions(config)

		// assert
		assert.Equal(t, wantOptions, gotOptions)
	})

	t.Run("Use of old tmsServiceKey parameter works as well", func(t *testing.T) {
		t.Parallel()

		// init
		config := tmsUploadOptions{TmsServiceKey: mockServiceKey}
		wantOptions := tms.Options{ServiceKey: mockServiceKey, CustomDescription: "Created by Piper"}

		// test
		gotOptions := convertUploadOptions(config)

		// assert
		assert.Equal(t, wantOptions, gotOptions)
	})

	t.Run("Use of both tmsServiceKey and serviceKey parameter favors the new serviceKey parameter", func(t *testing.T) {
		t.Parallel()

		// init
		config := tmsUploadOptions{ServiceKey: mockServiceKey, TmsServiceKey: "some other string"}
		wantOptions := tms.Options{ServiceKey: mockServiceKey, CustomDescription: "Created by Piper"}

		// test
		gotOptions := convertUploadOptions(config)

		// assert
		assert.Equal(t, wantOptions, gotOptions)
	})
}
