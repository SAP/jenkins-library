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

// In this test the piper command golangBuild performs testing, BOM file creation and building a project with entry point in the cmd/server/server.go
// The configuration for golangBuild can be found in testdata/TestGolangIntegration/golang-project1/.pipeline/config.yml
func TestGolangBuild_Project1(t *testing.T) {
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir("")
	defer os.RemoveAll(tempDir) // clean up
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestGolangIntegration", "golang-project1"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := fmt.Sprintf(`#!/bin/sh
cd /test
/piperbin/piper golangBuild >test-log.txt 2>&1
`)
	ioutil.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "golang:1",
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
	assert.Contains(t, output, "info  golangBuild - running command: go install gotest.tools/gotestsum@latest")
	assert.Contains(t, output, "info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest")
	assert.Contains(t, output, "info  golangBuild - running command: gotestsum --junitfile TEST-go.xml -- -coverprofile=cover.out ./...")
	assert.Contains(t, output, "info  golangBuild - DONE 8 tests")
	assert.Contains(t, output, "info  golangBuild - running command: go tool cover -html cover.out -o coverage.html")
	assert.Contains(t, output, "info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml -- -tags=integration ./...")
	assert.Contains(t, output, "info  golangBuild - running command: cyclonedx-gomod mod -licenses -test -output bom.xml")
	assert.Contains(t, output, "info  golangBuild - running command: go build -trimpath -o golang-app-linux.amd64 cmd/server/server.go")
	assert.Contains(t, output, "info  golangBuild - SUCCESS")

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript = fmt.Sprintf(`#!/bin/sh
cd /test
ls -l >files-list.txt 2>&1
`)
	ioutil.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	code, err = nodeContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	content, err = ioutil.ReadFile(filepath.Join(tempDir, "/files-list.txt"))
	if err != nil {
		t.Fatal("Could not read files-list.txt.", err)
	}
	output = string(content)
	assert.Contains(t, output, "TEST-go.xml")
	assert.Contains(t, output, "TEST-integration.xml")
	assert.Contains(t, output, "bom.xml")
	assert.Contains(t, output, "cover.out")
	assert.Contains(t, output, "coverage.html")
	assert.Contains(t, output, "golang-app-linux.amd64")
}

// In this test, the piper golangBuild command only builds the project with the entry point at the project root.
// The configuration for golangBuild can be found in testdata/TestGolangIntegration/golang-project2/.pipeline/config.yml
func TestGolangBuild_Project2(t *testing.T) {
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir("")
	defer os.RemoveAll(tempDir) // clean up
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestGolangIntegration", "golang-project2"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := fmt.Sprintf(`#!/bin/sh
cd /test
/piperbin/piper golangBuild >test-log.txt 2>&1
`)
	ioutil.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "golang:1",
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
	assert.NotContains(t, output, "info  golangBuild - running command: go install gotest.tools/gotestsum@latest")
	assert.NotContains(t, output, "info  golangBuild - running command: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest")
	assert.NotContains(t, output, "info  golangBuild - running command: gotestsum --junitfile TEST-go.xml -- -coverprofile=cover.out ./...")
	assert.NotContains(t, output, "info  golangBuild - running command: go tool cover -html cover.out -o coverage.html")
	assert.NotContains(t, output, "info  golangBuild - running command: gotestsum --junitfile TEST-integration.xml -- -tags=integration ./...")
	assert.NotContains(t, output, "info  golangBuild - running command: cyclonedx-gomod mod -licenses -test -output bom.xml")
	assert.Contains(t, output, "info  golangBuild - running command: go build -trimpath -o golang-app-linux.amd64")
	assert.Contains(t, output, "info  golangBuild - SUCCESS")
}
