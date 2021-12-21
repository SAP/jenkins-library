//go:build integration
// +build integration

// can be execute with go test -tags=integration ./integration/...

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/sonar"
)

func TestPiperGithubPublishRelease(t *testing.T) {
	t.Parallel()
	token := os.Getenv("PIPER_INTEGRATION_GITHUB_TOKEN")
	if len(token) == 0 {
		t.Fatal("No GitHub token maintained")
	}

	owner := os.Getenv("PIPER_INTEGRATION_GITHUB_OWNER")
	if len(owner) == 0 {
		owner = "OliverNocon"
	}
	repository := os.Getenv("PIPER_INTEGRATION_GITHUB_REPOSITORY")
	if len(repository) == 0 {
		repository = "piper-integration"
	}

	dir, err := ioutil.TempDir("", "")
	defer os.RemoveAll(dir) // clean up
	assert.NoError(t, err, "Error when creating temp dir")

	testAsset := filepath.Join(dir, "test.txt")
	err = ioutil.WriteFile(testAsset, []byte("Test"), 0644)
	assert.NoError(t, err, "Error when writing temporary file")

	//prepare pipeline environment
	now := time.Now()
	piperenv.SetResourceParameter(filepath.Join(dir, ".pipeline"), "commonPipelineEnvironment", "artifactVersion", now.Format("20060102150405"))

	cmd := command.Command{}
	cmd.SetDir(dir)

	piperOptions := []string{
		"githubPublishRelease",
		"--owner",
		owner,
		"--repository",
		repository,
		"--token",
		token,
		"--assetPath",
		testAsset,
		"--noTelemetry",
	}

	err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)
}
