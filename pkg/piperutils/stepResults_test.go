package piperutils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersistReportAndLinks(t *testing.T) {
	workspace, err := ioutil.TempDir("", "workspace5")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(workspace)

	reports := []Path{Path{Target: "testFile1.json", Mandatory: true}, Path{Target: "testFile2.json"}}
	links := []Path{Path{Target: "https://1234568.com/test", Name: "Weblink"}}
	PersistReportsAndLinks("checkmarxExecuteScan", workspace, reports, links)

	reportsJSONPath := filepath.Join(workspace, "checkmarxExecuteScan_reports.json")
	reportsFileExists, err := FileExists(reportsJSONPath)
	assert.NoError(t, err, "No error expected but got one")
	assert.Equal(t, true, reportsFileExists, "checkmarxExecuteScan_reports.json missing")

	linksJSONPath := filepath.Join(workspace, "checkmarxExecuteScan_links.json")
	linksFileExists, err := FileExists(linksJSONPath)
	assert.NoError(t, err, "No error expected but got one")
	assert.Equal(t, true, linksFileExists, "checkmarxExecuteScan_links.json missing")

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
}
