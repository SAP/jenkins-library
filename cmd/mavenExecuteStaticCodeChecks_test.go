package cmd

import (
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"

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
			PmdExcludes:               []string{"*test.java", "*prod.java"},
			PmdRuleSets:               []string{"myRule.xml", "anotherRule.xml"},
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
				"-Dpmd.excludes=*test.java,*prod.java",
				"-Dpmd.rulesets=myRule.xml,anotherRule.xml",
				"--batch-mode",
				"com.github.spotbugs:spotbugs-maven-plugin:3.1.12:spotbugs",
				"org.apache.maven.plugins:maven-pmd-plugin:3.13.0:pmd",
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
	t.Run("should log fatal if all tools are turned off", func(t *testing.T) {
		var hasFailed bool
		log.Entry().Logger.ExitFunc = func(int) { hasFailed = true }
		execMockRunner := mock.ExecMockRunner{}
		config := mavenExecuteStaticCodeChecksOptions{
			SpotBugs: false,
			Pmd:      false,
		}
		_ = runMavenStaticCodeChecks(&config, nil, &execMockRunner)
		assert.True(t, hasFailed, "expected command to exit with fatal")
	})
}

func TestGetPmdMavenParameters(t *testing.T) {
	t.Run("should return maven options with excludes and rulesets", func(t *testing.T) {
		config := mavenExecuteStaticCodeChecksOptions{
			Pmd:         true,
			PmdExcludes: []string{"*test.java", "*prod.java"},
			PmdRuleSets: []string{"myRule.xml", "anotherRule.xml"},
		}
		expected := maven.ExecuteOptions{
			Goals:   []string{"org.apache.maven.plugins:maven-pmd-plugin:3.13.0:pmd"},
			Defines: []string{"-Dpmd.excludes=*test.java,*prod.java", "-Dpmd.rulesets=myRule.xml,anotherRule.xml"},
		}

		assert.Equal(t, &expected, getPmdMavenParameters(&config))
	})
	t.Run("should return maven goal only", func(t *testing.T) {
		config := mavenExecuteStaticCodeChecksOptions{}
		expected := maven.ExecuteOptions{
			Goals: []string{"org.apache.maven.plugins:maven-pmd-plugin:3.13.0:pmd"}}

		assert.Equal(t, &expected, getPmdMavenParameters(&config))
	})
}

func TestGetSpotBugsMavenParameters(t *testing.T) {
	t.Run("should return maven options with excludes and include filters", func(t *testing.T) {
		config := mavenExecuteStaticCodeChecksOptions{
			SpotBugs:                  true,
			SpotBugsExcludeFilterFile: "excludeFilter.xml",
			SpotBugsIncludeFilterFile: "includeFilter.xml",
		}
		expected := maven.ExecuteOptions{
			Goals:   []string{"com.github.spotbugs:spotbugs-maven-plugin:3.1.12:spotbugs"},
			Defines: []string{"-Dspotbugs.includeFilterFile=includeFilter.xml", "-Dspotbugs.excludeFilterFile=excludeFilter.xml"},
		}

		assert.Equal(t, &expected, getSpotBugsMavenParameters(&config))
	})
	t.Run("should return maven goal only", func(t *testing.T) {
		config := mavenExecuteStaticCodeChecksOptions{}
		expected := maven.ExecuteOptions{
			Goals: []string{"com.github.spotbugs:spotbugs-maven-plugin:3.1.12:spotbugs"}}

		assert.Equal(t, &expected, getSpotBugsMavenParameters(&config))
	})
}
