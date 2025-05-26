//go:build integration

// can be executed with
// go test -v -tags integration -run TestGitOpsIntegration ./integration/...

package main

import (
	"testing"
)

func TestGitOpsIntegrationUpdateDeployment(t *testing.T) {
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "nekottyo/kustomize-kubeval:kustomizev4",
		TestDir: []string{"testdata", "TestGitopsUpdateIntegration", "kustomize", "workdir"},
		Mounts:  map[string]string{"./testdata/TestGitopsUpdateIntegration/kustomize/gitopsRepo": "/gitopsRepo-source"},
		Setup:   []string{"cp -r /gitopsRepo-source /gitopsRepo"},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("gitopsUpdateDeployment", "--containerImageNameTag=image:456")
	if err != nil {
		t.Fatalf("Calling piper command failed %s\n", err)
	}
	err = container.Runner.RunExecutable("docker", "exec", container.ContainerName, "git", "clone", "/gitopsRepo", "/tmp/repo")
	if err != nil {
		t.Fatalf("Cloing of bare repo failed")
	}

	container.assertHasOutput(t, "SUCCESS", "[kustomize] updating")
	container.assertFileContentEquals(t, "/tmp/repo/kustomization.yaml", `images:
- name: test-project
  newName: image
  newTag: "456"
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`)
}
