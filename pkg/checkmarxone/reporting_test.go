package checkmarxOne

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSASTSarifRepositoryPaths(t *testing.T) {
	for _, directory := range []string{".", "componentA", "components/componentA"} {
		t.Run(directory, func(t *testing.T) {
			root := t.TempDir()
			_, err := git.PlainInit(root, false)
			require.NoError(t, err)
			workingDirectory := filepath.Join(root, directory)
			require.NoError(t, os.MkdirAll(workingDirectory, 0o700))
			t.Chdir(root)

			var results []ScanResult
			require.NoError(t, json.Unmarshal([]byte(`[{"type":"sast","severity":"HIGH","data":{"queryId":1,"queryName":"Test_Query","languageName":"Go","nodes":[{"fileName":"/main.go","line":1,"column":1},{"fileName":"/src/sink.go","line":2,"column":1}]}}]`), &results))
			sarif, err := ConvertCxSASTJSONToSarif(nil, "https://example.com", &results, &Scan{}, workingDirectory)
			require.NoError(t, err)
			require.Len(t, sarif.Runs[0].Results, 1)
			result := sarif.Runs[0].Results[0]
			assert.Equal(t, filepath.ToSlash(filepath.Join(directory, "main.go")), result.Locations[0].PhysicalLocation.ArtifactLocation.URI)
			for index, filename := range []string{"main.go", "src/sink.go"} {
				expected := filepath.ToSlash(filepath.Join(directory, filename))
				assert.Equal(t, expected, result.RelatedLocations[index].PhysicalLocation.ArtifactLocation.URI)
				assert.Equal(t, expected, result.CodeFlows[0].ThreadFlows[0].Locations[index].Location.PhysicalLocation.ArtifactLocation.URI)
			}
			t.Chdir(workingDirectory)
			workingDirectorySarif, err := ConvertCxSASTJSONToSarif(nil, "https://example.com", &results, &Scan{})
			require.NoError(t, err)
			assert.Equal(t, sarif.Runs[0].Results, workingDirectorySarif.Runs[0].Results)
		})
	}
}

func TestSarifSourcePathOutsideRepository(t *testing.T) {
	t.Chdir(t.TempDir())
	assert.Empty(t, sarifSourcePrefix())
	assert.Equal(t, "main.go", sarifSourcePath(sarifSourcePrefix(), "/main.go"))
	assert.Equal(t, "componentA/main.go", sarifSourcePath("componentA", "main.go"))
	assert.Equal(t, "componentA/main.go", sarifSourcePath("componentA", "//main.go"))
}

type sarifIACSystem struct {
	System
}

func (sarifIACSystem) GetIACFindingInfo(ScanResult) (IACFindingInfo, error) {
	return IACFindingInfo{}, nil
}

func TestIACSarifRepositoryPaths(t *testing.T) {
	root := t.TempDir()
	_, err := git.PlainInit(root, false)
	require.NoError(t, err)
	componentDirectory := filepath.Join(root, "componentA")
	require.NoError(t, os.MkdirAll(componentDirectory, 0o700))
	t.Chdir(root)

	var results []ScanResult
	require.NoError(t, json.Unmarshal([]byte(`[{"type":"kics","severity":"HIGH","data":{"queryId":"test-query","queryName":"Test_Query","platform":"Terraform","fileName":"/main.tf","line":1}}]`), &results))
	sarif, err := ConvertCxIACJSONToSarif(sarifIACSystem{}, "https://example.com", &results, &Scan{}, componentDirectory)
	require.NoError(t, err)
	require.Len(t, sarif.Runs[0].Results, 1)
	result := sarif.Runs[0].Results[0]
	assert.Equal(t, "componentA/main.tf", result.Locations[0].PhysicalLocation.ArtifactLocation.URI)
	assert.Equal(t, "componentA/main.tf", result.RelatedLocations[0].PhysicalLocation.ArtifactLocation.URI)
	assert.Equal(t, "componentA/main.tf", result.CodeFlows[0].ThreadFlows[0].Locations[0].Location.PhysicalLocation.ArtifactLocation.URI)
}

func TestCreateJSONReport(t *testing.T) {
	resultMap := map[string]any{}
	resultMap["ToolName"] = `checkmarxone`
	resultMap["ProjectName"] = `ssba`
	resultMap["Group"] = `test-group`
	resultMap["GroupFullPathOnReportDate"] = `test-group-path`
	resultMap["Application"] = `test-app`
	resultMap["ApplicationFullPathOnReportDate"] = `test-app-path`
	resultMap["DeepLink"] = `https://cx1.sap/projects/f5702f86-b396-417f-82e2-4949a55d5382/scans?branch=master&page=1&id=21e40b36-0dd7-48e5-9768-da1a8f36c907`
	resultMap["SastPreset"] = `Checkmarx Default`
	resultMap["ToolVersion"] = `v1`
	resultMap["SASTVersion"] = `SAST: 9.7.6`
	resultMap["ScanType"] = `Incremental`
	resultMap["ProjectId"] = `f5702f86-b396-417f-82e2-4949a55d5382`
	resultMap["ScanId"] = `21e40b36-0dd7-48e5-9768-da1a8f36c907`

	resultMap["Critical"] = map[string]int{}
	resultMap["High"] = map[string]int{}
	resultMap["Medium"] = map[string]int{}
	resultMap["Low"] = map[string]int{}
	resultMap["Information"] = map[string]int{}

	submap := map[string]int{}
	submap["Issues"] = 10
	submap["NotFalsePositive"] = 10
	resultMap["Critical"] = submap

	submap = map[string]int{}
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

	reportingData := CreateJSONHeaderReport(&resultMap, "sast")
	assert.Equal(t, "21e40b36-0dd7-48e5-9768-da1a8f36c907", reportingData.ScanID)
	assert.Equal(t, "ssba", reportingData.ProjectName)
	assert.Equal(t, "f5702f86-b396-417f-82e2-4949a55d5382", reportingData.ProjectID)
	assert.Equal(t, "test-group", reportingData.GroupID)
	assert.Equal(t, "test-group-path", reportingData.GroupName)
	assert.Equal(t, "CheckmarxOne", reportingData.ToolName)
	assert.Equal(t, "https://cx1.sap/projects/f5702f86-b396-417f-82e2-4949a55d5382/scans?branch=master&page=1&id=21e40b36-0dd7-48e5-9768-da1a8f36c907", reportingData.DeepLink)
	assert.Equal(t, "Checkmarx Default", reportingData.Preset)
	assert.Equal(t, "v1, SAST: 9.7.6", reportingData.ToolVersion)
	assert.Equal(t, "Incremental", reportingData.ScanType)

	lowList := (*reportingData.Findings)[3].LowPerQuery
	lowListLen := len(*lowList)
	assert.Equal(t, 2, lowListLen)

	for i := range lowListLen {
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
