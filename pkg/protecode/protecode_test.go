package protecode

import (
	"testing"

	"fmt"
	"os"
	"bytes"
	"io/ioutil"

	"github.com/stretchr/testify/assert"
)

func TestParseProteCodeResultSuccess(t *testing.T) {

	var result ProteCodeResult = ProteCodeResult{
		ProductId: "ProductId",
		ReportUrl: "ReportUrl",
		Status:    "B",
		Components: []ProteCodeComponent{
			{Vulns: []ProteCodeVulnerability{
				{Exact: true, Triage: "", Vuln: ProteCodeVuln{Cve: "Cve1", Cvss: 7.2, Cvss3Score: "0.0"}},
				{Exact: true, Triage: "triage2", Vuln: ProteCodeVuln{Cve: "Cve2", Cvss: 2.2, Cvss3Score: "2.3"}},
				{Exact: true, Triage: "", Vuln: ProteCodeVuln{Cve: "Cve2b", Cvss: 0.0, Cvss3Score: "0.0"}},
			},
			},
			{Vulns: []ProteCodeVulnerability{
				{Exact: true, Triage: "", Vuln: ProteCodeVuln{Cve: "Cve3", Cvss: 3.2, Cvss3Score: "7.3"}},
				{Exact: true, Triage: "", Vuln: ProteCodeVuln{Cve: "Cve4", Cvss: 8.0, Cvss3Score: "8.0"}},
				{Exact: false, Triage: "", Vuln: ProteCodeVuln{Cve: "Cve4b", Cvss: 8.0, Cvss3Score: "8.0"}},
			},
			},
		},
	}
	m := ParseResultToInflux(result, "Excluded CVES: Cve4,")
	t.Run("Parse Protecode Results", func(t *testing.T) {
		assert.Equal(t, 1, m["historical_vulnerabilities"])
		assert.Equal(t, 1, m["triaged_vulnerabilities"])
		assert.Equal(t, 1, m["excluded_vulnerabilities"])
		assert.Equal(t, 1, m["minor_vulnerabilities"])
		assert.Equal(t, 2, m["major_vulnerabilities"])
		assert.Equal(t, 3, m["vulnerabilities"])
	})
}

//func TestCmdExecGetProtecodeResultSuccess(t *testing.T) {
//
//	cases := []struct {
//		cmdName   string
//		cmdString string
//		want      ProteCodeResult
//	}{
//		{"echo", "test", ProteCodeResult{ProductId: "productID2"}},
//		{"echo", "Dummy-DeLiMiTeR-status=200", ProteCodeResult{ProductId: "productID1"}},
//	}
//	for _, c := range cases {
//
//		got := CmdExecGetProtecodeResult(c.cmdName, c.cmdString)
//		assert.Equal(t, c.want, got)
//	}
//}

func TestCreateRequestHeader(t *testing.T) {

	cases := []struct {
		verbose bool
		auth    string
		hMap    map[string][]string
		want    map[string][]string
	}{
		{true, "auth1",
			map[string][]string{
				"test": []string{"dummy1"}},
			map[string][]string{
				"test":                   []string{"dummy1"},
				"authentication":         []string{"Basic auth1"},
				"quiet":                  []string{"false"},
				"ignoreSslErrors":        []string{"true"},
				"consoleLogResponseBody": []string{"true"},
			}},
	}

	for _, c := range cases {

		got := CreateRequestHeader(c.verbose, c.auth, c.hMap)
		assert.Equal(t, c.want, got)
	}
}

func TestGetProteCodeResultData(t *testing.T) {

	cases := []struct {
		give string
		want ProteCodeResultData
	}{
		{`{"results": {"product_id": "ID1"}}`, ProteCodeResultData{Result: ProteCodeResult{ProductId: "ID1"}}},
	}

	for _, c := range cases {

		r := ioutil.NopCloser(bytes.NewReader([]byte(c.give)))
		got := GetProteCodeResultData(r)
		assert.Equal(t, c.want, *got)
	}
}

func TestGetProteCodeProductData(t *testing.T) {

	cases := []struct {
		give string
		want ProteCodeProductData
	}{
		{`{"products": [{"product_id": "ID1"}]}`, ProteCodeProductData{Products: []ProteCodeProduct{{ProductId: "ID1"}}}},
	}

	for _, c := range cases {

		r := ioutil.NopCloser(bytes.NewReader([]byte(c.give)))
		got := GetProteCodeProductData(r)
		assert.Equal(t, c.want, *got)
	}
}

var fileWriterContent []byte

func fileWriterMock(fileName string, b []byte, perm os.FileMode) error {

	switch fileName {
	case "VulnResult.txt":
		fileWriterContent = b
		return nil
	default:
		fileWriterContent = nil
		return fmt.Errorf("Wrong Path: %v", fileName)
	}
}

func TestWriteVulnResultToFileSuccess(t *testing.T) {

	var m map[string]int = make(map[string]int)
	m["count"] = 1
	m["cvss2GreaterOrEqualSeven"] = 2
	m["cvss3GreaterOrEqualSeven"] = 3
	m["historical_vulnerabilities"] = 4
	m["triaged_vulnerabilities"] = 5
	m["excluded_vulnerabilities"] = 6
	m["minor_vulnerabilities"] = 7
	m["major_vulnerabilities"] = 8
	m["vulnerabilities"] = 9

	cases := []struct {
		filename string
		m        map[string]int
		want     string
	}{
		{"dummy.txt", m, ""},
		{"VulnResult.txt", m, "{\"count\":1,\"cvss2GreaterOrEqualSeven\":2,\"cvss3GreaterOrEqualSeven\":3,\"excluded_vulnerabilities\":6,\"historical_vulnerabilities\":4,\"major_vulnerabilities\":8,\"minor_vulnerabilities\":7,\"triaged_vulnerabilities\":5,\"vulnerabilities\":9}"},
	}

	for _, c := range cases {

		err := WriteVulnResultToFile(c.m, c.filename, fileWriterMock)
		if(c.filename == "dummy.txt"){
			assert.NotNil(t, err)
		}else {
			assert.Nil(t, err)
		}
		assert.Equal(t, c.want, string(fileWriterContent[:]))

	}
}