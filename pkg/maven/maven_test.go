package maven

import (
	"errors"
	"github.com/SAP/jenkins-library/pkg/mock"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestExecute(t *testing.T) {
	t.Run("should return stdOut", func(t *testing.T) {
		expectedOutput := "mocked output"
		e := mock.ExecMockRunner{}
		e.StdoutReturn = map[string]string{"mvn --file pom.xml --batch-mode": "mocked output"}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: true}

		mavenOutput, _ := Execute(&opts, &e)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("should not return stdOut", func(t *testing.T) {
		expectedOutput := ""
		e := mock.ExecMockRunner{}
		e.StdoutReturn = map[string]string{"mvn --file pom.xml --batch-mode": "mocked output"}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: false}

		mavenOutput, _ := Execute(&opts, &e)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("should log that command failed if executing maven failed", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }
		e := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"mvn --file pom.xml --batch-mode": errors.New("error case")}}
		opts := ExecuteOptions{PomPath: "pom.xml", ReturnStdout: false}

		output, _ := Execute(&opts, &e)

		assert.True(t, hasFailed, "failed to execute run command")
		assert.Equal(t, output, "")
	})
	t.Run("should have all configured parameters in the exec call", func(t *testing.T) {
		e := mock.ExecMockRunner{}
		opts := ExecuteOptions{PomPath: "pom.xml", ProjectSettingsFile: "settings.xml",
			GlobalSettingsFile: "anotherSettings.xml", M2Path: ".m2/",
			Goals: []string{"flatten", "install"}, Defines: []string{"-Da=b"},
			Flags: []string{"-q"}, LogSuccessfulMavenTransfers: true,
			ReturnStdout: false}

		mavenOutput, _ := Execute(&opts, &e)

		assert.Equal(t, e.Calls[0], mock.ExecCall{Exec: "mvn", Params: []string{"--global-settings anotherSettings.xml", "--settings settings.xml",
			"-Dmaven.repo.local=.m2/", "--file pom.xml", "-q", "--batch-mode",
			"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn", "flatten", "install"}})
		assert.Equal(t, "", mavenOutput)
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

		parameters := getParametersFromOptions(&opts, &mockClient)

		assert.Equal(t, parameters, []string{"--global-settings globalSettings.xml", "--settings projectSettings.xml", "--file pom.xml", "--batch-mode"})
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
