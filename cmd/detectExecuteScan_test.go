package cmd

import (
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

func TestRunDetect(t *testing.T) {

	httpClient := piperhttp.Client{}

	t.Run("success case", func(t *testing.T) {
		s := mock.ShellMockRunner{}
		fileUtilsMock := mock.FilesMock{}
		err := runDetect(detectExecuteScanOptions{}, &s, &fileUtilsMock, &httpClient)

		assert.Nil(t, err)
		assert.Equal(t, ".", s.Dir, "Wrong execution directory used")
		assert.Equal(t, "/bin/bash", s.Shell[0], "Bash shell expected")
		expectedScript := "bash <(curl -s https://detect.synopsys.com/detect.sh) --blackduck.url= --blackduck.api.token= --detect.project.name= --detect.project.version.name= --detect.code.location.name="
		assert.Equal(t, expectedScript, s.Calls[0])
	})

	t.Run("failure case", func(t *testing.T) {
		s := mock.ShellMockRunner{ShouldFailOnCommand: map[string]error{"bash <(curl -s https://detect.synopsys.com/detect.sh) --blackduck.url= --blackduck.api.token= --detect.project.name= --detect.project.version.name= --detect.code.location.name=": fmt.Errorf("Test Error")}}
		fileUtilsMock := mock.FilesMock{}
		err := runDetect(detectExecuteScanOptions{}, &s, &fileUtilsMock, &httpClient)
		assert.NotNil(t, err)
	})

	t.Run("maven parameters", func(t *testing.T) {
		s := mock.ShellMockRunner{}
		fileUtilsMock := mock.FilesMock{
			CurrentDir: "root_folder",
		}
		err := runDetect(detectExecuteScanOptions{
			M2Path:              ".pipeline/local_repo",
			ProjectSettingsFile: "project-settings.xml",
			GlobalSettingsFile:  "global-settings.xml",
		}, &s, &fileUtilsMock, &httpClient)

		assert.Nil(t, err)
		assert.Equal(t, ".", s.Dir, "Wrong execution directory used")
		assert.Equal(t, "/bin/bash", s.Shell[0], "Bash shell expected")
		absoluteLocalPath := string(os.PathSeparator) + filepath.Join("root_folder", ".pipeline", "local_repo")

		expectedParam := "--detect.maven.build.command='--global-settings global-settings.xml --settings project-settings.xml -Dmaven.repo.local="+absoluteLocalPath+"'"
		assert.Contains(t, s.Calls[0], expectedParam)
	})
}

func TestAddDetectArgs(t *testing.T) {
	httpClient := piperhttp.Client{}
	fileUtilsMock := mock.FilesMock{}

	testData := []struct {
		args     []string
		options  detectExecuteScanOptions
		expected []string
	}{
		{
			args: []string{"--testProp1=1"},
			options: detectExecuteScanOptions{
				ScanProperties:  []string{"--scan1=1", "--scan2=2"},
				ServerURL:       "https://server.url",
				APIToken:        "apiToken",
				ProjectName:     "testName",
				Version:         "1.0",
				VersioningModel: "major-minor",
				CodeLocation:    "",
				Scanners:        []string{"signature"},
				ScanPaths:       []string{"path1", "path2"},
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
				ServerURL:       "https://server.url",
				APIToken:        "apiToken",
				ProjectName:     "testName",
				Version:         "1.0",
				VersioningModel: "major-minor",
				CodeLocation:    "testLocation",
				Scanners:        []string{"source"},
				ScanPaths:       []string{"path1", "path2"},
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
			got, err := addDetectArgs(v.args, v.options, &fileUtilsMock, &httpClient)
			assert.Nil(t, err)
			assert.Equal(t, v.expected, got)
		})
	}
}
