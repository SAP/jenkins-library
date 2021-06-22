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

func TestWriteReport(t *testing.T) {
	// init
	const expected = `{"serverUrl":"https://sonarcloud.io","projectKey":"Piper-Validation/Golang","taskId":"mock.Anything","numberOfIssues":{"blocker":0,"critical":1,"major":2,"minor":3,"info":4}}`
	testData := ReportData{
		ServerURL:  "https://sonarcloud.io",
		ProjectKey: "Piper-Validation/Golang",
		TaskID:     mock.Anything,
		NumberOfIssues: Issues{
			Critical: 1,
			Major:    2,
			Minor:    3,
			Info:     4,
		},
	}
	// test
	err := WriteReport(testData, "", writeToFileMock)
	// assert
	assert.NoError(t, err)
	assert.Equal(t, expected, fileContent)
	assert.Equal(t, reportFileName, fileName)
}
