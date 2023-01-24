//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestTmsUploadIntegration ./integration/...

package main

import (
	"fmt"
	"os"
	"testing"
)

func readEnv() string {
	//Reading TMS credentials from environment
	tmsServiceKey := os.Getenv("PIPER_tmsServiceKey")
	if len(tmsServiceKey) == 0 {
		tmsServiceKey := os.Getenv("PIPER_TMSSERVICEKEY")
		if len(tmsServiceKey) == 0 {
			fmt.Println("No tmsServiceKey maintained")
		}
	}
	return tmsServiceKey
}

func TestTmsUploadIntegrationBin(t *testing.T) {
	tmsServiceKey := readEnv()

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "devxci/mbtci-java11-node14",
		User:    "root",
		TestDir: []string{"testdata", "TestTmsUploadIntegration/bin_param"},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("tmsUpload",
		"--tmsServiceKey="+tmsServiceKey,
		"--mtaPath=scv_x.mtar",
		"--nodeName=PIPER-TEST",
		"--customDescription=Piper integration test",
		"--nodeExtDescriptorMapping={\"PIPER-TEST\":\"scv_x.mtaext\", \"PIPER-PROD\":\"scv_x.mtaext\"}",
		"--mtaVersion=1.0.0",
		"-v")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "tmsUpload - File uploaded successfully")
	container.assertHasOutput(t, "tmsUpload - SUCCESS")
}

func TestTmsUploadIntegrationYaml(t *testing.T) {
	tmsServiceKey := readEnv()

	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "devxci/mbtci-java11-node14",
		User:        "root",
		TestDir:     []string{"testdata", "TestTmsUploadIntegration/yaml"},
		Environment: map[string]string{"PIPER_tmsServiceKey": tmsServiceKey},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("tmsUpload")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "tmsUpload - File uploaded successfully")
	container.assertHasOutput(t, "tmsUpload - SUCCESS")
	container.assertHasNoOutput(t, "Config value for 'nodeExtDescriptorMapping' is of unexpected type map, expected string")
}
