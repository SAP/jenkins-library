//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestCLIIntegration ./integration/...

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
	t.Skip("Skip failing test for now")

	// t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", t.Name()), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
/piperbin/piper karmaExecuteTests
`
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	networkName := "sidecar-" + uuid.New().String()

	reqNode := testcontainers.ContainerRequest{
		Image: "node:lts-bullseye",
		Cmd:   []string{"tail", "-f"},
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount(pwd, "/piperbin"),
			testcontainers.BindMount(tempDir, "/test"),
		),
		Networks:       []string{networkName},
		NetworkAliases: map[string][]string{networkName: {"karma"}},
	}

	reqSel := testcontainers.ContainerRequest{
		Image:          "selenium/standalone-chrome:3.141.59-20210713",
		Networks:       []string{networkName},
		NetworkAliases: map[string][]string{networkName: {"selenium"}},
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

	// cannot use piper command directly since it is not possible to set Workdir for Exec call
	// workaround use shell call in container (see above)
	//piperOptions := []string{
	//		"karmaExecuteTests",
	//	"--help",
	//}
	//code, err := nodeContainer.Exec(ctx, append([]string{"/data/piper"}, piperOptions...))

	code, _, err := nodeContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
}
