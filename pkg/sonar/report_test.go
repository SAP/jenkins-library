//go:build unit
// +build unit

package sonar

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var fileContent string
var fileName string

func writeToFileMock(f string, d []byte, p os.FileMode) error {
	fileContent = string(d)
	fileName = f
	return nil
}

func TestWriteCodeCheckReport(t *testing.T) {
	// init
	const expected = `{"serverUrl":"https://sonarcloud.io","projectKey":"Piper-Validation/Golang","taskId":"mock.Anything","numberOfIssues":{"blocker":0,"critical":1,"major":2,"minor":3,"info":4},"scanResults":[{"severity":"CRITICAL","error_type":"CODE_SMELL","issues":10}],"coverage":{"coverage":13.7,"lineCoverage":37.1,"linesToCover":123,"uncoveredLines":23,"branchCoverage":42,"branchesToCover":30,"uncoveredBranches":3},"linesOfCode":{"total":327,"languageDistribution":[{"languageKey":"java","linesOfCode":327}]}}`
	testData := ReportCodeCheckData{
		ServerURL:  "https://sonarcloud.io",
		ProjectKey: "Piper-Validation/Golang",
		TaskID:     mock.Anything,
		ScanResults: []Severity{
			{
				SeverityType: "CRITICAL",
				IssueType:    "CODE_SMELL",
				IssueCount:   10,
			},
		},
		NumberOfIssues: &Issues{
			Critical: 1,
			Major:    2,
			Minor:    3,
			Info:     4,
		},
		Coverage: &SonarCoverage{
			Coverage:          13.7,
			BranchCoverage:    42,
			LineCoverage:      37.1,
			LinesToCover:      123,
			UncoveredLines:    23,
			BranchesToCover:   30,
			UncoveredBranches: 3,
		},
		LinesOfCode: &SonarLinesOfCode{
			Total:                327,
			LanguageDistribution: []SonarLanguageDistribution{{LanguageKey: "java", LinesOfCode: 327}},
		},
	}
	// test
	err := WriteCodeCheckReport(testData, "", writeToFileMock)
	// assert
	assert.NoError(t, err)
	assert.Equal(t, expected, fileContent)
	assert.Equal(t, reportCodeCheckFileName, fileName)
}

func TestWriteHotSpotReport(t *testing.T) {
	// init
	const expected = `{"serverUrl":"https://sonarcloud.io","projectKey":"Piper-Validation/Golang","taskId":"mock.Anything","securityHotspots":[{"priority":"HIGH","hotspots":1},{"priority":"LOW","hotspots":4}]}`
	testData := ReportHotSpotData{
		ServerURL:  "https://sonarcloud.io",
		ProjectKey: "Piper-Validation/Golang",
		TaskID:     mock.Anything,
		SecurityHotspots: []SecurityHotspot{
			{Priority: "HIGH", Hotspots: 1},
			{Priority: "LOW", Hotspots: 4},
		},
	}
	// test
	err := WriteHotSpotReport(testData, "", writeToFileMock)
	// assert
	assert.NoError(t, err)
	assert.Equal(t, expected, fileContent)
	assert.Equal(t, reportHotSpotFileName, fileName)
}

func TestWriteCombinedReport(t *testing.T) {
	// init
	const expected = `{"numberOfIssues":{"blocker":0,"critical":1,"major":2,"minor":3,"info":4},"scanResults":[{"severity":"CRITICAL","error_type":"CODE_SMELL","issues":10}],"securityHotspots":[{"priority":"HIGH","hotspots":1},{"priority":"LOW","hotspots":4}]}`
	combinedData := ReportCombinedData{
		ScanResults: []Severity{
			{
				SeverityType: "CRITICAL",
				IssueType:    "CODE_SMELL",
				IssueCount:   10,
			},
		},
		NumberOfIssues: &Issues{
			Critical: 1,
			Major:    2,
			Minor:    3,
			Info:     4,
		},
		SecurityHotspots: []SecurityHotspot{
			{Priority: "HIGH", Hotspots: 1},
			{Priority: "LOW", Hotspots: 4},
		},
	}
	// test
	err := WriteCombinedReport(combinedData, "", writeToFileMock)
	// assert
	assert.NoError(t, err)
	assert.Equal(t, expected, fileContent)
	assert.Equal(t, reportCombinedFileName, fileName)
}
