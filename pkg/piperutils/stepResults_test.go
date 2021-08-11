package piperutils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/gcs/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistReportAndLinks(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		workspace, err := ioutil.TempDir("", "workspace5")
		require.NoError(t, err, "Failed to create temporary workspace directory")
		// clean up tmp dir
		defer os.RemoveAll(workspace)

		reports := []Path{{Target: "testFile1.json", Mandatory: true}, {Target: "testFile2.json"}}
		links := []Path{{Target: "https://1234568.com/test", Name: "Weblink"}}
		PersistReportsAndLinks("checkmarxExecuteScan", workspace, reports, links, nil, "piper")

		reportsJSONPath := filepath.Join(workspace, "checkmarxExecuteScan_reports.json")
		assert.FileExists(t, reportsJSONPath)

		linksJSONPath := filepath.Join(workspace, "checkmarxExecuteScan_links.json")
		assert.FileExists(t, linksJSONPath)

		var reportsLoaded []Path
		var linksLoaded []Path
		reportsFileData, err := ioutil.ReadFile(reportsJSONPath)
		reportsDataString := string(reportsFileData)
		println(reportsDataString)
		assert.NoError(t, err, "No error expected but got one")

		linksFileData, err := ioutil.ReadFile(linksJSONPath)
		linksDataString := string(linksFileData)
		println(linksDataString)
		assert.NoError(t, err, "No error expected but got one")
		json.Unmarshal(reportsFileData, &reportsLoaded)
		json.Unmarshal(linksFileData, &linksLoaded)

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

	t.Run("empty list", func(t *testing.T) {
		// init
		workspace, err := ioutil.TempDir("", "sonar-")
		require.NoError(t, err, "Failed to create temporary workspace directory")
		// clean up tmp dir
		defer os.RemoveAll(workspace)

		reportsJSONPath := filepath.Join(workspace, "sonarExecuteScan_reports.json")
		linksJSONPath := filepath.Join(workspace, "sonarExecuteScan_links.json")

		// prepare uninitialised parameters
		var reports, links []Path
		require.Empty(t, reports)
		require.Empty(t, links)

		// test
		PersistReportsAndLinks("sonarExecuteScan", workspace, reports, links, nil, "piper")
		// assert
		for _, reportFile := range []string{reportsJSONPath, linksJSONPath} {
			assert.FileExists(t, reportFile)
			reportsFileData, err := ioutil.ReadFile(reportFile)
			require.NoError(t, err, "No error expected but got one")
			assert.Equal(t, "[]", string(reportsFileData))
		}
	})

	t.Run("upload to Google Cloud Storage", func(t *testing.T) {
		workspace, err := ioutil.TempDir("", "workspace5")
		require.NoError(t, err, "Failed to create temporary workspace directory")
		// clean up tmp dir
		defer os.RemoveAll(workspace)

		reports := []Path{{Target: "testFile1.json", Mandatory: true}, {Target: "testFile2.json"}}
		links := []Path{}

		gcsBucketID := "piper"
		mockedGCSClient := &mocks.ClientInterface{}
		for _, report := range reports {
			mockedGCSClient.Mock.On("UploadFile", gcsBucketID, report.Target, report.Target).Return(func(pipelineId string, sourcePath string, targetPath string) error { return nil }).Once()
		}
		PersistReportsAndLinks("checkmarxExecuteScan", workspace, reports, links, mockedGCSClient, gcsBucketID)
		mockedGCSClient.Mock.AssertNumberOfCalls(t, "UploadFile", len(reports))
		mockedGCSClient.Mock.AssertExpectations(t)
	})
}
