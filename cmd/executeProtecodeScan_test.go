package cmd

import (
	"testing"

	"fmt"
	"os"

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
