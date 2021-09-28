package toolrecord_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/stretchr/testify/assert"
)

func TestToolRecord(t *testing.T) {
	workspace, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	defer os.RemoveAll(workspace)

	t.Run("Check toolrecord", func(t *testing.T) {
		tr := toolrecord.New(workspace, "dummyTool", "dummyInstance")

		tr.AddKeyData("Organization", "dummyOrgId", "dummyOrgName", "dummyOrgUrl")
		tr.AddKeyData("Project", "dummyProjectId", "dummyProjName", "dummyProjUrl")
		tr.AddKeyData("ScanId", "dummyScanId", "dummyScanName", "dummyScanUrl")
		context := map[string]interface{}{
			"demo": "data",
			"anything": struct {
				s1 string
				i1 int
			}{"goes", 42},
		}
		tr.AddContext("DemoContext", context)
		context2 := "a string"
		tr.AddContext("Context2", context2)
		var context3 [2]string
		context3[0] = "c3_1"
		context3[1] = "c3_2"
		tr.AddContext("Context3", context3)
		err := tr.Persist()
		assert.Nil(t, err, "internal error %s")
		assert.FileExists(t, tr.GetFileName(), "toolrecord not persisted %s")
	})
}
