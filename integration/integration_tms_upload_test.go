//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestTmsIntegration ./integration

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const tmsTestDockerImage = "alpine:latest"

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
		Image:       tmsTestDockerImage,
		User:        "root",
		TestDir:     []string{"testdata", "TestTmsIntegration"},
		Environment: map[string]string{"PIPER_serviceKey": tmsServiceKey},
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
	container.assertHasOutput(t, "description: Piper integration test")
	container.assertHasOutput(t, "tmsUpload - File uploaded successfully")
	container.assertHasOutput(t, "tmsUpload - Node upload executed successfully")
	container.assertHasOutput(t, "tmsUpload - SUCCESS")
}

func TestTmsUploadIntegrationBinNoDescriptionSuccess(t *testing.T) {
	// success case: run cmd without --nodeExtDescriptorMapping and --customDescription
	readEnv()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       tmsTestDockerImage,
		User:        "root",
		TestDir:     []string{"testdata", "TestTmsIntegration"},
		Environment: map[string]string{"PIPER_serviceKey": tmsServiceKey},
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
	// error case: run cmd with nodeExtDescriptorMapping
	readEnv()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:   tmsTestDockerImage,
		User:    "root",
		TestDir: []string{"testdata", "TestTmsIntegration"},
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

func TestTmsUploadIntegrationBinFailDescription(t *testing.T) {
	// error case: run cmd with invalid description
	readEnv()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       tmsTestDockerImage,
		User:        "root",
		TestDir:     []string{"testdata", "TestTmsIntegration"},
		Environment: map[string]string{"PIPER_serviceKey": tmsServiceKey},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("tmsUpload",
		"--mtaPath=scv_x.mtar",
		"--nodeName=PIPER-TEST",
		"--customDescription={Bad description}")

	assert.Error(t, err, "Did expect error")
	container.assertHasOutput(t, "error tmsUpload - HTTP request failed with error")
	container.assertHasOutput(t, "Failed to run tmsUpload step")
	container.assertHasOutput(t, "failed to upload file to node")
}

func TestTmsUploadIntegrationYaml(t *testing.T) {
	// success case: run with custom config
	readEnv()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       tmsTestDockerImage,
		User:        "root",
		TestDir:     []string{"testdata", "TestTmsIntegration"},
		Environment: map[string]string{"PIPER_serviceKey": tmsServiceKey},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("tmsUpload", "--customConfig=.pipeline/upload_config.yml")
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
