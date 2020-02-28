package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/command"
)

func TestCheckPmd(t *testing.T) {
	t.Run("should execute with modules excluded", func(t *testing.T) {
		opts := checkPmdOptions{
			Excludes:             nil,
			RuleSets:             nil,
			MavenModulesExcludes: []string{"unit-tests", "integration-tests"},
		}

		c := command.Command{}
		runCheckPmd(&opts, nil, &c)
	})
}
