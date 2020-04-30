package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRunDetect(t *testing.T) {

	t.Run("success case", func(t *testing.T) {
		s := mock.ShellMockRunner{}
		runDetect(detectExecuteScanOptions{}, &s)

		assert.Equal(t, ".", s.Dir, "Wrong execution directory used")
		assert.Equal(t, "/bin/bash", s.Shell[0], "Bash shell expected")
		expectedScript := "bash <(curl -s https://detect.synopsys.com/detect.sh) --blackduck.url= --blackduck.api.token= --detect.project.name= --detect.project.version.name= --detect.code.location.name="
		assert.Equal(t, expectedScript, s.Calls[0])
	})

	t.Run("failure case", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }

		s := mock.ShellMockRunner{ShouldFailOnCommand: map[string]error{"bash <(curl -s https://detect.synopsys.com/detect.sh) --blackduck.url= --blackduck.api.token= --detect.project.name= --detect.project.version.name= --detect.code.location.name=": fmt.Errorf("Test Error")}}
		runDetect(detectExecuteScanOptions{}, &s)
		assert.True(t, hasFailed, "expected command to exit with fatal")
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
