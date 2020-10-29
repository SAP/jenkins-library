package cmd

import (
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/stretchr/testify/assert"
)

func TestRunMavenStaticCodeChecks(t *testing.T) {
	t.Run("should run spotBugs and pmd with all configured options", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		config := mavenExecuteStaticCodeChecksOptions{
			SpotBugs:                  true,
			Pmd:                       true,
			PmdMaxAllowedViolations:   10,
			PmdFailurePriority:        2,
			SpotBugsExcludeFilterFile: "excludeFilter.xml",
			SpotBugsIncludeFilterFile: "includeFilter.xml",
			MavenModulesExcludes:      []string{"testing-lib", "test-helpers"},
		}
		expected := mock.ExecCall{
			Exec: "mvn",
			Params: []string{"-pl", "!unit-tests", "-pl", "!integration-tests",
				"-pl", "!testing-lib", "-pl", "!test-helpers",
				"-Dspotbugs.includeFilterFile=includeFilter.xml",
				"-Dspotbugs.excludeFilterFile=excludeFilter.xml",
				"-Dpmd.maxAllowedViolations=10",
				"-Dpmd.failurePriority=2",
				"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
				"--batch-mode",
				"com.github.spotbugs:spotbugs-maven-plugin:4.1.4:check",
				"org.apache.maven.plugins:maven-pmd-plugin:3.13.0:check",
			},
		}

		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatal("Could not get current working directory")
		}
		defer os.Chdir(currentDir)
		os.Chdir("../test/resources/maven/")

		err = runMavenStaticCodeChecks(&config, nil, &execMockRunner)

		assert.Nil(t, err)
		assert.Equal(t, expected, execMockRunner.Calls[0])
	})
	t.Run("should warn and skip execution if all tools are turned off", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		config := mavenExecuteStaticCodeChecksOptions{
			SpotBugs: false,
			Pmd:      false,
		}
		err := runMavenStaticCodeChecks(&config, nil, &execMockRunner)
		assert.Nil(t, err)
		assert.Nil(t, execMockRunner.Calls)
	})
}

func TestGetPmdMavenParameters(t *testing.T) {
	t.Run("should return maven options with max allowed violations and failrure priority", func(t *testing.T) {
		config := mavenExecuteStaticCodeChecksOptions{
			Pmd:                     true,
			PmdFailurePriority:      2,
			PmdMaxAllowedViolations: 5,
		}
		expected := maven.ExecuteOptions{
			Goals:   []string{"org.apache.maven.plugins:maven-pmd-plugin:3.13.0:check"},
			Defines: []string{"-Dpmd.maxAllowedViolations=5", "-Dpmd.failurePriority=2"},
		}

		assert.Equal(t, &expected, getPmdMavenParameters(&config))
	})
	t.Run("should return maven options without failure priority if out of bounds", func(t *testing.T) {
		config := mavenExecuteStaticCodeChecksOptions{
			Pmd:                     true,
			PmdFailurePriority:      123,
			PmdMaxAllowedViolations: 5,
		}
		expected := maven.ExecuteOptions{
			Goals:   []string{"org.apache.maven.plugins:maven-pmd-plugin:3.13.0:check"},
			Defines: []string{"-Dpmd.maxAllowedViolations=5"},
		}

		assert.Equal(t, &expected, getPmdMavenParameters(&config))
	})
	t.Run("should return maven goal only", func(t *testing.T) {
		config := mavenExecuteStaticCodeChecksOptions{}
		expected := maven.ExecuteOptions{
			Goals: []string{"org.apache.maven.plugins:maven-pmd-plugin:3.13.0:check"}}

		assert.Equal(t, &expected, getPmdMavenParameters(&config))
	})
}

func TestGetSpotBugsMavenParameters(t *testing.T) {
	t.Run("should return maven options with excludes-, include filters and max allowed violations", func(t *testing.T) {
		config := mavenExecuteStaticCodeChecksOptions{
			SpotBugs:                     true,
			SpotBugsExcludeFilterFile:    "excludeFilter.xml",
			SpotBugsIncludeFilterFile:    "includeFilter.xml",
			SpotBugsMaxAllowedViolations: 123,
		}
		expected := maven.ExecuteOptions{
			Goals:   []string{"com.github.spotbugs:spotbugs-maven-plugin:4.1.4:check"},
			Defines: []string{"-Dspotbugs.includeFilterFile=includeFilter.xml", "-Dspotbugs.excludeFilterFile=excludeFilter.xml", "-Dspotbugs.maxAllowedViolations=123"},
		}

		assert.Equal(t, &expected, getSpotBugsMavenParameters(&config))
	})
	t.Run("should return maven goal only", func(t *testing.T) {
		config := mavenExecuteStaticCodeChecksOptions{}
		expected := maven.ExecuteOptions{
			Goals: []string{"com.github.spotbugs:spotbugs-maven-plugin:4.1.4:check"}}

		assert.Equal(t, &expected, getSpotBugsMavenParameters(&config))
	})
}
