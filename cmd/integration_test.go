// +build integration

package cmd

import (
	//"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	//"github.com/testcontainers/testcontainers-go/wait"
)

func TestKarmaIntegration(t *testing.T) {
	//assert.Equal(t, "test", "it")

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image: "alpine",
		/*WaitingFor: wait.ForAll(
			wait.ForLog("command override!"),
		),*/
		//Cmd: []string{"echo", "command override!"},
		//Cmd:          []string{"tail", "-f"},
		//VolumeMounts: map[string]string{},
	}

	alpine, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	code, err := alpine.Exec(ctx, []string{"ls", "-la"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	logs, err := alpine.Logs(ctx)
	defer logs.Close()
	res, err := ioutil.ReadAll(logs)

	//t.Logf("%v", alpine.)

	// ...so we convert it to a string by passing it through
	// a buffer first. A 'costly' but useful process.
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(logs)
	//res := buf.String()

	assert.Equal(t, "command override!", string(res))

	if err != nil {
		t.Error(err)
	}
	defer alpine.Terminate(ctx)
}
