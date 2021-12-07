package cmd

import (
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/tms"
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
	getNodesResponse                     []tms.Node
	getMtaExtDescriptorResponse          tms.MtaExtDescriptor
	updateMtaExtDescriptorResponse       tms.MtaExtDescriptor
	uploadMtaExtDescriptorToNodeResponse tms.MtaExtDescriptor
	uploadFileResponse                   tms.FileInfo
	uploadFileToNodeResponse             tms.NodeUploadResponseEntity
}

func (cim *communicationInstanceMock) GetNodes() ([]tms.Node, error) {
	return cim.getNodesResponse, nil
}

func (cim *communicationInstanceMock) GetMtaExtDescriptor(nodeId int64, mtaId, mtaVersion string) (tms.MtaExtDescriptor, error) {
	return cim.getMtaExtDescriptorResponse, nil
}

func (cim *communicationInstanceMock) UpdateMtaExtDescriptor(nodeId, idOfMtaExtDescriptor int64, file, mtaVersion, description, namedUser string) (tms.MtaExtDescriptor, error) {
	return cim.updateMtaExtDescriptorResponse, nil
}

func (cim *communicationInstanceMock) UploadMtaExtDescriptorToNode(nodeId int64, file, mtaVersion, description, namedUser string) (tms.MtaExtDescriptor, error) {
	return cim.uploadMtaExtDescriptorToNodeResponse, nil
}

func (cim *communicationInstanceMock) UploadFile(file, namedUser string) (tms.FileInfo, error) {
	return cim.uploadFileResponse, nil
}

func (cim *communicationInstanceMock) UploadFileToNode(nodeName, fileId, description, namedUser string) (tms.NodeUploadResponseEntity, error) {
	return cim.uploadFileToNodeResponse, nil
}

func TestRunTmsUpload(t *testing.T) {
	t.Parallel()

	// TODO: how to declare constants?
	nodeName := "TEST_NODE"
	mtaPathLocal := "example.mtar"
	mtaName := "example.mtar"
	mtaYamlPathLocal := "mta.yaml"
	mtaYamlPath := "./testdata/TestRunTmsUpload/valid/mta.yaml"
	mtaExtDescriptorPathLocal := "test.mtaext"
	mtaExtDescriptorPath := "./testdata/TestRunTmsUpload/valid/test.mtaext"
	customDescription := "This is a test description"
	namedUser := "techUser"
	mtaVersion := "1.0.0"

	t.Run("happy path: 1. get nodes 2. get MTA ext descriptor -> nothing obtained 3. upload MTA ext descriptor 4. upload file 5. upload file to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Name: nodeName, Id: 777}}
		fileInfo := tms.FileInfo{Id: 333, Name: mtaName}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, uploadFileResponse: fileInfo}

		utils := newTmsUploadTestsUtils()
		mtaYamlBytes, _ := os.ReadFile(mtaYamlPath)
		utils.AddFile(mtaYamlPathLocal, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(mtaExtDescriptorPath)
		utils.AddFile(mtaExtDescriptorPathLocal, mtaExtDescriptorBytes)

		utils.AddFile(mtaPathLocal, []byte("dummy content"))

		nodeNameExtDescriptorMapping := map[string]interface{}{nodeName: mtaExtDescriptorPathLocal}
		config := tmsUploadOptions{MtaPath: mtaPathLocal, CustomDescription: customDescription, NamedUser: namedUser, NodeName: nodeName, MtaVersion: mtaVersion, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsUpload(config, &communicationInstance, utils, nil)

		// assert
		assert.NoError(t, err)

		// TODO: does one need to clean the added files?
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		// config := tmsUploadOptions{}

		// utils := newTmsUploadTestsUtils()

		// test
		// err := runTmsUpload(&config, nil, utils)

		// assert
		// assert.EqualError(t, err, "cannot run without important file")
	})
}
