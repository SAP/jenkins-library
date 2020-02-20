package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunMavenExecute(t *testing.T) {
	t.Run("runMavenExecute should return stdOut", func(t *testing.T) {
		expectedOutput := "mocked output"
		e := execMockRunner{}
		e.stdoutReturn = map[string]string{"mvn --file pom.xml --batch-mode": "mocked output"}
		opts := mavenExecuteOptions{PomPath: "pom.xml", ReturnStdout: true}

		mavenOutput, _ := runMavenExecute(&opts, &e)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("runMavenExecute should not return stdOut", func(t *testing.T) {
		expectedOutput := ""
		e := execMockRunner{}
		e.stdoutReturn = map[string]string{"mvn --file pom.xml --batch-mode": "mocked output"}
		opts := mavenExecuteOptions{PomPath: "pom.xml", ReturnStdout: false}

		mavenOutput, _ := runMavenExecute(&opts, &e)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("runMavenExecute should have all config parameters in the exec call", func(t *testing.T) {
		e := execMockRunner{}
		opts := mavenExecuteOptions{PomPath: "pom.xml", ProjectSettingsFile: "settings.xml",
			GlobalSettingsFile: "anotherSettings.xml", M2Path: ".m2/",
			Goals: []string{"flatten", "install"}, Defines: []string{"-Da=b"},
			Flags: []string{"-q"}, LogSuccessfulMavenTransfers: true,
			ReturnStdout: false}

		mavenOutput, _ := runMavenExecute(&opts, &e)

		assert.Equal(t, e.calls[0], execCall{exec: "mvn", params: []string{"--global-settings anotherSettings.xml", "--settings settings.xml",
			"-Dmaven.repo.local=.m2/", "--file pom.xml", "-q", "--batch-mode",
			"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "flatten", "install"}})
		assert.Equal(t, "", mavenOutput)
	})
}
