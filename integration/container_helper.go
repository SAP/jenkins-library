//go:build integration
// +build integration

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
)

// ContainerConfig holds configuration for creating a test container
type ContainerConfig struct {
	Image    string // Docker image to use
	TestData string // Path relative to integration/testdata (e.g., "TestGradleIntegration/java-project")
	WorkDir  string // Working directory inside container (e.g., "/java-project")
	User     string // User to run as (optional, defaults to image default)
}

// StartPiperContainer starts a container with the piper binary mounted and test data copied.
// The container is automatically cleaned up when the test finishes via t.Cleanup.
func StartPiperContainer(t *testing.T, cfg ContainerConfig) testcontainers.Container {
	t.Helper()

	ctx := context.Background()
	projectRoot := getProjectRoot(t)

	req := testcontainers.ContainerRequest{
		Image: cfg.Image,
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(projectRoot, "integration", "testdata", cfg.TestData),
				ContainerFilePath: "/",
				FileMode:          0755,
			},
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{
				fmt.Sprintf("%s:/piperbin", projectRoot),
			}
		},
	}

	if cfg.User != "" {
		req.User = cfg.User
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "Failed to start container")

	// Register cleanup
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	})

	return container
}

// RunPiper executes a piper command in the container and returns the combined stdout/stderr output.
// It fails the test if the command returns a non-zero exit code or if there's an execution error.
func RunPiper(t *testing.T, container testcontainers.Container, workDir, command string, args ...string) string {
	t.Helper()

	ctx := context.Background()
	cmd := append([]string{"/piperbin/piper", command}, args...)

	code, reader, err := container.Exec(ctx, cmd, exec.WithWorkingDir(workDir))

	output, readErr := io.ReadAll(reader)
	outputStr := string(output)

	require.NoError(t, err, "Failed to execute piper command: %v\nCommand: %v\nWorkDir: %s", err, cmd, workDir)
	require.NoError(t, readErr, "Failed to read command output for: %v", cmd)

	require.Equal(t, 0, code,
		"Piper command failed with exit code %d\nCommand: %v\nWorkDir: %s\nOutput:\n%s",
		code, cmd, workDir, outputStr)

	return outputStr
}

// RunPiperExpectFailure executes a piper command expecting it to fail.
// It returns the exit code and output. Use this for negative test cases.
func RunPiperExpectFailure(t *testing.T, container testcontainers.Container, workDir, command string, args ...string) (int, string) {
	t.Helper()

	ctx := context.Background()
	cmd := append([]string{"/piperbin/piper", command}, args...)

	code, reader, err := container.Exec(ctx, cmd, exec.WithWorkingDir(workDir))

	output, readErr := io.ReadAll(reader)
	outputStr := string(output)

	require.NoError(t, err,
		"Failed to execute piper command: %v\nCommand: %v\nWorkDir: %s",
		err, cmd, workDir)
	require.NoError(t, readErr,
		"Failed to read command output for: %v", cmd)

	if code == 0 {
		t.Logf("WARNING: Command succeeded with exit code 0 (expected failure)\nOutput:\n%s", outputStr)
	} else {
		t.Logf("Command failed as expected with exit code %d\nOutput:\n%s", code, outputStr)
	}

	return code, outputStr
}

// AssertFileExists checks that one or more files exist in the container.
// It fails the test if any file is missing.
func AssertFileExists(t *testing.T, container testcontainers.Container, paths ...string) {
	t.Helper()

	ctx := context.Background()
	cmd := append([]string{"stat"}, paths...)

	code, reader, err := container.Exec(ctx, cmd)
	output, _ := io.ReadAll(reader)

	assert.NoError(t, err, "Failed to execute stat command: %v", err)
	assert.Equal(t, 0, code,
		"One or more files do not exist: %v\nstat output:\n%s",
		paths, string(output))
}

// ReadFile reads the content of a file from the container and returns it as a byte slice.
func ReadFile(t *testing.T, container testcontainers.Container, path string) []byte {
	t.Helper()

	ctx := context.Background()
	reader, err := container.CopyFileFromContainer(ctx, path)
	require.NoError(t, err, "Failed to copy file '%s' from container", path)
	defer reader.Close()

	output, err := io.ReadAll(reader)
	require.NoError(t, err, "Failed to read content of file '%s'", path)

	return output
}

// ExecCommand executes an arbitrary command in the container.
// Use this for non-piper commands like ls, cat, etc.
func ExecCommand(t *testing.T, container testcontainers.Container, workDir string, command []string) string {
	t.Helper()

	ctx := context.Background()
	code, reader, err := container.Exec(ctx, command, exec.WithWorkingDir(workDir))

	// Always read output first
	output, readErr := io.ReadAll(reader)
	outputStr := string(output)

	require.NoError(t, err, "Failed to execute command: %v\nCommand: %v\nWorkDir: %s", err, command, workDir)
	require.NoError(t, readErr, "Failed to read output for command: %v", command)
	require.Equal(t, 0, code,
		"Command failed with exit code %d\nCommand: %v\nWorkDir: %s\nOutput:\n%s",
		code, command, workDir, outputStr)

	return outputStr
}

// getProjectRoot returns the absolute path to the project root directory.
// It assumes this is called from integration/testhelper and goes up two levels.
func getProjectRoot(t *testing.T) string {
	t.Helper()

	pwd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	// If we're in integration/testhelper, go up to integration, then to root
	// If we're in integration/, go up to root
	// This handles both running from integration/ and integration/testhelper/
	if filepath.Base(pwd) == "testhelper" {
		pwd = filepath.Dir(pwd)
	}
	if filepath.Base(pwd) == "integration" {
		pwd = filepath.Dir(pwd)
	}

	return pwd
}
