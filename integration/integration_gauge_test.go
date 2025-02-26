//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestGaugeIntegration ./integration/...

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

const (
	installCommand string = "npm install -g @getgauge/cli --prefix=~/.npm-global --unsafe-perm" //option --unsafe-perm need to install gauge in docker container. See this issue: https://github.com/getgauge/gauge/issues/1470
)

func runTest(t *testing.T, languageRunner string) {
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestGaugeIntegration", "gauge-"+languageRunner), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script until it is possible to set workdir for Exec call
	testScript := fmt.Sprintf(`#!/bin/sh
cd /test
/piperbin/piper gaugeExecuteTests --installCommand="%v" --languageRunner=%v --runCommand="run" >test-log.txt 2>&1
`, installCommand, languageRunner)

	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "getgauge/gocd-jdk-mvn-node",
		Cmd:   []string{"tail", "-f"},
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount(pwd, "/piperbin"),
			testcontainers.BindMount(tempDir, "/test"),
		),
	}

	if languageRunner == "js" {
		reqNode.Image = "node:lts"
	}

	nodeContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})
	require.NoError(t, err)

	code, _, err := nodeContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	t.Cleanup(func() {
		// Remove files that are created by the container. t.TempDir() will
		// fail to remove them since it does not have the root permission
		_, _, err := nodeContainer.Exec(ctx, []string{"sh", "-c", "find /test -name . -o -prune -exec rm -rf -- {} +"})
		assert.NoError(t, err)

		assert.NoError(t, nodeContainer.Terminate(ctx))
	})

	content, err := os.ReadFile(filepath.Join(tempDir, "/test-log.txt"))
	if err != nil {
		t.Fatal("Could not read test-log.txt.", err)
	}
	output := string(content)
	assert.Contains(t, output, "info  gaugeExecuteTests - Scenarios:	2 executed	2 passed	0 failed	0 skipped")
	assert.Contains(t, output, "info  gaugeExecuteTests - SUCCESS")
}

func TestGaugeIntegrationJava(t *testing.T) {
	// t.Parallel()
	runTest(t, "java")
}

func TestGaugeIntegrationJS(t *testing.T) {
	// t.Parallel()
	runTest(t, "js")
}
