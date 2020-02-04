// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

func TestKarmaIntegration(t *testing.T) {

	ctx := context.Background()

	dir, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	dir = filepath.Dir(dir)

	networkName := "sidecar-" + uuid.New().String()

	reqNode := testcontainers.ContainerRequest{
		Image:      "node:latest",
		Cmd:        []string{"tail", "-f"},
		BindMounts: map[string]string{dir: "/data"},
		Networks: []string{networkName},
		NetworkAliases: map[string][]string{networkName: []string{"node"}},
	}

	reqSel := testcontainers.ContainerRequest{
		Image:      "selenium/standalone-chrome",
		Networks: []string{networkName},
		NetworkAliases: map[string][]string{networkName: []string{"selenium"}},
	}

	provider, err := testcontainers.ProviderDocker.GetProvider()
	assert.NoError(t, err)

	network, err := provider.CreateNetwork(ctx, testcontainers.NetworkRequest{Name: networkName, CheckDuplicate: true})
	if err != nil {
		t.Fatal(err)
	}
	defer network.Remove(ctx)

	nodeContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer nodeContainer.Terminate(ctx)

	selContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqSel,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer selContainer.Terminate(ctx)

	piperOptions := []string{
		"karmaExecuteTests",
		"--help",
	}

	code, err := nodeContainer.Exec(ctx, append([]string{"/data/piper"}, piperOptions...))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
}
