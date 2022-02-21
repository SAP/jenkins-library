//go:build integration
// +build integration

// can be execute with go test -tags=integration ./integration/...

package main

import (
	"testing"
)

const (
	installCommand string = "npm install -g @getgauge/cli --prefix=~/.npm-global --unsafe-perm" // option --unsafe-perm need to install gauge in docker container. See this issue: https://github.com/getgauge/gauge/issues/1470
)

func runTest(t *testing.T, languageRunner string) {
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "getgauge/gocd-jdk-mvn-node",
		TestDir: []string{"testdata", "TestGaugeIntegration", "gauge-" + languageRunner},
	})

	err := container.whenRunningPiperCommand("gaugeExecuteTests", "--installCommand", installCommand, "--languageRunner", languageRunner, "--runCommand", "run")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "info  gaugeExecuteTests - Scenarios:	2 executed	2 passed	0 failed	0 skipped")
	container.assertHasOutput(t, "info  gaugeExecuteTests - SUCCESS")
}

func TestGaugeJava(t *testing.T) {
	t.Parallel()
	runTest(t, "java")
}

func TestGaugeJS(t *testing.T) {
	t.Parallel()
	runTest(t, "js")
}
