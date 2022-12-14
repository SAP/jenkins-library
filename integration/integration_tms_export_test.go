//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestTmsExportIntegration ./integration/...

package main

import (
	"os"
	"testing"
)

func TestTmsExportIntegration(t *testing.T) {
	//Reading TMS credentials from environment
	tmsServiceKey := os.Getenv("PIPER_tmsServiceKey")
	if len(tmsServiceKey) == 0 {
		tmsServiceKey := os.Getenv("PIPER_TMSSERVICEKEY")
		if len(tmsServiceKey) == 0 {
			t.Fatal("No tmsServiceKey maintained")
		}
	}

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "devxci/mbtci-java11-node14",
		User:        "root",
		TestDir:     []string{"testdata", "TestTmsUploadIntegration"},
		Environment: map[string]string{"PIPER_tmsServiceKey": tmsServiceKey},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("tmsExport",
		"--tmsServiceKey="+tmsServiceKey,
		"--mtaPath=scv_x.mtar",
		"--nodeName=PIPER-TEST",
		"--customDescription=Piper node export integration test",
		"--nodeExtDescriptorMapping={\"PIPER-TEST\":\"scv_x.mtaext\", \"PIPER-PROD\":\"scv_x.mtaext\"}",
		"--mtaVersion=1.0.0")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "tmsExport - SUCCESS")
}
