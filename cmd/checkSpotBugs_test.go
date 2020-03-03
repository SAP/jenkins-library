package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/mock"
)

func TestCheckSpotBugs(t *testing.T) {
	t.Run("mavenExecute should have include-, exclude filters and excluded maven modules", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		config := checkSpotBugsOptions{ExcludeFilterFile: "excludeFilter.xml", IncludeFilterFile: "includeFilter.xml", MavenModulesExcludes: []string{"my-tests"}}

		err := runCheckSpotBugs(&config, nil, &execMockRunner)

		assert.Nil(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"-Dspotbugs.includeFilterFile=includeFilter.xml", "-Dspotbugs.excludeFilterFile=excludeFilter.xml", "-pl", "!my-tests", "--batch-mode", "com.github.spotbugs:spotbugs-maven-plugin:3.1.12:spotbugs"}}, execMockRunner.Calls[0])
	})

	t.Run("mavenExecute should have standard excluded maven test modules", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		config := checkSpotBugsOptions{}
		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatal("Could not get current working directory")
		}
		defer os.Chdir(currentDir)
		os.Chdir("../test/resources/maven/")

		err = runCheckSpotBugs(&config, nil, &execMockRunner)

		assert.Nil(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"-pl", "!unit-tests", "-pl", "!integration-tests", "--batch-mode", "com.github.spotbugs:spotbugs-maven-plugin:3.1.12:spotbugs"}}, execMockRunner.Calls[0])
	})
}
