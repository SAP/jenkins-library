//go:build unit
// +build unit

package cmd

import (
	"errors"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type mavenExecuteIntegrationTestUtilsBundle struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func (m mavenExecuteIntegrationTestUtilsBundle) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return errors.New("Test should not download files.")
}

func TestIntegrationTestModuleDoesNotExist(t *testing.T) {
	t.Parallel()
	utils := newMavenIntegrationTestsUtilsBundle()
	config := mavenExecuteIntegrationOptions{}

	err := runMavenExecuteIntegration(&config, utils)

	assert.EqualError(t, err, "maven module 'integration-tests' does not exist in project structure")
}

func TestHappyPathIntegrationTests(t *testing.T) {
	t.Parallel()
	utils := newMavenIntegrationTestsUtilsBundle()
	utils.FilesMock.AddFile("integration-tests/pom.xml", []byte(`<project> </project>`))

	config := mavenExecuteIntegrationOptions{
		Retry:     2,
		ForkCount: "1C",
		Goal:      "post-integration-test",
	}

	err := runMavenExecuteIntegration(&config, utils)
	if err != nil {
		t.Fatalf("Error %s", err)
	}

	expectedParameters1 := []string{
		"--file",
		filepath.Join(".", "integration-tests", "pom.xml"),
		"-Dsurefire.rerunFailingTestsCount=2",
		"-Dsurefire.forkCount=1C",
		"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
		"--batch-mode",
		"org.jacoco:jacoco-maven-plugin:prepare-agent",
		"post-integration-test",
	}

	assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters1}, utils.ExecMockRunner.Calls[0])
}

func TestHappyPathIntegrationTestsWithReactorInstall(t *testing.T) {
	t.Parallel()
	utils := newMavenIntegrationTestsUtilsBundle()
	utils.FilesMock.AddFile("integration-tests/pom.xml", []byte(`<project> </project>`))

	config := mavenExecuteIntegrationOptions{
		Retry:              2,
		ForkCount:          "1C",
		Goal:               "post-integration-test",
		InstallWithReactor: true,
	}

	err := runMavenExecuteIntegration(&config, utils)
	if err != nil {
		t.Fatalf("Error %s", err)
	}

	expectedParameters1 := []string{
		"-pl",
		"integration-tests",
		"-am",
		"-DskipTests",
		"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
		"--batch-mode",
		"install",
	}

	expectedParameters2 := []string{
		"--file",
		filepath.Join(".", "integration-tests", "pom.xml"),
		"-Dsurefire.rerunFailingTestsCount=2",
		"-Dsurefire.forkCount=1C",
		"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
		"--batch-mode",
		"org.jacoco:jacoco-maven-plugin:prepare-agent",
		"post-integration-test",
	}

	assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters1}, utils.ExecMockRunner.Calls[0])
	assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters2}, utils.ExecMockRunner.Calls[1])
}

func TestMutualExclusivityOfInstallFlags(t *testing.T) {
	t.Parallel()
	utils := newMavenIntegrationTestsUtilsBundle()
	utils.FilesMock.AddFile("integration-tests/pom.xml", []byte(`<project> </project>`))

	config := mavenExecuteIntegrationOptions{
		InstallArtifacts:   true,
		InstallWithReactor: true,
	}

	err := runMavenExecuteIntegration(&config, utils)

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "parameters 'installArtifacts' and 'installWithReactor' are mutually exclusive")
	}
}

func TestInvalidForkCountParam(t *testing.T) {
	t.Parallel()
	// init
	utils := newMavenIntegrationTestsUtilsBundle()
	utils.FilesMock.AddFile("integration-tests/pom.xml", []byte(`<project> </project>`))

	// test
	err := runMavenExecuteIntegration(&mavenExecuteIntegrationOptions{ForkCount: "4.2"}, utils)

	// assert
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid forkCount parameter")
	}
}

func TestValidateForkCount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		testValue     string
		expectedError string
	}{
		{
			name:          "valid integer",
			testValue:     "2",
			expectedError: "",
		},
		{
			name:          "zero is valid",
			testValue:     "0",
			expectedError: "",
		},
		{
			name:          "valid floating point",
			testValue:     "2.5C",
			expectedError: "",
		},
		{
			name:          "valid integer with C",
			testValue:     "2C",
			expectedError: "",
		},
		{
			name:          "invalid floating point",
			testValue:     "1.2",
			expectedError: "invalid forkCount parameter",
		},
		{
			name:          "invalid",
			testValue:     "C1",
			expectedError: "invalid forkCount parameter",
		},
		{
			name:          "another invalid",
			testValue:     "1 C",
			expectedError: "invalid forkCount parameter",
		},
		{
			name:          "invalid float",
			testValue:     "1..2C",
			expectedError: "invalid forkCount parameter",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			err := validateForkCount(testCase.testValue)
			if testCase.expectedError == "" {
				assert.NoError(t, err)
			} else if assert.Error(t, err) {
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func newMavenIntegrationTestsUtilsBundle() *mavenExecuteIntegrationTestUtilsBundle {
	utilsBundle := mavenExecuteIntegrationTestUtilsBundle{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return &utilsBundle
}
