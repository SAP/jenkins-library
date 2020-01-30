// +build integration

package integration

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	//"github.com/testcontainers/testcontainers-go/wait"
)

func TestPiperHelp(t *testing.T) {
	piperHelpCmd := command.Command{}

	var commandOutput bytes.Buffer
	piperHelpCmd.Stdout(&commandOutput)

	err := piperHelpCmd.RunExecutable(getPiperExecutable(), "--help")

	assert.NoError(t, err, "Calling piper --help failed")
	assert.Contains(t, commandOutput.String(), "Use \"piper [command] --help\" for more information about a command.")
}

func TestKarmaIntegration(t *testing.T) {
	//assert.Equal(t, "test", "it")

	ctx := context.Background()

	//piperBinAbsPath, err := filepath.Abs("../piper")

	req := testcontainers.ContainerRequest{
		Image: "node:latest",
		/*WaitingFor: wait.ForAll(
			wait.ForLog("Use \"piper [command] --help\" for more information about a command."),
		),*/
		//Cmd: []string{"echo", "command override!"},
		Cmd: []string{"tail", "-f"},
		//Cmd:        []string{"/data/piper", "--help"},
		BindMounts: map[string]string{"c:/Users/d032835/IdeaProjects/jenkins-library-os": "/data"},
	}

	alpine, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	code, err := alpine.Exec(ctx, []string{"/data/piper", "--help"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	//ToDo: check how to wait for complete stream to be available ...

	//logs, err := alpine.Logs(ctx)
	//defer logs.Close()
	//res, err := ioutil.ReadAll(logs)
	//assert.NoError(t, err)
	//assert.Equal(t, "Use \"piper [command] --help\" for more information about a command.", string(res))

	defer alpine.Terminate(ctx)
}

func getPiperExecutable() string {
	if p := os.Getenv("PIPER_TEST_EXECUTABLE"); len(p) > 0 {
		return p
	}
	return "piper"
}
