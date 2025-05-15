package checkmarxOne

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateJSONReport(t *testing.T) {
	resultMap := map[string]interface{}{}
	resultMap["ToolName"] = `checkmarxone`
	resultMap["ProjectName"] = `ssba`
	resultMap["Group"] = `test-group`
	resultMap["GroupFullPathOnReportDate"] = `test-group-path`
	resultMap["Application"] = `test-app`
	resultMap["ApplicationFullPathOnReportDate"] = `test-app-path`
	resultMap["DeepLink"] = `https://cx1.sap/projects/f5702f86-b396-417f-82e2-4949a55d5382/scans?branch=master&page=1&id=21e40b36-0dd7-48e5-9768-da1a8f36c907`
	resultMap["Preset"] = `Checkmarx Default`
	resultMap["ToolVersion"] = `v1`
	resultMap["ScanType"] = `Incremental`
	resultMap["ProjectId"] = `f5702f86-b396-417f-82e2-4949a55d5382`
	resultMap["ScanId"] = `21e40b36-0dd7-48e5-9768-da1a8f36c907`

	resultMap["High"] = map[string]int{}
	resultMap["Medium"] = map[string]int{}
	resultMap["Low"] = map[string]int{}
	resultMap["Information"] = map[string]int{}
	submap := map[string]int{}
	submap["Issues"] = 10
	submap["NotFalsePositive"] = 10
	resultMap["High"] = submap

	submap = map[string]int{}
	submap["Issues"] = 4
	submap["NotFalsePositive"] = 0
	resultMap["Medium"] = submap

	submap = map[string]int{}
	submap["Issues"] = 2
	submap["NotFalsePositive"] = 2
	submap["Confirmed"] = 1
	submap["NotExploitable"] = 1
	resultMap["Low"] = submap

	submap = map[string]int{}
	submap["Issues"] = 5
	submap["NotFalsePositive"] = 5
	resultMap["Information"] = submap

	lowPerQuery := map[string]map[string]int{}
	submap = map[string]int{}
	submap["Issues"] = 4
	submap["Confirmed"] = 0
	submap["NotExploitable"] = 0
	lowPerQuery["Low_Query_Name_1"] = submap

	submap = map[string]int{}
	submap["Issues"] = 5
	submap["Confirmed"] = 2
	submap["NotExploitable"] = 3
	lowPerQuery["Low_Query_Name_2"] = submap

	resultMap["LowPerQuery"] = lowPerQuery

	reportingData := CreateJSONHeaderReport(&resultMap)
	assert.Equal(t, "21e40b36-0dd7-48e5-9768-da1a8f36c907", reportingData.ScanID)
	assert.Equal(t, "ssba", reportingData.ProjectName)
	assert.Equal(t, "f5702f86-b396-417f-82e2-4949a55d5382", reportingData.ProjectID)
	assert.Equal(t, "test-group", reportingData.GroupID)
	assert.Equal(t, "test-group-path", reportingData.GroupName)
	assert.Equal(t, "CheckmarxOne", reportingData.ToolName)
	assert.Equal(t, "https://cx1.sap/projects/f5702f86-b396-417f-82e2-4949a55d5382/scans?branch=master&page=1&id=21e40b36-0dd7-48e5-9768-da1a8f36c907", reportingData.DeepLink)
	assert.Equal(t, "Checkmarx Default", reportingData.Preset)
	assert.Equal(t, "v1", reportingData.ToolVersion)
	assert.Equal(t, "Incremental", reportingData.ScanType)

	lowList := (*reportingData.Findings)[2].LowPerQuery
	lowListLen := len(*lowList)
	assert.Equal(t, 2, lowListLen)

	for i := 0; i < lowListLen; i++ {
		if (*lowList)[i].QueryName == "Low_Query_Name_1" {
			assert.Equal(t, 0, (*lowList)[i].Audited)
			assert.Equal(t, 4, (*lowList)[i].Total)
		}

		if (*lowList)[i].QueryName == "Low_Query_Name_2" {
			assert.Equal(t, 5, (*lowList)[i].Audited)
			assert.Equal(t, 5, (*lowList)[i].Total)
		}
	}

	lowPerQuery = map[string]map[string]int{}
	submap = map[string]int{}
	submap["Issues"] = 100
	submap["Confirmed"] = 10
	submap["NotExploitable"] = 0
	lowPerQuery["Low_Query_Name_1"] = submap

	submap = map[string]int{}
	submap["Issues"] = 5
	submap["Confirmed"] = 2
	submap["NotExploitable"] = 3
	lowPerQuery["Low_Query_Name_2"] = submap

	resultMap["LowPerQuery"] = lowPerQuery
}
