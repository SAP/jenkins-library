package cmd

import (
	"testing"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/stretchr/testify/assert"
)

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

func TestUploadScanOrDeclareFetch(t *testing.T) {
	requestURI := ""
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI
		response := protecode.Result{ProductId: 4711, ReportUrl: requestURI}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))
	}))

	// Close the server when test finishes
	defer server.Close()

	po := protecode.ProtecodeOptions{ServerURL: server.URL}
	pc := protecode.Protecode{}
	pc.SetOptions(po)

	cases := []struct {
		reuse    bool
		clean    string
		group    string
		fetchUrl string
		want     int
	}{
		{false, "test", "group1", "/api/fetch/", 4711},
	}

	for _, c := range cases {
		config := protecodeExecuteScanOptions{ReuseExisting: c.reuse, CleanupMode: c.clean, ProtecodeGroup: c.group, FetchURL: c.fetchUrl}

		got, _ := uploadScanOrDeclareFetch(config, 0, pc)

		assert.Equal(t, c.want, got)
		assert.Equal(t, c.fetchUrl, requestURI)
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

		err := writeResultAsJSONToFile(c.m, c.filename, fileWriterMock)
		if c.filename == "dummy.txt" {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, c.want, string(fileWriterContent[:]))

	}
}
