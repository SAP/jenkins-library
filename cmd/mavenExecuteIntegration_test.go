package cmd

import (
	"fmt"
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

	config := mavenExecuteIntegrationOptions{
		Retry:     2,
		ForkCount: "1C",
	}

	utils.FilesMock.AddDir("integration-tests")
	utils.FilesMock.AddFile("integration-tests/pom.xml", []byte(`<project> </project>`))

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

func TestValidateForkCount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		testValue     string
		expectedError error
	}{
		{
			name:          "valid integer",
			testValue:     "2",
			expectedError: nil,
		},
		{
			name:          "zero is valid",
			testValue:     "0",
			expectedError: nil,
		},
		{
			name:          "valid floating point",
			testValue:     "2.5C",
			expectedError: nil,
		},
		{
			name:          "invalid floating point",
			testValue:     "1.2",
			expectedError: fmt.Errorf("invalid forkCount parameter: A non-integer value may only be provided with a 'C' suffix"),
		},
		{
			name:          "invalid",
			testValue:     "C1",
			expectedError: fmt.Errorf("invalid forkCount parameter"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateForkCount(testCase.testValue)
			if testCase.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, testCase.expectedError.Error())
			}
		})
	}
}

func newMavenIntegrationTestsUtilsBundle() mavenExecuteIntegrationTestUtilsBundle {
	utilsBundle := mavenExecuteIntegrationTestUtilsBundle{
		ExecMockRunner: &mock.ExecMockRunner{
			Dir:                 nil,
			Env:                 nil,
			ExitCode:            0,
			Calls:               nil,
			StdoutReturn:        nil,
			ShouldFailOnCommand: nil,
		},
		FilesMock: &mock.FilesMock{},
	}
	return utilsBundle
}
