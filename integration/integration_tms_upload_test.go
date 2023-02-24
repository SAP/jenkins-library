//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestTmsUploadIntegration ./integration

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tmsServiceKey string

func readEnv() {
	//Reading TMS credentials from environment
	tmsServiceKey = os.Getenv("PIPER_tmsServiceKey")
	if len(tmsServiceKey) == 0 {
		fmt.Println("Env. variable PIPER_tmsServiceKey is not provided")
		os.Exit(1)
	}
}

func TestTmsUploadIntegrationBinSuccess(t *testing.T) {
	// success case: run cmd without nodeExtDescriptorMapping
	readEnv()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "devxci/mbtci-java11-node14",
		User:        "root",
		TestDir:     []string{"testdata", "TestTmsUploadIntegration/bin_param"},
		Environment: map[string]string{"PIPER_tmsServiceKey": tmsServiceKey},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("tmsUpload",
		"--mtaPath=scv_x.mtar",
		"--nodeName=PIPER-TEST",
		"--customDescription=Piper integration test",
		"--mtaVersion=1.0.0",
		"-v")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}
	container.assertHasOutput(t, "tmsUpload - File uploaded successfully")
	container.assertHasOutput(t, "tmsUpload - Node upload executed successfully")
	container.assertHasOutput(t, "tmsUpload - SUCCESS")
}

func TestTmsUploadIntegrationBinNoDescriptionSuccess(t *testing.T) {
	// success case: run cmd without --nodeExtDescriptorMapping and --customDescription
	readEnv()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "devxci/mbtci-java11-node14",
		User:        "root",
		TestDir:     []string{"testdata", "TestTmsUploadIntegration/bin_param"},
		Environment: map[string]string{"PIPER_tmsServiceKey": tmsServiceKey},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("tmsUpload",
		"--mtaPath=scv_x.mtar",
		"--nodeName=PIPER-TEST",
		"--mtaVersion=1.0.0",
		"-v")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}
	container.assertHasOutput(t, "description: Created by Piper")
	container.assertHasOutput(t, "tmsUpload - File uploaded successfully")
	container.assertHasOutput(t, "tmsUpload - Node upload executed successfully")
	container.assertHasOutput(t, "tmsUpload - SUCCESS")
}

func TestTmsUploadIntegrationBinFailParam(t *testing.T) {
	// run cmd with nodeExtDescriptorMapping
	readEnv()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   "devxci/mbtci-java11-node14",
		User:    "root",
		TestDir: []string{"testdata", "TestTmsUploadIntegration/bin_param"},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("tmsUpload",
		"--mtaPath=scv_x.mtar",
		"--nodeName=PIPER-TEST",
		"--customDescription=Piper integration test",
		"--nodeExtDescriptorMapping={\"PIPER-TEST\":\"scv_x.mtaext\", \"PIPER-PROD\":\"scv_x.mtaext\"}",
		"--mtaVersion=1.0.0",
		"-v")

	assert.Error(t, err, "Did expect error")
	container.assertHasOutput(t, "Error: unknown flag: --nodeExtDescriptorMapping")
}

func TestTmsUploadIntegrationYaml(t *testing.T) {
	// success case: run with config.yaml
	readEnv()
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
	container.assertHasOutput(t, "tmsUpload - MTA extension descriptor updated successfully")
	container.assertHasOutput(t, "tmsUpload - Node upload executed successfully")
	container.assertHasOutput(t, "tmsUpload - SUCCESS")
	//  test that oauth token is not exposed
	container.assertHasNoOutput(t, "eyJ")
}