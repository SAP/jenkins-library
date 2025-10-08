package sonar

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadTaskReport(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// test
		result, err := ReadTaskReport("./testData/valid")
		// assert
		assert.Equal(t, "piper-test", result.ProjectKey)
		assert.Equal(t, "AXERR2JBbm9IiM5TEST", result.TaskID)
		assert.Equal(t, "https://sonarcloud.io/api/ce/task?id=AXERR2JBbm9IiMTEST", result.TaskURL)
		assert.Equal(t, "https://sonarcloud.io/dashboard/index/piper-test", result.DashboardURL)
		assert.Equal(t, "https://sonarcloud.io", result.ServerURL)
		assert.Equal(t, "8.0.0.12345", result.ServerVersion)
		assert.NoError(t, err)
	})

	t.Run("missing file", func(t *testing.T) {
		// test
		result, err := ReadTaskReport("./testData/missing")
		// assert
		assert.Empty(t, result.ProjectKey)
		assert.True(t,
			(filepath.FromSlash("open testData/missing/.scannerwork/report-task.txt: The system cannot find the path specified.") == err.Error()) ||
				(filepath.FromSlash("open testData/missing/.scannerwork/report-task.txt: no such file or directory") == err.Error()))
	})

	t.Run("invalid file", func(t *testing.T) {
		// test
		result, err := ReadTaskReport("./testData/invalid")
		// assert
		assert.Empty(t, result.ProjectKey)
		assert.EqualError(t, err, filepath.FromSlash("decode testData/invalid/.scannerwork/report-task.txt: missing required key projectKey"))
	})
}
