package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/mock"
)

func TestCheckSpotBugs(t *testing.T) {
	t.Run("mavenExecute should have include and exclude filters", func(t *testing.T) {
		execMockRunner := mock.ExecMockRunner{}
		config := checkSpotBugsOptions{ExcludeFilterFile: "excludeFilter.xml", IncludeFilterFile: "includeFilter.xml"}

		err := runCheckSpotBugs(&config, nil, &execMockRunner)

		assert.Nil(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "mvn", Params: []string{"-Dspotbugs.includeFilterFile=includeFilter.xml", "-Dspotbugs.excludeFilterFile=excludeFilter.xml", "--batch-mode", "com.github.spotbugs:spotbugs-maven-plugin:3.1.12:spotbugs"}}, execMockRunner.Calls[0])
	})
}
