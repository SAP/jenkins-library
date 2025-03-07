//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestGradleIntegration ./integration/...

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

func TestGradleIntegrationExecuteBuildJavaProjectBOMCreationUsingWrapper(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestGradleIntegration", "java-project"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := fmt.Sprintf(`#!/bin/sh

	# Defining variables
	max_retries=2
	retry_count=0
	log_file="test-log.txt"
	piper_command="/piperbin/piper gradleExecuteBuild"

	# Function to execute piper command with retries
	execute_command() {
	  while [ $retry_count -lt $max_retries ]; do
	    echo "Attempt $(($retry_count + 1)) of $max_retries: Executing piper command..."
	    $piper_command >$1 2>&1
	
	    if [ -f "$1" ]; then
	      echo "File with test results created successfully: $1"
	      return 0
	    else
	      echo "Error: File not found after running tests - $1"
	      retry_count=$(($retry_count + 1))
	      echo "Waiting for 2 seconds before retry..."
	      sleep 2
	    fi
	  done
	
	  echo "Failed to create file after $max_retries attempts."
	  return 1
	}
	cd /test

	# Execute command with retry logic
	execute_command $log_file
	`)
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "adoptopenjdk/openjdk11:jdk-11.0.11_9-alpine",
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
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: ./gradlew tasks")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: ./gradlew cyclonedxBom --init-script initScript.gradle.tmp")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: ./gradlew build")
	assert.Contains(t, output, "info  gradleExecuteBuild - BUILD SUCCESSFUL")
	assert.Contains(t, output, "info  gradleExecuteBuild - SUCCESS")

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript = fmt.Sprintf(`#!/bin/sh
cd /test
ls -l ./build/reports/ >files-list.txt 2>&1
`)
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	code, _, err = nodeContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	content, err = os.ReadFile(filepath.Join(tempDir, "/files-list.txt"))
	if err != nil {
		t.Fatal("Could not read files-list.txt.", err)
	}
	output = string(content)
	assert.Contains(t, output, "bom-gradle.xml")
}

func TestGradleIntegrationExecuteBuildJavaProjectWithBomPlugin(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	// using custom createTmpDir function to avoid issues with symlinks on Docker for Mac
	tempDir, err := createTmpDir(t)
	assert.NoError(t, err, "Error when creating temp dir")

	err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestGradleIntegration", "java-project-with-bom-plugin"), tempDir)
	if err != nil {
		t.Fatal("Failed to copy test project.")
	}

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript := fmt.Sprintf(`#!/bin/sh
cd /test
/piperbin/piper gradleExecuteBuild >test-log.txt 2>&1
`)
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	reqNode := testcontainers.ContainerRequest{
		Image: "gradle:6-jdk11-alpine",
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
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: gradle tasks")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: gradle cyclonedxBom")
	assert.Contains(t, output, "info  gradleExecuteBuild - running command: gradle build")
	assert.Contains(t, output, "info  gradleExecuteBuild - BUILD SUCCESSFUL")
	assert.Contains(t, output, "info  gradleExecuteBuild - SUCCESS")

	//workaround to use test script util it is possible to set workdir for Exec call
	testScript = fmt.Sprintf(`#!/bin/sh
cd /test
ls -l ./build/reports/ >files-list.txt 2>&1
`)
	os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)

	code, _, err = nodeContainer.Exec(ctx, []string{"sh", "/test/runPiper.sh"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	content, err = os.ReadFile(filepath.Join(tempDir, "/files-list.txt"))
	if err != nil {
		t.Fatal("Could not read files-list.txt.", err)
	}
	output = string(content)
	assert.Contains(t, output, "bom-gradle.xml")
}
