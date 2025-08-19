//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestPNPMIntegration ./integration/...

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

func TestPNPMIntegration(t *testing.T) {
	testCases := []struct {
		name             string
		testDataDir      string
		piperArgs        string
		expectedContains []string
		expectedExitCode int
	}{
		{
			name:             "Basic Install",
			testDataDir:      "install",
			piperArgs:        "",
			expectedContains: []string{"info  npmExecuteScripts - SUCCESS"},
			expectedExitCode: 0,
		},
		{
			name:        "Specific Version",
			testDataDir: "install",
			piperArgs:   "--pnpmVersion=10.0.0",
			expectedContains: []string{
				"info  npmExecuteScripts - SUCCESS",
				"Using locally installed pnpm version 10.0.0",
			},
			expectedExitCode: 0,
		},
		{
			name:        "With Scripts",
			testDataDir: "scripts",
			piperArgs:   "--runScripts=test,build",
			expectedContains: []string{
				"info  npmExecuteScripts - SUCCESS",
				"Running tests",
				"Building project",
			},
			expectedExitCode: 0,
		},
		{
			name:        "Multi Module",
			testDataDir: "multi-module",
			piperArgs:   "--runScripts=test",
			expectedContains: []string{
				"info  npmExecuteScripts - SUCCESS",
				"Root test",
				"Module1 test",
				"Module2 test",
			},
			expectedExitCode: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			pwd, err := os.Getwd()
			assert.NoError(t, err, "Getting current working directory failed.")
			pwd = filepath.Dir(pwd)

			tempDir, err := createTmpDir(t)
			assert.NoError(t, err, "Error when creating temp dir")

			err = copyDir(filepath.Join(pwd, "integration", "testdata", "TestPnpmIntegration", tc.testDataDir), tempDir)
			if err != nil {
				t.Fatal("Failed to copy test project.")
			}

			testScript := fmt.Sprintf(`#!/bin/sh
cd /test
/piperbin/piper npmExecuteScripts %s >test-log.txt 2>&1
`, tc.piperArgs)
			err = os.WriteFile(filepath.Join(tempDir, "runPiper.sh"), []byte(testScript), 0700)
			assert.NoError(t, err, "Failed to write runPiper.sh")

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
			assert.Equal(t, tc.expectedExitCode, code)

			content, err := os.ReadFile(filepath.Join(tempDir, "/test-log.txt"))
			if err != nil {
				t.Fatal("Could not read test-log.txt.", err)
			}
			output := string(content)

			for _, expected := range tc.expectedContains {
				assert.Contains(t, output, expected)
			}
		})
	}
}
