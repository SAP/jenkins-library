package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type codeqlExecuteScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newCodeqlExecuteScanTestsUtils() codeqlExecuteScanMockUtils {
	utils := codeqlExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunCodeqlExecuteScan(t *testing.T) {

	t.Run("Valid CodeqlExecuteScan", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ModulePath: "./"}
		assert.Equal(t, nil, runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils()))
	})

	t.Run("No auth token passed on upload results", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", UploadResults: true, ModulePath: "./"}
		assert.Error(t, runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils()))
	})

	t.Run("GitCommitID is NA on upload results", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", UploadResults: true, ModulePath: "./", CommitID: "NA"}
		assert.Error(t, runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils()))
	})

	t.Run("Upload results with token", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ModulePath: "./", UploadResults: true, GithubToken: "test"}
		assert.Equal(t, nil, runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils()))
	})

	t.Run("Custom buildtool", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "custom", Language: "javascript", ModulePath: "./", GithubToken: "test"}
		assert.Equal(t, nil, runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils()))
	})

	t.Run("Custom buildtool but no language specified", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "custom", ModulePath: "./", GithubToken: "test"}
		assert.Error(t, runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils()))
	})

	t.Run("Invalid buildtool and no language specified", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "test", ModulePath: "./", GithubToken: "test"}
		assert.Error(t, runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils()))
	})

	t.Run("Invalid buildtool but language specified", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "test", Language: "javascript", ModulePath: "./", GithubToken: "test"}
		assert.Equal(t, nil, runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils()))
	})
}

func TestGetGitRepoInfo(t *testing.T) {
	t.Run("Valid URL1", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
		assert.Equal(t, "Testing/fortify", repoInfo.repo)
	})

	t.Run("Valid URL2", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
		assert.Equal(t, "Testing/fortify", repoInfo.repo)
	})

	t.Run("Invalid URL as no org/owner passed", func(t *testing.T) {
		var repoInfo RepoInfo
		assert.Error(t, getGitRepoInfo("https://github.com/fortify", &repoInfo))
	})

	t.Run("Invalid URL as no protocol passed", func(t *testing.T) {
		var repoInfo RepoInfo
		assert.Error(t, getGitRepoInfo("github.hello.test/Testing/fortify", &repoInfo))
	})
}
