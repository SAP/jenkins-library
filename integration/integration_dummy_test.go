//go:build integration

// can be executed with
// go test -v -tags integration -run TestAPICLIIntegration ./integration/...

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func TestDummyIntegration(t *testing.T) {
	t.Skip("Skipping testing - this is just to show how it can be done")
	ctx := context.Background()

	dir, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	dir = filepath.Dir(dir)

	req := testcontainers.ContainerRequest{
		Image: "node:lts-bookworm",
		Cmd:   []string{"tail", "-f"},
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount(dir, "/data"),
		),
		//ToDo: we may set up a tmp directory and mount it in addition, e.g. for runtime artifacts ...
	}

	testContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	piperOptions := []string{
		"<piperStep>",
		"-- <piperFlag1>",
		"<piperFlag1Value>",
		"...",
		"--noTelemetry",
	}

	code, _, err := testContainer.Exec(ctx, append([]string{"/data/piper"}, piperOptions...))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
}
