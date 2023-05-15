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
	resultMap["DeepLink"] = `https://cx1.sap/projects/f5702f86-b396-417f-82e2-4949a55d5382/scans?branch=master&page=1&id=21e40b36-0dd7-48e5-9768-da1a8f36c907`
	resultMap["Preset"] = `Checkmarx Default`
	resultMap["CheckmarxVersion"] = `v1`
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
	assert.Equal(t, "test-group", reportingData.GroupName)
	assert.Equal(t, "test-group-path", reportingData.GroupPath)
	assert.Equal(t, "checkmarxone", reportingData.ToolName)
	assert.Equal(t, "https://cx1.sap/projects/f5702f86-b396-417f-82e2-4949a55d5382/scans?branch=master&page=1&id=21e40b36-0dd7-48e5-9768-da1a8f36c907", reportingData.DeepLink)
	assert.Equal(t, "Checkmarx Default", reportingData.Preset)
	assert.Equal(t, "v1", reportingData.CheckmarxVersion)
	assert.Equal(t, "Incremental", reportingData.ScanType)

	assert.Equal(t, 10, reportingData.HighTotal)
	assert.Equal(t, 0, reportingData.HighAudited)
	assert.Equal(t, 4, reportingData.MediumTotal)
	assert.Equal(t, 4, reportingData.MediumAudited)
	assert.Equal(t, 2, reportingData.LowTotal)
	assert.Equal(t, 2, reportingData.LowAudited)
	assert.Equal(t, 5, reportingData.InformationTotal)
	assert.Equal(t, 0, reportingData.InformationAudited)
	assert.Equal(t, false, reportingData.IsLowPerQueryAudited)
	assert.Equal(t, 2, len(*reportingData.LowPerQuery))
	if (*reportingData.LowPerQuery)[0].QueryName == "Low_Query_Name_1" {
		assert.Equal(t, "Low_Query_Name_1", (*reportingData.LowPerQuery)[0].QueryName)
		assert.Equal(t, 0, (*reportingData.LowPerQuery)[0].Audited)
		assert.Equal(t, 4, (*reportingData.LowPerQuery)[0].Total)
		assert.Equal(t, "Low_Query_Name_2", (*reportingData.LowPerQuery)[1].QueryName)
		assert.Equal(t, 5, (*reportingData.LowPerQuery)[1].Audited)
		assert.Equal(t, 5, (*reportingData.LowPerQuery)[1].Total)
	} else {
		assert.Equal(t, "Low_Query_Name_1", (*reportingData.LowPerQuery)[1].QueryName)
		assert.Equal(t, 0, (*reportingData.LowPerQuery)[1].Audited)
		assert.Equal(t, 4, (*reportingData.LowPerQuery)[1].Total)
		assert.Equal(t, "Low_Query_Name_2", (*reportingData.LowPerQuery)[0].QueryName)
		assert.Equal(t, 5, (*reportingData.LowPerQuery)[0].Audited)
		assert.Equal(t, 5, (*reportingData.LowPerQuery)[0].Total)
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
	reportingData = CreateJSONHeaderReport(&resultMap)
	assert.Equal(t, true, reportingData.IsLowPerQueryAudited)

	lowPerQuery = map[string]map[string]int{}
	submap = map[string]int{}
	submap["Issues"] = 200
	submap["Confirmed"] = 3
	submap["NotExploitable"] = 2
	lowPerQuery["Low_Query_Name_1"] = submap

	resultMap["LowPerQuery"] = lowPerQuery
	reportingData = CreateJSONHeaderReport(&resultMap)
	assert.Equal(t, false, reportingData.IsLowPerQueryAudited)

	lowPerQuery = map[string]map[string]int{}
	submap = map[string]int{}
	submap["Issues"] = 200
	submap["Confirmed"] = 5
	submap["NotExploitable"] = 5
	lowPerQuery["Low_Query_Name_1"] = submap

	resultMap["LowPerQuery"] = lowPerQuery
	reportingData = CreateJSONHeaderReport(&resultMap)
	assert.Equal(t, true, reportingData.IsLowPerQueryAudited)
}
