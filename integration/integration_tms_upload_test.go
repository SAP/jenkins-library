//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestTmsUploadIntegration ./integration/...

package main

import (
	"testing"
)

func TestTmsUploadIntegration1(t *testing.T) {
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "devxci/mbtci-java11-node14",
		User:    "root",
		TestDir: []string{"testdata", "TestTmsUploadIntegration"},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("tmsUpload",
		"--mtaPath=scv_x.mtar",
		"--nodeName=PIPER-TEST",
		"--customDescription=Piper integration test",
		"--nodeExtDescriptorMapping={\"PIPER-TEST\":\"scv_x.mtaext\", \"PIPER-PROD\":\"scv_x.mtaext\"}",
		"--mtaVersion=1.0.0")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "tmsUpload - SUCCESS")
}
