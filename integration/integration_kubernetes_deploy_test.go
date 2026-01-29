//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestKubernetesDeployIntegration ./integration/...

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestKubernetesDeployIntegrationKubectlSecretNamespace tests that secrets are created
// in the correct namespace when using kubectl deployment with a real k3d cluster.
func TestKubernetesDeployIntegrationKubectlSecretNamespace(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping k3d cluster test in short mode")
	}

	container := StartK3dContainer(t, K3dContainerConfig{
		TestData:   "TestKubernetesDeployIntegration/kubectl",
		Namespaces: []string{"test-namespace"},
	})

	t.Log("Running piper kubernetesDeploy...")
	output := RunPiper(t, container, "/kubectl",
		"kubernetesDeploy",
		"--deployTool=kubectl",
		"--containerRegistryUrl=https://my.registry:55555",
		"--containerRegistryUser=testuser",
		"--containerRegistryPassword=testpassword",
		"--containerRegistrySecret=reg-secret",
		"--namespace=test-namespace",
		"--appTemplate=/kubectl/deployment.yaml",
		"--deployCommand=apply",
		"--dockerConfigJSON=/kubectl/.pipeline/docker/config.json",
		"--containerImageName=nginx",
		"--containerImageTag=latest",
		"--insecureSkipTLSVerify=true",
	)

	assert.Contains(t, output, "Creating container registry secret 'reg-secret'")

	// Verify secret was created in the correct namespace
	t.Log("Verifying secret in test-namespace...")
	secretOutput := ExecCommand(t, container, "/", []string{
		"kubectl", "get", "secret", "reg-secret", "-n", "test-namespace", "-o", "name",
	})
	assert.Contains(t, secretOutput, "secret/reg-secret")

	// Verify secret does NOT exist in default namespace
	exitCode, _ := ExecCommandExpectFailure(t, container, "/", []string{
		"kubectl", "get", "secret", "reg-secret", "-n", "default",
	})
	assert.NotEqual(t, 0, exitCode, "Secret should NOT exist in default namespace")

	t.Log("Test passed: Secret was created in the correct namespace")
}

// TestKubernetesDeployIntegrationKubectlMultipleNamespaces tests deploying to multiple namespaces.
func TestKubernetesDeployIntegrationKubectlMultipleNamespaces(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping k3d cluster test in short mode")
	}

	container := StartK3dContainer(t, K3dContainerConfig{
		TestData:   "TestKubernetesDeployIntegration/kubectl",
		Namespaces: []string{"namespace-a", "namespace-b"},
	})

	t.Log("Deploying to namespace-a...")
	RunPiper(t, container, "/kubectl",
		"kubernetesDeploy",
		"--deployTool=kubectl",
		"--containerRegistryUrl=https://my.registry:55555",
		"--containerRegistryUser=testuser",
		"--containerRegistryPassword=testpassword",
		"--containerRegistrySecret=secret-a",
		"--namespace=namespace-a",
		"--appTemplate=/kubectl/deployment.yaml",
		"--deployCommand=apply",
		"--dockerConfigJSON=/kubectl/.pipeline/docker/config.json",
		"--containerImageName=nginx",
		"--containerImageTag=latest",
		"--insecureSkipTLSVerify=true",
	)

	t.Log("Deploying to namespace-b...")
	RunPiper(t, container, "/kubectl",
		"kubernetesDeploy",
		"--deployTool=kubectl",
		"--containerRegistryUrl=https://my.registry:55555",
		"--containerRegistryUser=testuser",
		"--containerRegistryPassword=testpassword",
		"--containerRegistrySecret=secret-b",
		"--namespace=namespace-b",
		"--appTemplate=/kubectl/deployment.yaml",
		"--deployCommand=apply",
		"--dockerConfigJSON=/kubectl/.pipeline/docker/config.json",
		"--containerImageName=nginx",
		"--containerImageTag=latest",
		"--insecureSkipTLSVerify=true",
	)

	// Verify secrets are in correct namespaces
	ExecCommand(t, container, "/", []string{"kubectl", "get", "secret", "secret-a", "-n", "namespace-a"})
	ExecCommand(t, container, "/", []string{"kubectl", "get", "secret", "secret-b", "-n", "namespace-b"})

	// Cross-check: secrets should NOT exist in wrong namespaces
	exitCode, _ := ExecCommandExpectFailure(t, container, "/", []string{"kubectl", "get", "secret", "secret-a", "-n", "namespace-b"})
	assert.NotEqual(t, 0, exitCode)
	exitCode, _ = ExecCommandExpectFailure(t, container, "/", []string{"kubectl", "get", "secret", "secret-b", "-n", "namespace-a"})
	assert.NotEqual(t, 0, exitCode)

	t.Log("Test passed: Secrets were created in their correct namespaces")
}
