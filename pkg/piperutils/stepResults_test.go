package piperutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fileMock struct {
	fileMap     map[string][]byte
	writeErrors map[string]error
}

func (f *fileMock) WriteFile(filename string, data []byte, perm os.FileMode) error {
	if f.writeErrors != nil && f.writeErrors[filename] != nil {
		return f.writeErrors[filename]
	}
	f.fileMap[filename] = data
	return nil
}

func (f *fileMock) ReadFile(name string) ([]byte, error) {
	return f.fileMap[name], nil
}

func (f *fileMock) FileExists(name string) bool {
	return f.fileMap[name] != nil
}

func TestPersistReportAndLinks(t *testing.T) {
	workspace := ""
	t.Run("success - default", func(t *testing.T) {
		files := fileMock{fileMap: map[string][]byte{}}

		reports := []Path{{Target: "testFile1.json", Mandatory: true}, {Target: "testFile2.json"}}
		links := []Path{{Target: "https://1234568.com/test", Name: "Weblink"}}
		err := PersistReportsAndLinks("checkmarxExecuteScan", workspace, &files, reports, links)
		assert.NoError(t, err)

		reportsJSONPath := filepath.Join(workspace, "checkmarxExecuteScan_reports.json")
		assert.True(t, files.FileExists(reportsJSONPath))

		linksJSONPath := filepath.Join(workspace, "checkmarxExecuteScan_links.json")
		assert.True(t, files.FileExists(linksJSONPath))

		var reportsLoaded []Path
		var linksLoaded []Path
		reportsFileData, err := files.ReadFile(reportsJSONPath)
		reportsDataString := string(reportsFileData)
		println(reportsDataString)
		assert.NoError(t, err, "No error expected but got one")

		linksFileData, err := files.ReadFile(linksJSONPath)
		linksDataString := string(linksFileData)
		println(linksDataString)
		assert.NoError(t, err, "No error expected but got one")
		_ = json.Unmarshal(reportsFileData, &reportsLoaded)
		_ = json.Unmarshal(linksFileData, &linksLoaded)

		assert.Equal(t, 2, len(reportsLoaded), "wrong number of reports")
		assert.Equal(t, 1, len(linksLoaded), "wrong number of links")
		assert.Equal(t, true, reportsLoaded[0].Mandatory, "mandatory flag on report 1 has wrong value")
		assert.Equal(t, "testFile1.json", reportsLoaded[0].Target, "target value on report 1 has wrong value")
		assert.Equal(t, false, reportsLoaded[1].Mandatory, "mandatory flag on report 2 has wrong value")
		assert.Equal(t, "testFile2.json", reportsLoaded[1].Target, "target value on report 1 has wrong value")
		assert.Equal(t, false, linksLoaded[0].Mandatory, "mandatory flag on link 1 has wrong value")
		assert.Equal(t, "https://1234568.com/test", linksLoaded[0].Target, "target value on link 1 has wrong value")
		assert.Equal(t, "Weblink", linksLoaded[0].Name, "name value on link 1 has wrong value")
	})

	t.Run("success - empty list", func(t *testing.T) {
		files := fileMock{fileMap: map[string][]byte{}}

		reportsJSONPath := filepath.Join(workspace, "sonarExecuteScan_reports.json")
		linksJSONPath := filepath.Join(workspace, "sonarExecuteScan_links.json")

		// prepare uninitialised parameters
		var reports, links []Path
		require.Empty(t, reports)
		require.Empty(t, links)

		// test
		err := PersistReportsAndLinks("sonarExecuteScan", workspace, &files, reports, links)
		// assert
		assert.NoError(t, err)
		for _, reportFile := range []string{reportsJSONPath, linksJSONPath} {
			assert.True(t, files.FileExists(reportFile))
			reportsFileData, err := files.ReadFile(reportFile)
			require.NoError(t, err, "No error expected but got one")
			assert.Equal(t, "[]", string(reportsFileData))
		}
	})

	t.Run("failure - write reports", func(t *testing.T) {
		stepName := "checkmarxExecuteScan"
		files := fileMock{
			fileMap:     map[string][]byte{},
			writeErrors: map[string]error{filepath.Join(workspace, fmt.Sprintf("%v_reports.json", stepName)): fmt.Errorf("write error")},
		}

		reports := []Path{{Target: "testFile1.json"}, {Target: "testFile2.json"}}
		links := []Path{{Target: "https://1234568.com/test", Name: "Weblink"}}
		err := PersistReportsAndLinks(stepName, workspace, &files, reports, links)

		assert.EqualError(t, err, "failed to write reports.json: write error")
	})

	t.Run("failure - write links", func(t *testing.T) {
		stepName := "checkmarxExecuteScan"
		files := fileMock{
			fileMap:     map[string][]byte{},
			writeErrors: map[string]error{filepath.Join(workspace, fmt.Sprintf("%v_links.json", stepName)): fmt.Errorf("write error")},
		}

		reports := []Path{{Target: "testFile1.json"}, {Target: "testFile2.json"}}
		links := []Path{{Target: "https://1234568.com/test", Name: "Weblink"}}
		err := PersistReportsAndLinks(stepName, workspace, &files, reports, links)

		assert.EqualError(t, err, "failed to write links.json: write error")
	})
}
