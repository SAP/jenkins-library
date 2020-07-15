// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

func TestFirstChangeMe(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir("")
	defer os.RemoveAll(tempDir) // clean up
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestCfDeployIntegration"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	username := os.Getenv("PIPER_INTEGRATION_CF_USERNAME")
	if len(username) == 0 {
		t.Fatal("Username for SAP Cloud Platform required")
	}
	password := os.Getenv("PIPER_INTEGRATION_CF_PASSWORD")
	if len(username) == 0 {
		t.Fatal("Password for SAP Cloud Platform required")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := fmt.Sprintf(`#!/bin/sh
cd /app
/piperbin/piper cloudFoundryDeploy --username %s --password '%s' >/test/test-log.txt 2>&1
`, username, password)
	ioutil.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	// Prepared docker image with deployable artifact, sources are at:
	// https://github.com/piper-validation/cloud-s4-sdk-book/tree/consumer-test
	reqNode := testcontainers.ContainerRequest{
		Image: "fwilhe/cf-it",
		Cmd:   []string{"tail", "-f"},
		BindMounts: map[string]string{
			pwd:     "/piperbin",
			tempDir: "/test",
		},
	}

	nodeContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqNode,
		Started:          true,
	})

	code, err := nodeContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	content, err := ioutil.ReadFile(filepath.Join(tempDir, "/test-log.txt"))
	if err != nil {
		t.Fatal("Could not read test-log.txt.", err)
	}
	output := string(content)
	assert.Contains(t, output, "info  cloudFoundryDeploy - name:              devops-docker-images-IT")
	assert.Contains(t, output, "info  cloudFoundryDeploy - Logged out successfully")
}
