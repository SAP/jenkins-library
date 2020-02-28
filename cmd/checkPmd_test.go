package cmd

import (
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
			MavenModulesExcludes: []string{"unit-tests", "integration-tests"},
		}
		expectedCall := mock.ExecCall{Exec: "mvn", Params: []string{"-Dpmd.excludes=*test*.java", "-Dpmd.rulesets=myRuleset.xml", "-pl", "!unit-tests", "-pl", "!integration-tests", "--batch-mode", "org.apache.maven.plugins:maven-pmd-plugin:3.13.0:pmd"}}

		err := runCheckPmd(&opts, nil, &execMockRunner)

		assert.Nil(t, err)
		assert.Equal(t, expectedCall, execMockRunner.Calls[0])
	})
}
