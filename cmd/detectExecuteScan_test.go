package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunDetect(t *testing.T) {

	t.Run("success case", func(t *testing.T) {
		s := shellMockRunner{}
		err := runDetect(detectExecuteScanOptions{}, &s)

		assert.NoError(t, err, "Error occured but none expected")
		assert.Equal(t, ".", s.Dir, "Wrong execution directory used")
		assert.Equal(t, "/bin/bash", s.shell[0], "Bash shell expected")
		expectedScript := "bash <(curl -s https://detect.synopsys.com/detect.sh)"
		assert.Equal(t, expectedScript, s.calls[0])
	})

	t.Run("failure case", func(t *testing.T) {
		s := shellMockRunner{}
		err := runDetect(detectExecuteScanOptions{}, &s)
		if err == nil {
			t.Errorf("expected an error")
		}
	})
}

func TestAddDetectArgs(t *testing.T) {
	testData := []struct {
		args     []string
		options  detectExecuteScanOptions
		expected []string
	}{
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ScanProperties: []string{"--scan1=1", "--scan2=2"},
				ServerURL:      "https://server.url",
				APIToken:       "apiToken",
				ProjectName:    "testName",
				ProjectVersion: "1.0",
				CodeLocation:   "",
				Scanners:       []string{"signature"},
				ScanPaths:      []string{"path1", "path2"},
			},
			expected: []string{
				"--testProp1=1",
				"--scan1=1",
				"--scan2=2",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"--detect.project.name=testName",
				"--detect.project.version.name=1.0",
				"--detect.code.location.name=testName/1.0",
				"--detect.blackduck.signature.scanner.paths=path1,path2",
			},
		},
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ServerURL:      "https://server.url",
				APIToken:       "apiToken",
				ProjectName:    "testName",
				ProjectVersion: "1.0",
				CodeLocation:   "testLocation",
				Scanners:       []string{"source"},
				ScanPaths:      []string{"path1", "path2"},
			},
			expected: []string{
				"--testProp1=1",
				"--blackduck.url=https://server.url",
				"--blackduck.api.token=apiToken",
				"--detect.project.name=testName",
				"--detect.project.version.name=1.0",
				"--detect.code.location.name=testLocation",
				"--detect.source.path=path1",
			},
		},
	}

	for k, v := range testData {
		t.Run(fmt.Sprintf("run %v", k), func(t *testing.T) {
			got := addDetectArgs(v.args, v.options)
			assert.Equal(t, v.expected, got)
		})
	}
}
