//go:build integration
// +build integration

// can be execute with go test -tags=integration ./integration/...

package main

import (
	"testing"
)

func TestBuildProject(t *testing.T) {
	t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "python:3.9",
		TestDir: []string{"testdata", "TestPythonIntegration"},
	})

	err := container.whenRunningPiperCommand("python3", "-m", "pip", "install", "--upgrade", "build")
	if err != nil {
		t.Fatalf("Calling piper command failed %s", err)
	}
	var flags []string
	flags = append(flags, "-m", "build", "/testdata/TestPythonIntegration")

	err = container.whenRunningPiperCommand("python3", flags...)
	if err != nil {
		t.Fatalf("Calling piper command failed %s", err)
	}

	container.assertHasOutput(t, "Successfully built example-package-YOUR-USERNAME-HERE-0.0.1.tar.gz and example_package_YOUR_USERNAME_HERE-0.0.1-py3-none-any.whl")
	container.assertHasFile(t, "/dist/example-package-YOUR-USERNAME-HERE-0.0.1.tar.gz")
	container.assertHasFile(t, "/dist/example_package_YOUR_USERNAME_HERE-0.0.1-py3-none-any.whl")
}
