//go:build unit

package maven

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/mock"

	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockUtils struct {
	shouldFail     bool
	requestedUrls  []string
	requestedFiles []string
	*mock.FilesMock
	*mock.ExecMockRunner
}

func (m *MockUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	m.requestedUrls = append(m.requestedUrls, url)
	m.requestedFiles = append(m.requestedFiles, filename)
	if m.shouldFail {
		return errors.New("something happened")
	}
	return nil
}

func NewMockUtils(downloadShouldFail bool) MockUtils {
	utils := MockUtils{
		shouldFail:     downloadShouldFail,
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

func TestExecute(t *testing.T) {
	t.Run("should return stdOut", func(t *testing.T) {
		expectedOutput := "mocked output"
		utils := NewMockUtils(false)
		utils.StdoutReturn = map[string]string{"mvn --file pom.xml -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode": "mocked output"}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: true}

		mavenOutput, _ := Execute(&opts, &utils)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("should not return stdOut", func(t *testing.T) {
		expectedOutput := ""
		utils := NewMockUtils(false)
		utils.StdoutReturn = map[string]string{"mvn --file pom.xml -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode": "mocked output"}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: false}

		mavenOutput, _ := Execute(&opts, &utils)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("should log that command failed if executing maven failed", func(t *testing.T) {
		utils := NewMockUtils(false)
		utils.ShouldFailOnCommand = map[string]error{"mvn --file pom.xml -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode": errors.New("error case")}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: false}

		output, err := Execute(&opts, &utils)

		assert.EqualError(t, err, "failed to run executable, command: '[mvn --file pom.xml -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode]', error: error case")
		assert.Equal(t, "", output)
	})
	t.Run("should have all configured parameters in the exec call", func(t *testing.T) {
		utils := NewMockUtils(false)
		opts := ExecuteOptions{PomPath: "pom.xml", ProjectSettingsFile: "settings.xml",
			GlobalSettingsFile: "anotherSettings.xml", M2Path: ".m2/",
			Goals: []string{"flatten", "install"}, Defines: []string{"-Da=b"},
			Flags: []string{"-q"}, LogSuccessfulMavenTransfers: true,
			ReturnStdout: false}
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, "anotherSettings.xml")
		projectSettingsPath := filepath.Join(dir, "settings.xml")
		expectedParameters := []string{"--global-settings", globalSettingsPath, "--settings", projectSettingsPath,
			"-Dmaven.repo.local=.m2/", "--file", "pom.xml", "-q", "-Da=b", "--batch-mode",
			"flatten", "install"}

		mavenOutput, _ := Execute(&opts, &utils)

		assert.Equal(t, len(expectedParameters), len(utils.Calls[0].Params))
		assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters}, utils.Calls[0])
		assert.Equal(t, "", mavenOutput)
	})
}

func TestEvaluate(t *testing.T) {
	t.Run("should evaluate expression", func(t *testing.T) {
		utils := NewMockUtils(false)
		utils.StdoutReturn = map[string]string{"mvn --file pom.xml -Dexpression=project.groupId -DforceStdout -q -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate": "com.awesome"}

		result, err := Evaluate(&EvaluateOptions{PomPath: "pom.xml"}, "project.groupId", &utils)
		if assert.NoError(t, err) {
			assert.Equal(t, "com.awesome", result)
		}
	})
	t.Run("should not evaluate expression", func(t *testing.T) {
		utils := NewMockUtils(false)
		utils.StdoutReturn = map[string]string{"mvn --file pom.xml -Dexpression=project.groupId -DforceStdout -q -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate": "null object or invalid expression"}

		result, err := Evaluate(&EvaluateOptions{PomPath: "pom.xml"}, "project.groupId", &utils)
		if assert.EqualError(t, err, "expression 'project.groupId' in file 'pom.xml' could not be resolved") {
			assert.Equal(t, "", result)
		}
	})
}

func TestGetParameters(t *testing.T) {
	t.Run("should resolve configured parameters and download the settings files", func(t *testing.T) {
		utils := NewMockUtils(false)
		opts := ExecuteOptions{PomPath: "pom.xml", GlobalSettingsFile: "https://mysettings.com", ProjectSettingsFile: "http://myprojectsettings.com", ReturnStdout: false}
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, ".pipeline", "mavenGlobalSettings.xml")
		projectSettingsPath := filepath.Join(dir, ".pipeline", "mavenProjectSettings.xml")
		expectedParameters := []string{
			"--global-settings", globalSettingsPath,
			"--settings", projectSettingsPath,
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
		utils := NewMockUtils(false)
		utils.AddFile(".pipeline/mavenGlobalSettings.xml", []byte("dummyContent"))
		utils.AddFile(".pipeline/mavenProjectSettings.xml", []byte("dummyContent"))
		opts := ExecuteOptions{PomPath: "pom.xml", GlobalSettingsFile: "https://mysettings.com", ProjectSettingsFile: "http://myprojectsettings.com", ReturnStdout: false}
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, ".pipeline", "mavenGlobalSettings.xml")
		projectSettingsPath := filepath.Join(dir, ".pipeline", "mavenProjectSettings.xml")
		expectedParameters := []string{
			"--global-settings", globalSettingsPath,
			"--settings", projectSettingsPath,
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

func TestGetTestModulesExcludes(t *testing.T) {
	t.Run("Should return excludes for unit- and integration-tests", func(t *testing.T) {
		utils := NewMockUtils(false)
		utils.AddFile("unit-tests/pom.xml", []byte("dummyContent"))
		utils.AddFile("integration-tests/pom.xml", []byte("dummyContent"))
		expected := []string{"-pl", "!unit-tests", "-pl", "!integration-tests"}

		modulesExcludes := GetTestModulesExcludes(&utils)
		assert.Equal(t, expected, modulesExcludes)
	})
	t.Run("Should not return excludes for unit- and integration-tests", func(t *testing.T) {
		utils := NewMockUtils(false)

		var expected []string

		modulesExcludes := GetTestModulesExcludes(&utils)
		assert.Equal(t, expected, modulesExcludes)
	})
}

func TestMavenInstall(t *testing.T) {
	t.Parallel()
	t.Run("Should return path to jar file", func(t *testing.T) {
		actual := jarFile("app", "my-app")
		assert.Equal(t, filepath.Join("app", "target", "my-app.jar"), actual)
	})

	t.Run("Should return path to war file", func(t *testing.T) {
		actual := warFile("app", "my-app")
		assert.Equal(t, filepath.Join("app", "target", "my-app.war"), actual)
	})

	t.Run("Install a file", func(t *testing.T) {
		utils := NewMockUtils(false)
		expectedParameters := []string{"-Dfile=app.jar", "-Dpackaging=jar", "-DpomFile=pom.xml", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "install:install-file"}

		err := InstallFile("app.jar", "pom.xml", &EvaluateOptions{}, &utils)

		assert.NoError(t, err)
		if assert.Equal(t, len(expectedParameters), len(utils.Calls[0].Params)) {
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters}, utils.Calls[0])
		}
	})

	t.Run("Install files in a project", func(t *testing.T) {
		utils := NewMockUtils(false)
		utils.AddFile("target/foo.jar", []byte("dummyContent"))
		utils.AddFile("target/foo.war", []byte("dummyContent"))
		utils.AddFile("pom.xml", []byte("<project></project>"))

		options := EvaluateOptions{}
		options.ProjectSettingsFile = "settings.xml"
		dir, _ := os.Getwd()
		projectSettingsPath := filepath.Join(dir, "settings.xml")
		utils.StdoutReturn = map[string]string{"mvn --settings " + projectSettingsPath + " --file pom.xml -Dexpression=project.build.finalName -DforceStdout -q -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate": "foo"}
		err := doInstallMavenArtifacts(&options, &utils)
		assert.NoError(t, err)
		if assert.Equal(t, 5, len(utils.Calls)) {
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--settings", projectSettingsPath, "-Dflatten.mode=resolveCiFriendliesOnly", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "flatten:flatten"}}, utils.Calls[0])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--settings", projectSettingsPath, "--file", "pom.xml", "-Dexpression=project.packaging", "-DforceStdout", "-q", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"}}, utils.Calls[1])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--settings", projectSettingsPath, "--file", "pom.xml", "-Dexpression=project.build.finalName", "-DforceStdout", "-q", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"}}, utils.Calls[2])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--settings", projectSettingsPath, "-Dfile=" + filepath.Join(".", "target", "foo.jar"), "-Dpackaging=jar", "-DpomFile=pom.xml", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "install:install-file"}}, utils.Calls[3])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--settings", projectSettingsPath, "-Dfile=" + filepath.Join(".", "target", "foo.war"), "-DpomFile=pom.xml", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "install:install-file"}}, utils.Calls[4])
		}
	})

	t.Run("Install files in a spring-boot project", func(t *testing.T) {
		utils := NewMockUtils(false)
		utils.AddFile("target/foo.jar", []byte("dummyContent"))
		utils.AddFile("target/foo.jar.original", []byte("dummyContent"))
		utils.AddFile("pom.xml", []byte("<project></project>"))

		options := EvaluateOptions{}
		utils.StdoutReturn = map[string]string{"mvn --file pom.xml -Dexpression=project.build.finalName -DforceStdout -q -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate": "foo"}
		err := doInstallMavenArtifacts(&options, &utils)

		assert.NoError(t, err)
		if assert.Equal(t, 4, len(utils.Calls)) {
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"-Dflatten.mode=resolveCiFriendliesOnly", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "flatten:flatten"}}, utils.Calls[0])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--file", "pom.xml", "-Dexpression=project.packaging", "-DforceStdout", "-q", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"}}, utils.Calls[1])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--file", "pom.xml", "-Dexpression=project.build.finalName", "-DforceStdout", "-q", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"}}, utils.Calls[2])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"-Dfile=" + filepath.Join(".", "target", "foo.jar.original"), "-Dpackaging=jar", "-DpomFile=pom.xml", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "install:install-file"}}, utils.Calls[3])
		}
	})

	t.Run("Install files in a multi-module-project", func(t *testing.T) {
		utils := NewMockUtils(false)
		utils.AddFile("parent/module1/target/module1.jar", []byte("dummyContent"))
		utils.AddFile("parent/module1/target/module1.jar.original", []byte("dummyContent"))
		utils.AddFile("parent/pom.xml", []byte("<project></project>"))
		utils.AddFile("parent/module1/pom.xml", []byte(
			"<project></project>"))

		options := EvaluateOptions{
			PomPath: filepath.Join(".", "parent", "pom.xml"),
		}
		utils.StdoutReturn = map[string]string{`mvn --file ` + `.*module1.*` + `-Dexpression=project.build.finalName -DforceStdout -q -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --batch-mode org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate`: "module1"}
		err := doInstallMavenArtifacts(&options, &utils)

		assert.NoError(t, err)
		if assert.Equal(t, 7, len(utils.Calls)) {
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--file", filepath.Join(".", "parent", "pom.xml"), "-Dflatten.mode=resolveCiFriendliesOnly", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "flatten:flatten"}}, utils.Calls[0])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--file", filepath.Join(".", "parent", "module1", "pom.xml"), "-Dexpression=project.packaging", "-DforceStdout", "-q", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"}}, utils.Calls[1])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--file", filepath.Join(".", "parent", "module1", "pom.xml"), "-Dexpression=project.build.finalName", "-DforceStdout", "-q", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"}}, utils.Calls[2])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"-Dfile=" + filepath.Join(".", "parent/module1/target", "module1.jar.original"), "-Dpackaging=jar", "-DpomFile=" + filepath.Join("parent/module1/pom.xml"), "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "install:install-file"}}, utils.Calls[3])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--file", filepath.Join(".", "parent", "pom.xml"), "-Dexpression=project.packaging", "-DforceStdout", "-q", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"}}, utils.Calls[4])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"--file", filepath.Join(".", "parent", "pom.xml"), "-Dexpression=project.build.finalName", "-DforceStdout", "-q", "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"}}, utils.Calls[5])
			assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"-Dfile=" + filepath.Join(".", "parent", "pom.xml"), "-DpomFile=" + filepath.Join("parent/pom.xml"), "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "--batch-mode", "install:install-file"}}, utils.Calls[6])
		}
	})

}
