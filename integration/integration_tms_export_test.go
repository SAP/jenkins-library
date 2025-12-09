//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestTmsExportIntegration ./integration/...

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTmsExportIntegrationYaml(t *testing.T) {
	// success case: run with custom config
	readEnv()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       tmsTestDockerImage,
		User:        "root",
		TestDir:     []string{"testdata", "TestTmsIntegration"},
		Environment: map[string]string{"PIPER_serviceKey": tmsServiceKey},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("tmsExport", "--customConfig=.pipeline/export_config.yml")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "tmsExport - File uploaded successfully")
	container.assertHasOutput(t, "tmsExport - MTA extension descriptor updated successfully")
	container.assertHasOutput(t, "tmsExport - Node export executed successfully")
	container.assertHasOutput(t, "tmsExport - SUCCESS")
}

func TestTmsExportIntegrationBinFailDescription(t *testing.T) {
	// error case: run cmd with invalid description
	readEnv()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       tmsTestDockerImage,
		User:        "root",
		TestDir:     []string{"testdata", "TestTmsIntegration"},
		Environment: map[string]string{"PIPER_serviceKey": tmsServiceKey},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("tmsExport",
		"--mtaPath=scv_x.mtar",
		"--nodeName=PIPER-TEST",
		"--customDescription={Bad description}")

	assert.Error(t, err, "Did expect error")
	container.assertHasOutput(t, "error tmsExport - HTTP request failed with error")
	container.assertHasOutput(t, "Failed to run tmsExport")
	container.assertHasOutput(t, "failed to export file to node")
}
