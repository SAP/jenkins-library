package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKubernetesDeployCommand(t *testing.T) {
	t.Parallel()

	testCmd := KubernetesDeployCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "kubernetesDeploy", testCmd.Use, "command name incorrect")

}
