package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMavenExecute(t *testing.T) {
	t.Run("mavenExecute should cleanup the parameters", func(t *testing.T) {
		mockRunner := mock.ExecMockRunner{}

		config := mavenExecuteOptions{
			Flags:   []string{"--errors --fail-fast "},
			Defines: []string{"  -DoutputFile=mvnDependencyTree.txt"},
			Goals:   []string{" dependency:tree"},
		}

		err := runMavenExecute(config, &mockRunner)

		expectedParams := []string{
			"--errors",
			"--fail-fast",
			"-DoutputFile=mvnDependencyTree.txt",
			"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
			"--batch-mode",
			"dependency:tree",
		}

		assert.NoError(t, err)
		assert.Equal(t, "mvn", mockRunner.Calls[0].Exec)
		assert.Equal(t, expectedParams, mockRunner.Calls[0].Params)
	})

}
