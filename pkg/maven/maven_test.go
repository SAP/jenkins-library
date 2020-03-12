package maven

import (
	"errors"
	"os"

	"github.com/SAP/jenkins-library/pkg/mock"

	"net/http"
	"testing"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	t.Run("should return stdOut", func(t *testing.T) {
		expectedOutput := "mocked output"
		execMockRunner := mock.ExecMockRunner{}
		execMockRunner.StdoutReturn = map[string]string{"mvn --file pom.xml --batch-mode": "mocked output"}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: true}

		mavenOutput, _ := Execute(&opts, &execMockRunner)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("should not return stdOut", func(t *testing.T) {
		expectedOutput := ""
		execMockRunner := mock.ExecMockRunner{}
		execMockRunner.StdoutReturn = map[string]string{"mvn --file pom.xml --batch-mode": "mocked output"}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: false}

		mavenOutput, _ := Execute(&opts, &execMockRunner)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("should log that command failed if executing maven failed", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }
		execMockRunner := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"mvn --file pom.xml --batch-mode": errors.New("error case")}}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: false}

		output, _ := Execute(&opts, &execMockRunner)

		assert.True(t, hasFailed, "failed to execute run command")
		assert.Equal(t, output, "")
	})
	t.Run("should have all configured parameters in the exec call", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		opts := ExecuteOptions{PomPath: "pom.xml", ProjectSettingsFile: "settings.xml",
			GlobalSettingsFile: "anotherSettings.xml", M2Path: ".m2/",
			Goals: []string{"flatten", "install"}, Defines: []string{"-Da=b"},
			Flags: []string{"-q"}, LogSuccessfulMavenTransfers: true,
			ReturnStdout: false}
		expectedParameters := []string{"--global-settings", "anotherSettings.xml", "--settings", "settings.xml",
			"-Dmaven.repo.local=.m2/", "--file", "pom.xml", "-q", "-Da=b", "--batch-mode",
			"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "flatten", "install"}

		mavenOutput, _ := Execute(&opts, &execMockRunner)

		assert.Equal(t, len(execMockRunner.Calls[0].Params), len(expectedParameters))
		assert.Equal(t, execMockRunner.Calls[0], mock.ExecCall{Exec: "mvn", Params: expectedParameters})
		assert.Equal(t, "", mavenOutput)
	})
}

func TestEvaluate(t *testing.T) {
	t.Run("should evaluate expression", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		execMockRunner.StdoutReturn = map[string]string{"mvn --file pom.xml -Dexpression=project.groupId -DforceStdout -q --batch-mode org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate": "com.awesome"}

		result, err := Evaluate("pom.xml", "project.groupId", &execMockRunner)
		if assert.NoError(t, err) {
			assert.Equal(t, "com.awesome", result)
		}
	})
	t.Run("should not evaluate expression", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		execMockRunner.StdoutReturn = map[string]string{"mvn --file pom.xml -Dexpression=project.groupId -DforceStdout -q --batch-mode org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate": "null object or invalid expression"}

		result, err := Evaluate("pom.xml", "project.groupId", &execMockRunner)
		if assert.EqualError(t, err, "expression 'project.groupId' in file 'pom.xml' could not be resolved") {
			assert.Equal(t, "", result)
		}
	})
}

type mockDownloader struct {
	shouldFail bool
}

func (m *mockDownloader) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	if m.shouldFail {
		return errors.New("something happened")
	}
	return nil
}

func (m *mockDownloader) SetOptions(options piperHttp.ClientOptions) {
	return
}

func TestGetParameters(t *testing.T) {
	t.Run("should resolve configured parameters and download the settings files", func(t *testing.T) {
		mockClient := mockDownloader{shouldFail: false}
		opts := ExecuteOptions{PomPath: "pom.xml", GlobalSettingsFile: "https://mysettings.com", ProjectSettingsFile: "http://myprojectsettings.com", ReturnStdout: false}
		expectedParameters := []string{"--global-settings", "globalSettings.xml", "--settings", "projectSettings.xml", "--file", "pom.xml", "--batch-mode"}

		parameters := getParametersFromOptions(&opts, &mockClient)

		assert.Equal(t, len(parameters), len(expectedParameters))
		assert.Equal(t, parameters, expectedParameters)
	})
}

func TestDownloadSettingsFromURL(t *testing.T) {
	t.Run("should pass if download is successful", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }
		mockClient := mockDownloader{shouldFail: false}

		downloadSettingsFromURL("anyURL", "settings.xml", &mockClient)

		assert.False(t, hasFailed)
	})
	t.Run("should fail if download fails", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }
		mockClient := mockDownloader{shouldFail: true}

		downloadSettingsFromURL("anyURL", "settings.xml", &mockClient)
		assert.True(t, hasFailed, "expected command to exit with fatal")
	})
}

func TestGetTestModulesExcludes(t *testing.T) {
	t.Run("Should return excludes for unit- and integration-tests", func(t *testing.T) {
		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatal("Failed to get current working directory")
		}
		defer os.Chdir(currentDir)
		err = os.Chdir("../../test/resources/maven")
		if err != nil {
			t.Fatal("Failed to change to test directory")
		}

		expected := []string{"-pl", "!unit-tests", "-pl", "!integration-tests"}

		modulesExcludes := GetTestModulesExcludes()
		assert.Equal(t, expected, modulesExcludes)
	})
}
