//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestPythonIntegration ./integration/...

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

func TestPythonIntegrationBuildProject(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestPythonIntegration", "python-project"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := fmt.Sprintf(`#!/bin/sh
		cd /test
		/piperbin/piper pythonBuild >test-log.txt 2>&1`)
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "python:3.10",
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

	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/python setup.py sdist bdist_wheel")
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/pip install --upgrade --root-user-action=ignore cyclonedx-bom==")
	assert.Contains(t, output, "info  pythonBuild - running command: piperBuild-env/bin/cyclonedx-py env --output-file bom-pip.xml")
	assert.Contains(t, output, "info  pythonBuild - SUCCESS")

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript = fmt.Sprintf(`#!/bin/sh
		cd /test
		ls -l . dist build >files-list.txt 2>&1`)
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	code, _, err = nodeContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	content, err = os.ReadFile(filepath.Join(tempDir, "/files-list.txt"))
	if err != nil {
		t.Fatal("Could not read files-list.txt.", err)
	}
	output = string(content)
	assert.Contains(t, output, "bom-pip.xml")
	assert.Contains(t, output, "example_pkg-0.0.1.tar.gz")
	assert.Contains(t, output, "example_pkg-0.0.1-py3-none-any.whl")
}
