package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mavenExecuteIntegrationTestUtilsBundle struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func TestIntegrationTestModuleDoesNotExist(t *testing.T) {
	utils := newMavenIntegrationTestsUtilsBundle()
	config := mavenExecuteIntegrationOptions{}

	err := runMavenExecuteIntegration(&config, utils)

	assert.EqualError(t, err, "maven module 'integration-tests' does not exist in project structure")
}

func TestHappyPathIntegrationTests(t *testing.T) {
	utils := newMavenIntegrationTestsUtilsBundle()
	utils.FilesMock.AddFile("integration-tests/pom.xml", []byte(`<project> </project>`))

	config := mavenExecuteIntegrationOptions{
		Retry:     2,
		ForkCount: "1C",
	}

	err := runMavenExecuteIntegration(&config, utils)
	if err != nil {
		t.Fatalf("Error %s", err)
	}

	expectedParameters1 := []string{
		"--file",
		"integration-tests/pom.xml",
		"-Dsurefire.rerunFailingTestsCount=2",
		"-Dsurefire.forkCount=1C",
		"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
		"--batch-mode",
		"org.jacoco:jacoco-maven-plugin:prepare-agent",
		"test",
	}

	assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters1}, utils.ExecMockRunner.Calls[0])
}

func TestIntegrationTestsWithInstallArtifacts(t *testing.T) {
	utils := newMavenIntegrationTestsUtilsBundle()
	utils.FilesMock.AddFile("pom.xml", []byte(`<project> </project>`))
	utils.FilesMock.AddFile("application/pom.xml", []byte(`<project> </project>`))
	utils.FilesMock.AddFile("application/target/application.jar", []byte(`<project> </project>`))
	utils.FilesMock.AddFile("integration-tests/pom.xml", []byte(`<project> </project>`))

	config := mavenExecuteIntegrationOptions{
		Retry:     2,
		ForkCount: "1C",
		InstallArtifacts: true,
	}

	err := runMavenExecuteIntegration(&config, utils)
	if err != nil {
		t.Fatalf("Error %s", err)
	}

	expectedParameters1 := []string{
		"--file",
		"integration-tests/pom.xml",
		"-Dsurefire.rerunFailingTestsCount=2",
		"-Dsurefire.forkCount=1C",
		"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
		"--batch-mode",
		"org.jacoco:jacoco-maven-plugin:prepare-agent",
		"test",
	}

	assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: expectedParameters1}, utils.ExecMockRunner.Calls[0])
}

func TestInvalidForkCountParam(t *testing.T) {
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
		t.Run(testCase.name, func(t *testing.T) {
			err := validateForkCount(testCase.testValue)
			if testCase.expectedError == "" {
				assert.NoError(t, err)
			} else if assert.Error(t, err) {
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}

func newMavenIntegrationTestsUtilsBundle() mavenExecuteIntegrationTestUtilsBundle {
	utilsBundle := mavenExecuteIntegrationTestUtilsBundle{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utilsBundle
}
