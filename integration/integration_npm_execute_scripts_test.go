// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

func TestRunScriptsWithOptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir("")
	defer os.RemoveAll(tempDir) // clean up
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestNpmIntegration", "runScriptsWithOptions"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts --runScripts=start --scriptOptions=--tag,tag1 >test-log.txt 2>&1
`
	ioutil.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:12-slim",
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
	assert.Contains(t, output, "info  npmExecuteScripts - running command: npm run start -- --tag tag1")
	assert.Contains(t, output, "info  npmExecuteScripts - [ '--tag', 'tag1' ]")
}

func TestRegistrySetInFlags(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir("")
	defer os.RemoveAll(tempDir) // clean up
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestNpmIntegration", "registrySetInFlags"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts --install --runScripts=ci-build,ci-backend-unit-test --defaultNpmRegistry=https://foo.bar >test-log.txt 2>&1
`
	ioutil.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:12-slim",
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
	assert.Contains(t, output, "info  npmExecuteScripts - https://foo.bar")
}

func TestRegistrySetInNpmrc(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir("")
	defer os.RemoveAll(tempDir) // clean up
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestNpmIntegration", "registrySetInNpmrc"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts --install --runScripts=ci-build,ci-backend-unit-test >test-log.txt 2>&1
`
	ioutil.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:12-slim",
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
	assert.Contains(t, output, "info  npmExecuteScripts - https://example.com")
}

func TestRegistryWithTwoModules(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir("")
	defer os.RemoveAll(tempDir) // clean up
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestNpmIntegration", "registryWithTwoModules"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := `#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts --install --runScripts=ci-build,ci-backend-unit-test --defaultNpmRegistry=https://foo.bar >test-log.txt 2>&1
`
	ioutil.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "node:12-slim",
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
	assert.Contains(t, output, "info  npmExecuteScripts - https://example.com")
	assert.Contains(t, output, "info  npmExecuteScripts - https://foo.bar")
}
