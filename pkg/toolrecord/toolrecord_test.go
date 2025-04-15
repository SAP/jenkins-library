//go:build unit

package toolrecord_test

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/stretchr/testify/assert"
)

func TestToolRecord(t *testing.T) {
	workspace := t.TempDir()

	t.Run("Check toolrecord", func(t *testing.T) {
		fileMock := mock.FilesMock{}
		tr := toolrecord.New(&fileMock, workspace, "dummyTool", "dummyInstance")

		_ = tr.AddKeyData("Organization", "dummyOrgId", "dummyOrgName", "dummyOrgUrl")
		_ = tr.AddKeyData("Project", "dummyProjectId", "dummyProjName", "dummyProjUrl")
		_ = tr.AddKeyData("ScanId", "dummyScanId", "dummyScanName", "dummyScanUrl")
		context := map[string]interface{}{
			"demo": "data",
			"anything": struct {
				s1 string
				i1 int
			}{"goes", 42},
		}
		_ = tr.AddContext("DemoContext", context)
		context2 := "a string"
		_ = tr.AddContext("Context2", context2)
		var context3 [2]string
		context3[0] = "c3_1"
		context3[1] = "c3_2"
		_ = tr.AddContext("Context3", context3)
		err := tr.Persist()
		assert.Nil(t, err, "internal error %s")
		assert.True(t, fileMock.HasFile(tr.GetFileName()), "toolrecord not persisted %s")
	})
}
