package protecode

import (
	"testing"

	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/assert"
)

func TestParseResultSuccess(t *testing.T) {

	var result Result = Result{
		ProductId: 4712,
		ReportUrl: "ReportUrl",
		Status:    "B",
		Components: []Component{
			{Vulns: []Vulnerability{
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve1", Cvss: 7.2, Cvss3Score: "0.0"}},
				{Exact: true, Triage: []Triage{{Id: 1}}, Vuln: Vuln{Cve: "Cve2", Cvss: 2.2, Cvss3Score: "2.3"}},
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve2b", Cvss: 0.0, Cvss3Score: "0.0"}},
			},
			},
			{Vulns: []Vulnerability{
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve3", Cvss: 3.2, Cvss3Score: "7.3"}},
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve4", Cvss: 8.0, Cvss3Score: "8.0"}},
				{Exact: false, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve4b", Cvss: 8.0, Cvss3Score: "8.0"}},
			},
			},
		},
	}
	m := ParseResultForInflux(result, "Excluded CVES: Cve4,")
	t.Run("Parse Protecode Results", func(t *testing.T) {
		assert.Equal(t, 1, m["historical_vulnerabilities"])
		assert.Equal(t, 1, m["triaged_vulnerabilities"])
		assert.Equal(t, 1, m["excluded_vulnerabilities"])
		assert.Equal(t, 1, m["minor_vulnerabilities"])
		assert.Equal(t, 2, m["major_vulnerabilities"])
		assert.Equal(t, 3, m["vulnerabilities"])
	})
}

/*
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
				"test": []string{"dummy1"},
				//"authentication":         []string{"Basic auth1"},
				//"quiet":                  []string{"false"},
				//"ignoreSslErrors":        []string{"true"},
				//"consoleLogResponseBody": []string{"true"},
			}},
	}

	for _, c := range cases {

		got := CreateRequestHeader(c.verbose, c.auth, c.hMap)
		assert.Equal(t, c.want, got)
	}
}
*/
func TestGetResultData(t *testing.T) {

	cases := []struct {
		give string
		want ResultData
	}{
		{`{"results": {"product_id": 1}}`, ResultData{Result: Result{ProductId: 1}}},
	}

	for _, c := range cases {

		r := ioutil.NopCloser(bytes.NewReader([]byte(c.give)))
		got, _ := GetResultData(r)
		assert.Equal(t, c.want, *got)
	}
}

func TestGetProductData(t *testing.T) {

	cases := []struct {
		give string
		want ProductData
	}{
		{`{"products": [{"product_id": 1}]}`, ProductData{Products: []Product{{ProductId: 1}}}},
	}

	for _, c := range cases {

		r := ioutil.NopCloser(bytes.NewReader([]byte(c.give)))
		got, _ := GetProductData(r)
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

func TestWriteResultAsJSONToFileSuccess(t *testing.T) {

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

		err := WriteResultAsJSONToFile(c.m, c.filename, fileWriterMock)
		if c.filename == "dummy.txt" {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, c.want, string(fileWriterContent[:]))

	}
}

func TestParseResultViolations(t *testing.T) {

	violations := filepath.Join("testdata", "protecode_result_violations.json")
	byteContent, err := ioutil.ReadFile(violations)
	if err != nil {
		t.Fatalf("failed reading %v", violations)
	}

	resultData, _ := GetResultData(ioutil.NopCloser(strings.NewReader(string(byteContent))))

	m := ParseResultForInflux(resultData.Result, "CVE-2018-1, CVE-2017-1000382")
	t.Run("Parse Protecode Results", func(t *testing.T) {
		assert.Equal(t, 1125, m["historical_vulnerabilities"])
		assert.Equal(t, 0, m["triaged_vulnerabilities"])
		assert.Equal(t, 1, m["excluded_vulnerabilities"])
		assert.Equal(t, 129, m["cvss3GreaterOrEqualSeven"])
		assert.Equal(t, 13, m["cvss2GreaterOrEqualSeven"])
		assert.Equal(t, 226, m["vulnerabilities"])
	})
}

func TestParseResultNoViolations(t *testing.T) {

	noViolations := filepath.Join("testdata", "protecode_result_no_violations.json")
	byteContent, err := ioutil.ReadFile(noViolations)
	if err != nil {
		t.Fatalf("failed reading %v", noViolations)
	}

	resultData, _ := GetResultData(ioutil.NopCloser(strings.NewReader(string(byteContent))))

	m := ParseResultForInflux(resultData.Result, "CVE-2018-1, CVE-2017-1000382")
	t.Run("Parse Protecode Results", func(t *testing.T) {
		assert.Equal(t, 27, m["historical_vulnerabilities"])
		assert.Equal(t, 0, m["triaged_vulnerabilities"])
		assert.Equal(t, 0, m["excluded_vulnerabilities"])
		assert.Equal(t, 0, m["cvss3GreaterOrEqualSeven"])
		assert.Equal(t, 0, m["cvss2GreaterOrEqualSeven"])
		assert.Equal(t, 0, m["vulnerabilities"])
	})
}

func TestParseResultTriaged(t *testing.T) {

	triaged := filepath.Join("testdata", "protecode_result_triaging.json")
	byteContent, err := ioutil.ReadFile(triaged)
	if err != nil {
		t.Fatalf("failed reading %v", triaged)
	}

	resultData, _ := GetResultData(ioutil.NopCloser(strings.NewReader(string(byteContent))))

	m := ParseResultForInflux(resultData.Result, "")
	t.Run("Parse Protecode Results", func(t *testing.T) {
		assert.Equal(t, 1132, m["historical_vulnerabilities"])
		assert.Equal(t, 187, m["triaged_vulnerabilities"])
		assert.Equal(t, 0, m["excluded_vulnerabilities"])
		assert.Equal(t, 15, m["cvss3GreaterOrEqualSeven"])
		assert.Equal(t, 0, m["cvss2GreaterOrEqualSeven"])
		assert.Equal(t, 36, m["vulnerabilities"])
	})
}
