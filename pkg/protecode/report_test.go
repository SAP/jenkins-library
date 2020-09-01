package protecode

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fileContent string

func writeToFileMock(f string, d []byte, p os.FileMode) error {
	fileContent = string(d)
	return nil
}

func TestWriteReportDataToJSONFile(t *testing.T) {
	expected := "{\"target\":\"REPORTFILENAME\",\"mandatory\":true,\"productID\":\"4711\",\"serverUrl\":\"DUMMYURL\",\"count\":\"0\",\"cvss2GreaterOrEqualSeven\":\"4\",\"cvss3GreaterOrEqualSeven\":\"3\",\"excludedVulnerabilities\":\"2\",\"triagedVulnerabilities\":\"0\",\"historicalVulnerabilities\":\"1\",\"Vulnerabilities\":[{\"cve\":\"Vulnerability\",\"cvss\":2.5,\"cvss3_score\":\"5.5\"}]}"

	var parsedResult map[string]int = make(map[string]int)
	parsedResult["historical_vulnerabilities"] = 1
	parsedResult["excluded_vulnerabilities"] = 2
	parsedResult["cvss3GreaterOrEqualSeven"] = 3
	parsedResult["cvss2GreaterOrEqualSeven"] = 4
	parsedResult["vulnerabilities"] = 5

	WriteReport("DUMMYURL", false, "", "REPORTFILENAME", ".", "", parsedResult, 4711, []Vuln{{"Vulnerability", 2.5, "5.5"}}, writeToFileMock)
	assert.Equal(t, fileContent, expected, "content should be not empty")
}
