//go:build unit
// +build unit

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
	t.Run("Valid URL1 with dots", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/com.sap.fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
		assert.Equal(t, "Testing/com.sap.fortify", repoInfo.repo)
	})

	t.Run("Valid URL2 with dots", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/com.sap.fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
		assert.Equal(t, "Testing/com.sap.fortify", repoInfo.repo)
	})
	t.Run("Valid URL1 with username and token", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://username:token@github.hello.test/Testing/fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
		assert.Equal(t, "Testing/fortify", repoInfo.repo)
	})

	t.Run("Valid URL2 with username and token", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://username:token@github.hello.test/Testing/fortify", &repoInfo)
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

func TestParseRepositoryURL(t *testing.T) {
	t.Run("Valid repository", func(t *testing.T) {
		repository := "https://github.hello.test/Testing/fortify.git"
		toolInstance, orgName, repoName, err := parseRepositoryURL(repository)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", toolInstance)
		assert.Equal(t, "Testing", orgName)
		assert.Equal(t, "fortify", repoName)
	})
	t.Run("valid repository 2", func(t *testing.T) {
		repository := "https://github.hello.test/Testing/fortify"
		toolInstance, orgName, repoName, err := parseRepositoryURL(repository)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", toolInstance)
		assert.Equal(t, "Testing", orgName)
		assert.Equal(t, "fortify", repoName)
	})
	t.Run("Invalid repository without repo name", func(t *testing.T) {
		repository := "https://github.hello.test/Testing"
		toolInstance, orgName, repoName, err := parseRepositoryURL(repository)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Unable to parse organization and repo names")
		assert.Equal(t, "", toolInstance)
		assert.Equal(t, "", orgName)
		assert.Equal(t, "", repoName)
	})
	t.Run("Invalid repository without organization name", func(t *testing.T) {
		repository := "https://github.hello.test/fortify"
		toolInstance, orgName, repoName, err := parseRepositoryURL(repository)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Unable to parse organization and repo names")
		assert.Equal(t, "", toolInstance)
		assert.Equal(t, "", orgName)
		assert.Equal(t, "", repoName)
	})
	t.Run("Invalid repository without tool instance", func(t *testing.T) {
		repository := "/Testing/fortify"
		toolInstance, orgName, repoName, err := parseRepositoryURL(repository)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Unable to parse tool instance")
		assert.Equal(t, "", toolInstance)
		assert.Equal(t, "", orgName)
		assert.Equal(t, "", repoName)
	})
	t.Run("Empty repository", func(t *testing.T) {
		repository := ""
		toolInstance, orgName, repoName, err := parseRepositoryURL(repository)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Repository param is not set")
		assert.Equal(t, "", toolInstance)
		assert.Equal(t, "", orgName)
		assert.Equal(t, "", repoName)
	})
}

func TestBuildRepoReference(t *testing.T) {
	t.Run("Valid ref with branch", func(t *testing.T) {
		repository := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/head/branch"
		ref, err := buildRepoReference(repository, analyzedRef)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test/Testing/fortify/tree/branch", ref)
	})
	t.Run("Valid ref with PR", func(t *testing.T) {
		repository := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/pull/1/merge"
		ref, err := buildRepoReference(repository, analyzedRef)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test/Testing/fortify/pull/1", ref)
	})
	t.Run("Invalid ref without branch name", func(t *testing.T) {
		repository := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/head"
		ref, err := buildRepoReference(repository, analyzedRef)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Wrong analyzedRef format")
		assert.Equal(t, "", ref)
	})
	t.Run("Invalid ref without PR id", func(t *testing.T) {
		repository := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/pull/merge"
		ref, err := buildRepoReference(repository, analyzedRef)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Wrong analyzedRef format")
		assert.Equal(t, "", ref)
	})
	t.Run("Empty repository", func(t *testing.T) {
		repository := ""
		analyzedRef := "refs/pull/merge"
		ref, err := buildRepoReference(repository, analyzedRef)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Repository or analyzedRef param is not set")
		assert.Equal(t, "", ref)
	})
	t.Run("Empty analyzedRef", func(t *testing.T) {
		repository := "https://github.hello.test/Testing/fortify"
		analyzedRef := ""
		ref, err := buildRepoReference(repository, analyzedRef)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Repository or analyzedRef param is not set")
		assert.Equal(t, "", ref)
	})
}

func TestCreateToolRecordCodeql(t *testing.T) {
	t.Run("Valid toolrun file", func(t *testing.T) {
		config := codeqlExecuteScanOptions{
			Repository:  "https://github.hello.test/Testing/fortify.git",
			AnalyzedRef: "refs/head/branch",
			CommitID:    "test",
		}
		fileName, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), "test", config)
		assert.NoError(t, err)
		assert.Contains(t, fileName, "toolrun_codeql")
	})
	t.Run("Empty repository URL", func(t *testing.T) {
		config := codeqlExecuteScanOptions{
			Repository:  "",
			AnalyzedRef: "refs/head/branch",
			CommitID:    "test",
		}
		fileName, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), "", config)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Repository param is not set")
		assert.Empty(t, fileName)
	})
	t.Run("Invalid repository URL", func(t *testing.T) {
		config := codeqlExecuteScanOptions{
			Repository:  "https://github.hello.test/Testing",
			AnalyzedRef: "refs/head/branch",
			CommitID:    "test",
		}
		fileName, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), "test", config)
		assert.Error(t, err)
		assert.Regexp(t, "^Unable to parse [a-z ]+ from repository url$", err.Error())
		assert.Empty(t, fileName)
	})
	t.Run("Empty workspace", func(t *testing.T) {
		config := codeqlExecuteScanOptions{
			Repository:  "https://github.hello.test/Testing/fortify.git",
			AnalyzedRef: "refs/head/branch",
			CommitID:    "test",
		}
		fileName, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), "", config)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "TR_PERSIST: empty workspace")
		assert.Empty(t, fileName)
	})
	t.Run("Empty analyzedRef", func(t *testing.T) {
		config := codeqlExecuteScanOptions{
			Repository:  "https://github.hello.test/Testing/fortify.git",
			AnalyzedRef: "",
			CommitID:    "test",
		}
		fileName, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), "test", config)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "TR_ADD_KEY: empty keyvalue")
		assert.Empty(t, fileName, "toolrun_codeql")
	})
	t.Run("Invalid analyzedRef", func(t *testing.T) {
		config := codeqlExecuteScanOptions{
			Repository:  "https://github.hello.test/Testing/fortify.git",
			AnalyzedRef: "refs/head",
			CommitID:    "test",
		}
		fileName, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), "test", config)
		assert.NoError(t, err)
		assert.Contains(t, fileName, "toolrun_codeql")
	})
}
