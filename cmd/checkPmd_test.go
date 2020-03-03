package cmd

import (
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCheckPmd(t *testing.T) {
	t.Run("should execute with excluded files, rulesets defined and modules excluded", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		opts := checkPmdOptions{
			Excludes:             []string{"*test*.java"},
			RuleSets:             []string{"myRuleset.xml"},
			MavenModulesExcludes: []string{"my-tests"},
		}
		expectedCall := mock.ExecCall{Exec: "mvn", Params: []string{"-Dpmd.excludes=*test*.java", "-Dpmd.rulesets=myRuleset.xml", "-pl", "!my-tests", "--batch-mode", "org.apache.maven.plugins:maven-pmd-plugin:3.13.0:pmd"}}

		err := runCheckPmd(&opts, nil, &execMockRunner)

		assert.Nil(t, err)
		assert.Equal(t, expectedCall, execMockRunner.Calls[0])
	})
	t.Run("should execute with standard excluded maven test modules", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		config := checkPmdOptions{}
		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatal("Could not get current working directory")
		}
		defer os.Chdir(currentDir)
		os.Chdir("../test/resources/maven/")

		err = runCheckPmd(&config, nil, &execMockRunner)

		assert.Nil(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"-pl", "!unit-tests", "-pl", "!integration-tests", "--batch-mode", "org.apache.maven.plugins:maven-pmd-plugin:3.13.0:pmd"}}, execMockRunner.Calls[0])
	})
}
