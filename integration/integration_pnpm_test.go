//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestPNPMIntegration ./integration/...

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

func TestPNPMIntegrationInstall(t *testing.T) {
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestPnpmIntegration", "install"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	testScript := `#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts --install >test-log.txt 2>&1
`
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:20-slim",
		Cmd:   []string{"tail", "-f"},
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount(pwd, "/piperbin"),
			testcontainers.BindMount(tempDir, "/test"),
		),
	}

	nodeContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, _, err := nodeContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	content, err := os.ReadFile(filepath.Join(tempDir, "/test-log.txt"))
	if err != nil {
		t.Fatal("Could not read test-log.txt.", err)
	}
	output := string(content)

	assert.Contains(t, output, "info  npmExecuteScripts - SUCCESS")
}

func TestPNPMIntegrationBomGeneration(t *testing.T) {
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestPnpmIntegration", "bom"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	testScript := `#!/bin/sh
	cd /test
	apt-get update && apt-get install -y ca-certificates libicu72
	/piperbin/piper npmExecuteScripts --install --createBOM --verbose >test-log.txt 2>&1
ls -la >> test-log.txt 2>&1
pwd >> test-log.txt 2>&1
find / -name bom-npm.xml >> test-log.txt 2>&1
`
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:20-slim",
		Cmd:   []string{"tail", "-f"},
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount(pwd, "/piperbin"),
			testcontainers.BindMount(tempDir, "/test"),
			testcontainers.BindMount("/etc/ssl/certs", "/etc/ssl/certs"),
		),
	}

	nodeContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, _, err := nodeContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	content, err := os.ReadFile(filepath.Join(tempDir, "/test-log.txt"))
	if err != nil {
		t.Fatal("Could not read test-log.txt.", err)
	}
	output := string(content)

	t.Logf("Test output: %s", output)

	// Update assertions to match command output
	assert.Contains(t, output, "info  npmExecuteScripts - Creating CycloneDX")
	assert.FileExists(t, filepath.Join(tempDir, "bom-npm.xml"))
}

func TestPNPMIntegrationBomGenerationError(t *testing.T) {
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestPnpmIntegration", "bom-error"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	testScript := `#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts --install --createBOM >test-log.txt 2>&1
`
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:20-slim",
		Cmd:   []string{"tail", "-f"},
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount(pwd, "/piperbin"),
			testcontainers.BindMount(tempDir, "/test"),
		),
	}

	nodeContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, _, err := nodeContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})
	assert.Error(t, err)
	assert.NotEqual(t, 0, code)

	content, err := os.ReadFile(filepath.Join(tempDir, "/test-log.txt"))
	if err != nil {
		t.Fatal("Could not read test-log.txt.", err)
	}
	output := string(content)

	assert.Contains(t, output, "error  npmExecuteScripts - failed to generate CycloneDX BOM")
}
