// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

func TestKarmaIntegration(t *testing.T) {

	//ToDo: implement networking, etc ...
	ctx := context.Background()

	dir, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	dir = filepath.Dir(dir)

	req := testcontainers.ContainerRequest{
		Image:      "node:latest",
		Cmd:        []string{"tail", "-f"},
		BindMounts: map[string]string{dir: "/data"},
	}

	nodeContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	piperOptions := []string{
		"karmaExecuteTests",
		"--help",
	}

	code, err := nodeContainer.Exec(ctx, append([]string{"/data/piper"}, piperOptions...))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
}
