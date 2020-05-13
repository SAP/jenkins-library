package maven

import (
	"errors"
	"fmt"
	"os"

	"github.com/SAP/jenkins-library/pkg/mock"

	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockUtils struct {
	shouldFail     bool
	requestedUrls  []string
	requestedFiles []string
	files          map[string][]byte
}

func (m *mockUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	m.requestedUrls = append(m.requestedUrls, url)
	m.requestedFiles = append(m.requestedFiles, filename)
	if m.shouldFail {
		return errors.New("something happened")
	}
	return nil
}

func (m *mockUtils) FileExists(path string) (bool, error) {
	content := m.files[path]
	if content == nil {
		return false, fmt.Errorf("'%s': %w", path, os.ErrNotExist)
	}
	return true, nil
}

func newMockUtils(downloadShouldFail bool) mockUtils {
	utils := mockUtils{shouldFail: downloadShouldFail}
	utils.files = map[string][]byte{}
	return utils
}

func TestExecute(t *testing.T) {
	t.Run("should return stdOut", func(t *testing.T) {
		expectedOutput := "mocked output"
		execMockRunner := mock.ExecMockRunner{}
		execMockRunner.StdoutReturn = map[string]string{"mvn --file pom.xml -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode": "mocked output"}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: true}

		mavenOutput, _ := Execute(&opts, &execMockRunner)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("should not return stdOut", func(t *testing.T) {
		expectedOutput := ""
		execMockRunner := mock.ExecMockRunner{}
		execMockRunner.StdoutReturn = map[string]string{"mvn --file pom.xml -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode": "mocked output"}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: false}

		mavenOutput, _ := Execute(&opts, &execMockRunner)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("should log that command failed if executing maven failed", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"mvn --file pom.xml -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode": errors.New("error case")}}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: false}

		output, err := Execute(&opts, &execMockRunner)

		assert.EqualError(t, err, "failed to run executable, command: '[mvn --file pom.xml -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode]', error: error case")
		assert.Equal(t, "", output)
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
			"flatten", "install"}

		mavenOutput, _ := Execute(&opts, &execMockRunner)

		assert.Equal(t, len(expectedParameters), len(execMockRunner.Calls[0].Params))
		assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters}, execMockRunner.Calls[0])
		assert.Equal(t, "", mavenOutput)
	})
}

func TestEvaluate(t *testing.T) {
	t.Run("should evaluate expression", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		execMockRunner.StdoutReturn = map[string]string{"mvn --file pom.xml -Dexpression=project.groupId -DforceStdout -q -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate": "com.awesome"}

		result, err := Evaluate("pom.xml", "project.groupId", &execMockRunner)
		if assert.NoError(t, err) {
			assert.Equal(t, "com.awesome", result)
		}
	})
	t.Run("should not evaluate expression", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		execMockRunner.StdoutReturn = map[string]string{"mvn --file pom.xml -Dexpression=project.groupId -DforceStdout -q -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate": "null object or invalid expression"}

		result, err := Evaluate("pom.xml", "project.groupId", &execMockRunner)
		if assert.EqualError(t, err, "expression 'project.groupId' in file 'pom.xml' could not be resolved") {
			assert.Equal(t, "", result)
		}
	})
}

func TestGetParameters(t *testing.T) {
	t.Run("should resolve configured parameters and download the settings files", func(t *testing.T) {
		utils := newMockUtils(false)
		opts := ExecuteOptions{PomPath: "pom.xml", GlobalSettingsFile: "https://mysettings.com", ProjectSettingsFile: "http://myprojectsettings.com", ReturnStdout: false}
		expectedParameters := []string{
			"--global-settings", ".pipeline/mavenGlobalSettings.xml",
			"--settings", ".pipeline/mavenProjectSettings.xml",
			"--file", "pom.xml",
			"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
			"--batch-mode"}

		parameters, err := getParametersFromOptions(&opts, &utils)
		if assert.NoError(t, err) {
			assert.Equal(t, len(expectedParameters), len(parameters))
			assert.Equal(t, expectedParameters, parameters)
			if assert.Equal(t, 2, len(utils.requestedUrls)) {
				assert.Equal(t, "https://mysettings.com", utils.requestedUrls[0])
				assert.Equal(t, ".pipeline/mavenGlobalSettings.xml", utils.requestedFiles[0])
				assert.Equal(t, "http://myprojectsettings.com", utils.requestedUrls[1])
				assert.Equal(t, ".pipeline/mavenProjectSettings.xml", utils.requestedFiles[1])
			}
		}
	})
	t.Run("should resolve configured parameters and not download existing settings files", func(t *testing.T) {
		utils := newMockUtils(false)
		utils.files[".pipeline/mavenGlobalSettings.xml"] = []byte("dummyContent")
		utils.files[".pipeline/mavenProjectSettings.xml"] = []byte("dummyContent")
		opts := ExecuteOptions{PomPath: "pom.xml", GlobalSettingsFile: "https://mysettings.com", ProjectSettingsFile: "http://myprojectsettings.com", ReturnStdout: false}
		expectedParameters := []string{
			"--global-settings", ".pipeline/mavenGlobalSettings.xml",
			"--settings", ".pipeline/mavenProjectSettings.xml",
			"--file", "pom.xml",
			"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
			"--batch-mode"}

		parameters, err := getParametersFromOptions(&opts, &utils)
		if assert.NoError(t, err) {
			assert.Equal(t, len(expectedParameters), len(parameters))
			assert.Equal(t, expectedParameters, parameters)
			assert.Equal(t, 0, len(utils.requestedUrls))
		}
	})
}

func TestDownloadSettingsFromURL(t *testing.T) {
	t.Run("should pass if download is successful", func(t *testing.T) {
		utils := newMockUtils(false)
		err := downloadSettingsFromURL("anyURL", "settings.xml", &utils)
		assert.NoError(t, err)
	})
	t.Run("should fail if download fails", func(t *testing.T) {
		utils := newMockUtils(true)
		err := downloadSettingsFromURL("anyURL", "settings.xml", &utils)
		assert.EqualError(t, err, "failed to download maven settings from URL 'anyURL' to file 'settings.xml': something happened")
	})
}

func TestGetTestModulesExcludes(t *testing.T) {
	t.Run("Should return excludes for unit- and integration-tests", func(t *testing.T) {
		utils := newMockUtils(false)
		utils.files["unit-tests/pom.xml"] = []byte("dummyPomContent")
		utils.files["integration-tests/pom.xml"] = []byte("dummyPomContent")

		expected := []string{"-pl", "!unit-tests", "-pl", "!integration-tests"}

		modulesExcludes := getTestModulesExcludes(&utils)
		assert.Equal(t, expected, modulesExcludes)
	})
	t.Run("Should not return excludes for unit- and integration-tests", func(t *testing.T) {
		utils := newMockUtils(false)

		var expected []string

		modulesExcludes := getTestModulesExcludes(&utils)
		assert.Equal(t, expected, modulesExcludes)
	})
}
