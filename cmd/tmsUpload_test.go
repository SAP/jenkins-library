package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/tms"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

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
	if cim.isErrorOnGetMtaExtDescriptor {
		var mtaExtDescriptor tms.MtaExtDescriptor
		return mtaExtDescriptor, errors.New("Something went wrong on getting MTA extension descriptor")
	} else {
		return cim.getMtaExtDescriptorResponse, nil
	}
}

func (cim *communicationInstanceMock) UpdateMtaExtDescriptor(nodeId, idOfMtaExtDescriptor int64, file, mtaVersion, description, namedUser string) (tms.MtaExtDescriptor, error) {
	if cim.isErrorOnUpdateMtaExtDescriptor {
		var mtaExtDescriptor tms.MtaExtDescriptor
		return mtaExtDescriptor, errors.New("Something went wrong on updating MTA extension descriptor")
	} else {
		return cim.updateMtaExtDescriptorResponse, nil
	}
}

func (cim *communicationInstanceMock) UploadMtaExtDescriptorToNode(nodeId int64, file, mtaVersion, description, namedUser string) (tms.MtaExtDescriptor, error) {
	if cim.isErrorOnUploadMtaExtDescriptorToNode {
		var mtaExtDescriptor tms.MtaExtDescriptor
		return mtaExtDescriptor, errors.New("Something went wrong on uploading MTA extension descriptor to node")
	} else {
		return cim.uploadMtaExtDescriptorToNodeResponse, nil
	}
}

func (cim *communicationInstanceMock) UploadFile(file, namedUser string) (tms.FileInfo, error) {
	if cim.isErrorOnUploadFile {
		var fileInfo tms.FileInfo
		return fileInfo, errors.New("Something went wrong on uploading file")
	} else {
		return cim.uploadFileResponse, nil
	}
}

func (cim *communicationInstanceMock) UploadFileToNode(nodeName, fileId, description, namedUser string) (tms.NodeUploadResponseEntity, error) {
	if cim.isErrorOnUploadFileToNode {
		var nodeUploadResponseEntity tms.NodeUploadResponseEntity
		return nodeUploadResponseEntity, errors.New("Something went wrong on uploading file to node")
	} else {
		return cim.uploadFileToNodeResponse, nil
	}
}

func TestRunTmsUpload(t *testing.T) {
	t.Parallel()

	// TODO: how to declare constants?
	nodeName := "TEST_NODE"
	mtaPathLocal := "example.mtar"
	mtaName := "example.mtar"
	mtaId := "com.sap.tms.upload.test"
	mtaExtId := "com.sap.tms.upload.test_ext"
	mtaYamlPathLocal := "mta.yaml"
	mtaYamlPath := "./testdata/TestRunTmsUpload/valid/mta.yaml"
	invalidMtaYamlPath := "./testdata/TestRunTmsUpload/invalid/mta.yaml"
	mtaExtDescriptorPathLocal := "test.mtaext"
	invalidMtaExtDescriptorPathLocal := "wrong_content.mtaext"
	invalidMtaExtDescriptorPathLocal2 := "wrong_extends_parameter.mtaext"
	mtaExtDescriptorPath := "./testdata/TestRunTmsUpload/valid/test.mtaext"
	invalidMtaExtDescriptorPath := "./testdata/TestRunTmsUpload/invalid/wrong_content.mtaext"
	invalidMtaExtDescriptorPath2 := "./testdata/TestRunTmsUpload/invalid/wrong_extends_parameter.mtaext"
	customDescription := "This is a test description"
	namedUser := "techUser"
	mtaVersion := "1.0.0"
	wrongMtaVersion := "3.2.1"
	lastChangedAt := "2021-11-16T13:06:05.711Z"

	t.Run("happy path: 1. get nodes 2. get MTA ext descriptor -> nothing obtained 3. upload MTA ext descriptor to node 4. upload file 5. upload file to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: 777, Name: nodeName}}
		fileInfo := tms.FileInfo{Id: 333, Name: mtaName}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, uploadFileResponse: fileInfo}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(mtaYamlPath)
		utils.AddFile(mtaYamlPathLocal, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(mtaExtDescriptorPath)
		utils.AddFile(mtaExtDescriptorPathLocal, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.NoError(t, err)
	})

	t.Run("happy path: 1. get nodes 2. get MTA ext descriptor 3. update the MTA ext descriptor 4. upload file 5. upload file to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: 777, Name: nodeName}}
		mtaExtDescriptor := tms.MtaExtDescriptor{Id: 456, Description: "Some existing description", MtaId: mtaId, MtaExtId: mtaExtId, MtaVersion: mtaVersion, LastChangedAt: lastChangedAt}
		fileInfo := tms.FileInfo{Id: 333, Name: mtaName}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, getMtaExtDescriptorResponse: mtaExtDescriptor, uploadFileResponse: fileInfo}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(mtaYamlPath)
		utils.AddFile(mtaYamlPathLocal, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(mtaExtDescriptorPath)
		utils.AddFile(mtaExtDescriptorPathLocal, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path: MTA file does not exist", func(t *testing.T) {
		t.Parallel()

		// init
		communicationInstance := communicationInstanceMock{}
		utils := newTmsUploadTestsUtils()

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, fmt.Sprintf("mta file %s not found", mtaPathLocal))
	})

	t.Run("error path: error while getting nodes", func(t *testing.T) {
		t.Parallel()

		// init
		communicationInstance := communicationInstanceMock{isErrorOnGetNodes: true}
		utils := newTmsUploadTestsUtils()
		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to get nodes: Something went wrong on getting nodes")
	})

	t.Run("error path: cannot read mta.yaml (the file is missing)", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: 777, Name: nodeName}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes}
		utils := newTmsUploadTestsUtils()
		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to get mta.yaml as map: could not read 'mta.yaml'")
	})

	t.Run("error path: cannot unmarshal mta.yaml (the file does not represent a yaml)", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: 777, Name: nodeName}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes}
		utils := newTmsUploadTestsUtils()
		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(invalidMtaYamlPath)
		utils.AddFile(mtaYamlPathLocal, mtaYamlBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to get mta.yaml as map: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type map[string]interface {}")
	})

	t.Run("error path: errors on validating the mapping between node names and MTA extension descriptor paths", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: 777, Name: nodeName}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes}
		utils := newTmsUploadTestsUtils()
		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(mtaYamlPath)
		utils.AddFile(mtaYamlPathLocal, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(mtaExtDescriptorPath)
		utils.AddFile(mtaExtDescriptorPathLocal, mtaExtDescriptorBytes)

		invalidMtaExtDescriptorBytes, _ := os.ReadFile(invalidMtaExtDescriptorPath)
		utils.AddFile(invalidMtaExtDescriptorPathLocal, invalidMtaExtDescriptorBytes)

		invalidMtaExtDescriptorBytes2, _ := os.ReadFile(invalidMtaExtDescriptorPath2)
		utils.AddFile(invalidMtaExtDescriptorPathLocal2, invalidMtaExtDescriptorBytes2)

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal, "UNEXISTING_NODE": "unexisting.mtaext", "ONE_MORE_UNEXISTING_NODE": invalidMtaExtDescriptorPathLocal, "ONE_MORE_UNEXISTING_NODE_2": invalidMtaExtDescriptorPathLocal2}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: wrongMtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

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
		nodes := []tms.Node{{Id: 777, Name: nodeName}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, isErrorOnGetMtaExtDescriptor: true}
		utils := newTmsUploadTestsUtils()
		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(mtaYamlPath)
		utils.AddFile(mtaYamlPathLocal, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(mtaExtDescriptorPath)
		utils.AddFile(mtaExtDescriptorPathLocal, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to get MTA extension descriptor: Something went wrong on getting MTA extension descriptor")
	})

	t.Run("error path: error while updating MTA extension descriptor", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: 777, Name: nodeName}}
		mtaExtDescriptor := tms.MtaExtDescriptor{Id: 456, Description: "Some existing description", MtaId: mtaId, MtaExtId: mtaExtId, MtaVersion: mtaVersion, LastChangedAt: lastChangedAt}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, getMtaExtDescriptorResponse: mtaExtDescriptor, isErrorOnUpdateMtaExtDescriptor: true}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(mtaYamlPath)
		utils.AddFile(mtaYamlPathLocal, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(mtaExtDescriptorPath)
		utils.AddFile(mtaExtDescriptorPathLocal, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to update MTA extension descriptor: Something went wrong on updating MTA extension descriptor")
	})

	t.Run("error path: error while uploading MTA extension descriptor to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: 777, Name: nodeName}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, isErrorOnUploadMtaExtDescriptorToNode: true}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(mtaYamlPath)
		utils.AddFile(mtaYamlPathLocal, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(mtaExtDescriptorPath)
		utils.AddFile(mtaExtDescriptorPathLocal, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to upload MTA extension descriptor to node: Something went wrong on uploading MTA extension descriptor to node")
	})

	t.Run("error path: error while uploading file", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: 777, Name: nodeName}}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, isErrorOnUploadFile: true}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(mtaYamlPath)
		utils.AddFile(mtaYamlPathLocal, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(mtaExtDescriptorPath)
		utils.AddFile(mtaExtDescriptorPathLocal, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to upload file: Something went wrong on uploading file")
	})

	t.Run("error path: error while uploading file to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: 777, Name: nodeName}}
		fileInfo := tms.FileInfo{Id: 333, Name: mtaName}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, uploadFileResponse: fileInfo, isErrorOnUploadFileToNode: true}

		utils := newTmsUploadTestsUtils()
		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(mtaYamlPath)
		utils.AddFile(mtaYamlPathLocal, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(mtaExtDescriptorPath)
		utils.AddFile(mtaExtDescriptorPathLocal, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.EqualError(t, err, "failed to upload file to node: Something went wrong on uploading file to node")
	})
}
