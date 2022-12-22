package cmd

import (
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/tms"
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
		nodeNameExtDescriptorMapStr, convErr := mapToJson(nodeNameExtDescriptorMapping)
		assert.NoError(t, convErr)
		config := tmsUploadOptions{MtaPath: MTA_PATH_LOCAL, CustomDescription: CUSTOM_DESCRIPTION, NamedUser: NAMED_USER, NodeName: NODE_NAME, MtaVersion: MTA_VERSION, NodeExtDescriptorMapping: nodeNameExtDescriptorMapStr}

		// test
		err := runTmsExport(config, &communicationInstance, utils)

		// assert
		assert.NoError(t, err)
	})
}
