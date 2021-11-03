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

func TestDummy(t *testing.T) {

	t.Skip("Skipping testing - this is just to show how it can be done")
	ctx := context.Background()

	dir, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	dir = filepath.Dir(dir)

	req := testcontainers.ContainerRequest{
		Image:      "node:lts-stretch",
		Cmd:        []string{"tail", "-f"},
		BindMounts: map[string]string{dir: "/data"},
		//ToDo: we may set up a tmp directory and mount it in addition, e.g. for runtime artifacts ...
	}

	testContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	piperOptions := []string{
		"<piperStep>",
		"-- <piperFlag1>",
		"<piperFlag1Value>",
		"...",
		"--noTelemetry",
	}

	code, err := testContainer.Exec(ctx, append([]string{"/data/piper"}, piperOptions...))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
}
