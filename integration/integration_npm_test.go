//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestNPMIntegration ./integration/...

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

func TestNPMIntegrationRunScriptsWithOptions(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestNpmIntegration", "runScriptsWithOptions"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts --runScripts=start --scriptOptions=--tag,tag1 >test-log-runScriptWithOptions.txt 2>&1
`
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:12-slim",
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

	content, err := os.ReadFile(filepath.Join(tempDir, "/test-log-runScriptWithOptions.txt"))
	if err != nil {
		t.Fatal("Could not read test-log-runScriptWithOptions.txt.", err)
	}
	output := string(content)
	assert.Contains(t, output, "info  npmExecuteScripts - running command: npm run start -- --tag tag1")
	assert.Contains(t, output, "info  npmExecuteScripts - [ '--tag', 'tag1' ]")
}

func TestNPMIntegrationRegistrySetInFlags(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestNpmIntegration", "registrySetInFlags"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts --install --runScripts=ci-build --defaultNpmRegistry=https://foo.bar >test-log-registrySetInFlags.txt 2>&1
`
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:12-slim",
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

	content, err := os.ReadFile(filepath.Join(tempDir, "/test-log-registrySetInFlags.txt"))
	if err != nil {
		t.Fatal("Could not read test-log-registrySetInFlags.txt.", err)
	}
	output := string(content)
	assert.Contains(t, output, "info  npmExecuteScripts - https://foo.bar")
}

func TestNPMIntegrationRegistrySetInNpmrc(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestNpmIntegration", "registrySetInNpmrc"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts --install --runScripts=ci-build >test-log-registrySetInNpmrc.txt 2>&1
`
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:12-slim",
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

	content, err := os.ReadFile(filepath.Join(tempDir, "/test-log-registrySetInNpmrc.txt"))
	if err != nil {
		t.Fatal("Could not read test-log-registrySetInNpmrc.txt.", err)
	}
	output := string(content)
	assert.Contains(t, output, "info  npmExecuteScripts - https://example.com")
}

func TestNPMIntegrationRegistryWithTwoModules(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestNpmIntegration", "registryWithTwoModules"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts --install --runScripts=ci-build --defaultNpmRegistry=https://foo.bar >test-log-registryWithTwoModules.txt 2>&1
`
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:12-slim",
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

	content, err := os.ReadFile(filepath.Join(tempDir, "/test-log-registryWithTwoModules.txt"))
	if err != nil {
		t.Fatal("Could not read test-log-registryWithTwoModules.txt.", err)
	}
	output := string(content)
	assert.Contains(t, output, "info  npmExecuteScripts - https://example.com")
	assert.Contains(t, output, "info  npmExecuteScripts - https://foo.bar")
}

func TestPnpm(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir(t)
	defer os.RemoveAll(tempDir) // clean up
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestNpmIntegration", "runPnpm"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts --install --runScripts=start --defaultNpmRegistry=https://foo.bar >test-log.txt 2>&1
`
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:14-slim",
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

	code, _, err := nodeContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	content, err := os.ReadFile(filepath.Join(tempDir, "/test-log.txt"))
	if err != nil {
		t.Fatal("Could not read test-log.txt.", err)
	}
	output := string(content)
	assert.Contains(t, output, "info  npmExecuteScripts - running command: npm install -g pnpm")
	assert.Contains(t, output, "info  npmExecuteScripts - added 1 package in")
	assert.Contains(t, output, "running command: pnpm config get registry")
	assert.Contains(t, output, "info  npmExecuteScripts - running command: pnpm config set registry https://foo.bar")
	assert.Contains(t, output, "info  npmExecuteScripts - running command: pnpm install")
	assert.Contains(t, output, "info  npmExecuteScripts - running command: npm run start")
}
