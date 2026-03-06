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
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerConfig holds configuration for creating a test container
type ContainerConfig struct {
	Image    string // Docker image to use
	TestData string // Path relative to integration/testdata (e.g., "TestGradleIntegration/java-project")
	WorkDir  string // Working directory inside container (e.g., "/java-project")
	User     string // User to run as (optional, defaults to image default)
}

// K3dContainerConfig holds configuration for creating a k3d cluster container
type K3dContainerConfig struct {
	TestData   string   // Path relative to integration/testdata
	Namespaces []string // Kubernetes namespaces to create (optional)
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
				FileMode:          0o755,
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

// StartK3dContainer starts a Docker-in-Docker container with a k3d Kubernetes cluster.
// Uses rancher/k3d:5.3.0-dind which has k3d pre-installed.
// The container is automatically cleaned up when the test finishes via t.Cleanup.
func StartK3dContainer(t *testing.T, cfg K3dContainerConfig) testcontainers.Container {
	t.Helper()

	ctx := context.Background()
	projectRoot := getProjectRoot(t)

	req := testcontainers.ContainerRequest{
		Image:      "rancher/k3d:5.3.0-dind",
		Privileged: true,
		Env: map[string]string{
			"DOCKER_TLS_CERTDIR": "",
		},
		ExposedPorts: []string{"2375/tcp"},
		WaitingFor:   wait.ForLog("API listen on").WithStartupTimeout(60 * time.Second),
	}

	if cfg.TestData != "" {
		req.Files = []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(projectRoot, "integration", "testdata", cfg.TestData),
				ContainerFilePath: "/",
				FileMode:          0o755,
			},
		}
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "Failed to start dind container")

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	})

	// Install kubectl (k3d is pre-installed in the image)
	t.Log("Installing kubectl...")
	setupCmd := `curl -Lo /usr/local/bin/kubectl https://dl.k8s.io/release/v1.28.0/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl`
	exitCode, reader, err := container.Exec(ctx, []string{"sh", "-c", setupCmd})
	require.NoError(t, err, "Failed to install kubectl")
	if exitCode != 0 {
		output, _ := io.ReadAll(reader)
		t.Fatalf("Failed to install kubectl: %s", string(output))
	}

	// Create k3d cluster
	t.Log("Creating k3d cluster...")
	exitCode, reader, err = container.Exec(ctx, []string{"sh", "-c", "k3d cluster create test --wait"})
	require.NoError(t, err, "Failed to create k3d cluster")
	if exitCode != 0 {
		output, _ := io.ReadAll(reader)
		t.Fatalf("Failed to create k3d cluster: %s", string(output))
	}
	t.Log("k3d cluster created successfully")

	// Create namespaces
	for _, ns := range cfg.Namespaces {
		exitCode, reader, err = container.Exec(ctx, []string{"kubectl", "create", "namespace", ns})
		require.NoError(t, err, "Failed to create namespace %s", ns)
		if exitCode != 0 {
			output, _ := io.ReadAll(reader)
			t.Fatalf("Failed to create namespace %s: %s", ns, string(output))
		}
	}

	// Copy piper binary
	err = container.CopyFileToContainer(ctx, filepath.Join(projectRoot, "piper"), "/piperbin/piper", 0o755)
	require.NoError(t, err, "Failed to copy piper binary - build it first with: CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o piper")

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
		t.Errorf("WARNING: Command succeeded with exit code 0 (expected failure)\nOutput:\n%s", outputStr)
	}

	return code, outputStr
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

// ExecCommandExpectNonZero executes a command and returns the exit code and output.
// Unlike ExecCommand, this does not fail the test on non-zero exit codes.
// Use this when you need to check the exit code yourself.
func ExecCommandExpectNonZero(t *testing.T, container testcontainers.Container, workDir string, command []string) (int, string) {
	t.Helper()

	ctx := context.Background()
	code, reader, err := container.Exec(ctx, command, exec.WithWorkingDir(workDir))

	output, readErr := io.ReadAll(reader)
	outputStr := string(output)

	require.NoError(t, err, "Failed to execute command: %v", command)
	require.NoError(t, readErr, "Failed to read output for command: %v", command)

	return code, outputStr
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
