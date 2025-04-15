//go:build unit

package cmd

import (
	"os"
	"strconv"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/tms"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type tmsExportMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTmsExportTestsUtils() tmsExportMockUtils {
	utils := tmsExportMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func (cim *communicationInstanceMock) ExportFileToNode(fileInfo tms.FileInfo, nodeName, description, namedUser string) (tms.NodeUploadResponseEntity, error) {
	fileId := strconv.FormatInt(fileInfo.Id, 10)
	var nodeUploadResponseEntity tms.NodeUploadResponseEntity
	if description != CUSTOM_DESCRIPTION || nodeName != NODE_NAME || fileId != strconv.FormatInt(FILE_ID, 10) || namedUser != NAMED_USER {
		return nodeUploadResponseEntity, errors.New(INVALID_INPUT_MSG)
	}

	if cim.isErrorOnExportFileToNode {
		return nodeUploadResponseEntity, errors.New("Something went wrong on exporting file to node")
	} else {
		return cim.exportFileToNodeResponse, nil
	}
}

func TestRunTmsExport(t *testing.T) {
	t.Parallel()

	t.Run("happy path: 1. get nodes 2. get MTA ext descriptor -> nothing obtained 3. upload MTA ext descriptor to node 4. upload file 5. export file to node", func(t *testing.T) {
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
		config := tmsExportOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsExport(config, &communicationInstance, utils)

		// assert
		assert.NoError(t, err)
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
		config := tmsExportOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsExport(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, "failed to upload file: Something went wrong on uploading file")
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
		config := tmsExportOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsExport(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, "failed to upload MTA extension descriptor to node: Something went wrong on uploading MTA extension descriptor to node")
	})

	t.Run("error path: error while exporting file to node", func(t *testing.T) {
		t.Parallel()

		// init
		nodes := []tms.Node{{Id: NODE_ID, Name: NODE_NAME}}
		fileInfo := tms.FileInfo{Id: FILE_ID, Name: MTA_NAME}
		communicationInstance := communicationInstanceMock{getNodesResponse: nodes, uploadFileResponse: fileInfo, isErrorOnExportFileToNode: true}

		utils := newTmsTestsUtils()
		utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

		mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
		utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

		mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
		utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

		nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}
		config := tmsExportOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapping}

		// test
		err := runTmsExport(config, &communicationInstance, utils)

		// assert
		assert.EqualError(t, err, "failed to export file to node: Something went wrong on exporting file to node")
	})
}

func Test_convertExportOptions(t *testing.T) {
	t.Parallel()
	mockServiceKey := `no real serviceKey json necessary for these tests`

	t.Run("Use of new serviceKey parameter works", func(t *testing.T) {
		t.Parallel()

		// init
		config := tmsExportOptions{ServiceKey: mockServiceKey}
		wantOptions := tms.Options{ServiceKey: mockServiceKey, CustomDescription: "Created by Piper"}

		// test
		gotOptions := convertExportOptions(config)

		// assert
		assert.Equal(t, wantOptions, gotOptions)
	})

	t.Run("Use of old tmsServiceKey parameter works as well", func(t *testing.T) {
		t.Parallel()

		// init
		config := tmsExportOptions{TmsServiceKey: mockServiceKey}
		wantOptions := tms.Options{ServiceKey: mockServiceKey, CustomDescription: "Created by Piper"}

		// test
		gotOptions := convertExportOptions(config)

		// assert
		assert.Equal(t, wantOptions, gotOptions)
	})

	t.Run("Use of both tmsServiceKey and serviceKey parameter favors the new serviceKey parameter", func(t *testing.T) {
		t.Parallel()

		// init
		config := tmsExportOptions{ServiceKey: mockServiceKey, TmsServiceKey: "some other string"}
		wantOptions := tms.Options{ServiceKey: mockServiceKey, CustomDescription: "Created by Piper"}

		// test
		gotOptions := convertExportOptions(config)

		// assert
		assert.Equal(t, wantOptions, gotOptions)
	})
}
