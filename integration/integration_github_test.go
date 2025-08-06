//go:build integration

// can be executed with
// go test -v -tags integration -run TestGitHubIntegration ./integration/...

package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/command"
	pipergithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/piperenv"
)

func TestGitHubIntegrationPiperPublishRelease(t *testing.T) {
	// t.Parallel()
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
	dir := t.TempDir()

	testAsset := filepath.Join(dir, "test.txt")
	err := os.WriteFile(testAsset, []byte("Test"), 0644)
	assert.NoError(t, err, "Error when writing temporary file")
	test2Asset := filepath.Join(dir, "test2.txt")
	err = os.WriteFile(test2Asset, []byte("Test"), 0644)
	assert.NoError(t, err, "Error when writing temporary file")

	t.Run("test single asset - success", func(t *testing.T) {
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
	})
	t.Run("test multiple assets - success", func(t *testing.T) {
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
			"--assetPathList",
			testAsset,
			"--assetPathList",
			test2Asset,
			"--noTelemetry",
		}

		err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
		assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)
	})
}

func TestGitHubIntegrationFetchCommitStatistics(t *testing.T) {
	// t.Parallel()
	// prepare
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
	// test
	result, err := pipergithub.FetchCommitStatistics(&pipergithub.FetchCommitOptions{
		Owner: owner, Repository: repository, APIURL: "https://api.github.com", Token: token, SHA: "3601ed6"})

	// assert
	assert.NoError(t, err)
	assert.Equal(t, 2, result.Additions)
	assert.Equal(t, 0, result.Deletions)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 1, result.Files)
}
