package tms

import (
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

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

const NODE_ID = 777
const NODE_NAME = "TEST_NODE"
const MTA_PATH_LOCAL = "example.mtar"
const MTA_YAML_PATH = "../../cmd/testdata/TestRunTmsUpload/valid/mta.yaml"
const MTA_YAML_PATH_LOCAL = "mta.yaml"
const MTA_EXT_DESCRIPTOR_PATH = "../../cmd/testdata/TestRunTmsUpload/valid/test.mtaext"
const MTA_EXT_DESCRIPTOR_PATH_LOCAL = "test.mtaext"
const MTA_VERSION = "1.0.0"

func TestFormNodeIdExtDescriptorMappingWithValidation2(t *testing.T) {
	utils := newTmsTestsUtils()
	nodes := []Node{{Id: NODE_ID, Name: NODE_NAME}}

	utils.AddFile(MTA_PATH_LOCAL, []byte("dummy content"))

	mtaYamlBytes, _ := os.ReadFile(MTA_YAML_PATH)
	utils.AddFile(MTA_YAML_PATH_LOCAL, mtaYamlBytes)

	mtaExtDescriptorBytes, _ := os.ReadFile(MTA_EXT_DESCRIPTOR_PATH)
	utils.AddFile(MTA_EXT_DESCRIPTOR_PATH_LOCAL, mtaExtDescriptorBytes)

	nodeNameExtDescriptorMapping := map[string]interface{}{NODE_NAME: MTA_EXT_DESCRIPTOR_PATH_LOCAL}

	mtaYamlMap, errGetMtaYamlAsMap := GetYamlAsMap(utils, MTA_YAML_PATH_LOCAL)
	assert.Nil(t, errGetMtaYamlAsMap)

	nodeIdExtDescriptorMapping, err := FormNodeIdExtDescriptorMappingWithValidation(utils, nodeNameExtDescriptorMapping, nodes, mtaYamlMap, MTA_VERSION)
	assert.NoError(t, err)
	assert.Equal(t, map[int64]string(map[int64]string{NODE_ID: MTA_EXT_DESCRIPTOR_PATH_LOCAL}), nodeIdExtDescriptorMapping)
}
