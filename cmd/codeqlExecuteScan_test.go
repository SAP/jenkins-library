package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
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
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
	})

	t.Run("No auth token passed on upload results", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", UploadResults: true, ModulePath: "./"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("GitCommitID is NA on upload results", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", UploadResults: true, ModulePath: "./", CommitID: "NA"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("Upload results fails as repository not specified", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "maven", ModulePath: "./", UploadResults: true, GithubToken: "test"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("Custom buildtool", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "custom", Language: "javascript", ModulePath: "./"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
	})

	t.Run("Custom buildtool but no language specified", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "custom", ModulePath: "./", GithubToken: "test"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("Invalid buildtool and no language specified", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "test", ModulePath: "./", GithubToken: "test"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.Error(t, err)
	})

	t.Run("Invalid buildtool but language specified", func(t *testing.T) {
		config := codeqlExecuteScanOptions{BuildTool: "test", Language: "javascript", ModulePath: "./", GithubToken: "test"}
		_, err := runCodeqlExecuteScan(&config, nil, newCodeqlExecuteScanTestsUtils())
		assert.NoError(t, err)
	})
}

func TestGetGitRepoInfo(t *testing.T) {
	t.Run("Valid URL1", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
		assert.Equal(t, "fortify", repoInfo.repo)
		assert.Equal(t, "Testing", repoInfo.owner)
	})

	t.Run("Valid URL2", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
		assert.Equal(t, "fortify", repoInfo.repo)
		assert.Equal(t, "Testing", repoInfo.owner)
	})
	t.Run("Valid URL1 with dots", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/com.sap.fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
		assert.Equal(t, "com.sap.fortify", repoInfo.repo)
		assert.Equal(t, "Testing", repoInfo.owner)
	})

	t.Run("Valid URL2 with dots", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://github.hello.test/Testing/com.sap.fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
		assert.Equal(t, "com.sap.fortify", repoInfo.repo)
		assert.Equal(t, "Testing", repoInfo.owner)
	})
	t.Run("Valid URL1 with username and token", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://username:token@github.hello.test/Testing/fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
		assert.Equal(t, "fortify", repoInfo.repo)
		assert.Equal(t, "Testing", repoInfo.owner)
	})

	t.Run("Valid URL2 with username and token", func(t *testing.T) {
		var repoInfo RepoInfo
		err := getGitRepoInfo("https://username:token@github.hello.test/Testing/fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
		assert.Equal(t, "fortify", repoInfo.repo)
		assert.Equal(t, "Testing", repoInfo.owner)
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

func TestInitGitInfo(t *testing.T) {
	t.Run("Valid URL1", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://github.hello.test/Testing/codeql.git", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo := initGitInfo(&config)
		assert.Equal(t, "abcd1234", repoInfo.commitId)
		assert.Equal(t, "Testing", repoInfo.owner)
		assert.Equal(t, "codeql", repoInfo.repo)
		assert.Equal(t, "refs/head/branch", repoInfo.ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
	})

	t.Run("Valid URL2", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://github.hello.test/Testing/codeql", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo := initGitInfo(&config)
		assert.Equal(t, "abcd1234", repoInfo.commitId)
		assert.Equal(t, "Testing", repoInfo.owner)
		assert.Equal(t, "codeql", repoInfo.repo)
		assert.Equal(t, "refs/head/branch", repoInfo.ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
	})

	t.Run("Valid url with dots URL1", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://github.hello.test/Testing/com.sap.codeql.git", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo := initGitInfo(&config)
		assert.Equal(t, "abcd1234", repoInfo.commitId)
		assert.Equal(t, "Testing", repoInfo.owner)
		assert.Equal(t, "com.sap.codeql", repoInfo.repo)
		assert.Equal(t, "refs/head/branch", repoInfo.ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
	})

	t.Run("Valid url with dots URL2", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://github.hello.test/Testing/com.sap.codeql", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo := initGitInfo(&config)
		assert.Equal(t, "abcd1234", repoInfo.commitId)
		assert.Equal(t, "Testing", repoInfo.owner)
		assert.Equal(t, "com.sap.codeql", repoInfo.repo)
		assert.Equal(t, "refs/head/branch", repoInfo.ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
	})

	t.Run("Valid url with username and token URL1", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://username:token@github.hello.test/Testing/codeql.git", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo := initGitInfo(&config)
		assert.Equal(t, "abcd1234", repoInfo.commitId)
		assert.Equal(t, "Testing", repoInfo.owner)
		assert.Equal(t, "codeql", repoInfo.repo)
		assert.Equal(t, "refs/head/branch", repoInfo.ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
	})

	t.Run("Valid url with username and token URL2", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://username:token@github.hello.test/Testing/codeql", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo := initGitInfo(&config)
		assert.Equal(t, "abcd1234", repoInfo.commitId)
		assert.Equal(t, "Testing", repoInfo.owner)
		assert.Equal(t, "codeql", repoInfo.repo)
		assert.Equal(t, "refs/head/branch", repoInfo.ref)
		assert.Equal(t, "https://github.hello.test", repoInfo.serverUrl)
	})

	t.Run("Invalid URL with no org/reponame", func(t *testing.T) {
		config := codeqlExecuteScanOptions{Repository: "https://github.hello.test", AnalyzedRef: "refs/head/branch", CommitID: "abcd1234"}
		repoInfo := initGitInfo(&config)
		_, err := orchestrator.NewOrchestratorSpecificConfigProvider()
		assert.Equal(t, "abcd1234", repoInfo.commitId)
		assert.Equal(t, "refs/head/branch", repoInfo.ref)
		if err != nil {
			assert.Equal(t, "", repoInfo.owner)
			assert.Equal(t, "", repoInfo.repo)
			assert.Equal(t, "", repoInfo.serverUrl)
		}
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
}

func getRepoReferences(repoInfo RepoInfo) (string, string, string) {
	repoUrl := fmt.Sprintf("%s/%s/%s", repoInfo.serverUrl, repoInfo.owner, repoInfo.repo)
	repoReference, _ := buildRepoReference(repoUrl, repoInfo.ref)
	repoCodeqlScanUrl := fmt.Sprintf("%s/security/code-scanning?query=is:open+ref:%s", repoUrl, repoInfo.ref)
	return repoUrl, repoReference, repoCodeqlScanUrl
}
func TestCreateToolRecordCodeql(t *testing.T) {
	t.Run("Valid toolrun file", func(t *testing.T) {
		repoInfo := RepoInfo{serverUrl: "https://github.hello.test", commitId: "test", ref: "refs/head/branch", owner: "Testing", repo: "fortify"}
		repoUrl, repoReference, repoCodeqlScanUrl := getRepoReferences(repoInfo)
		toolRecord, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, repoUrl, repoReference, repoCodeqlScanUrl)
		assert.NoError(t, err)
		assert.Equal(t, toolRecord.ToolName, "codeql")
		assert.Equal(t, toolRecord.ToolInstance, "https://github.hello.test")
		assert.Equal(t, toolRecord.DisplayName, "Testing fortify - refs/head/branch test")
		assert.Equal(t, toolRecord.DisplayURL, "https://github.hello.test/Testing/fortify/security/code-scanning?query=is:open+ref:refs/head/branch")
	})
	t.Run("Empty repository URL", func(t *testing.T) {
		repoInfo := RepoInfo{serverUrl: "", commitId: "test", ref: "refs/head/branch", owner: "Testing", repo: "fortify"}
		repoUrl, repoReference, repoCodeqlScanUrl := getRepoReferences(repoInfo)
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, repoUrl, repoReference, repoCodeqlScanUrl)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "Repository not set")
	})

	t.Run("Empty analyzedRef", func(t *testing.T) {
		repoInfo := RepoInfo{serverUrl: "https://github.hello.test", commitId: "test", ref: "", owner: "Testing", repo: "fortify"}
		repoUrl, repoReference, repoCodeqlScanUrl := getRepoReferences(repoInfo)
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, repoUrl, repoReference, repoCodeqlScanUrl)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "Analyzed Reference not set")
	})

	t.Run("Empty CommitId", func(t *testing.T) {
		repoInfo := RepoInfo{serverUrl: "https://github.hello.test", commitId: "", ref: "refs/head/branch", owner: "Testing", repo: "fortify"}
		repoUrl, repoReference, repoCodeqlScanUrl := getRepoReferences(repoInfo)
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, repoUrl, repoReference, repoCodeqlScanUrl)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "CommitId not set")
	})
	t.Run("Invalid analyzedRef", func(t *testing.T) {
		repoInfo := RepoInfo{serverUrl: "https://github.hello.test", commitId: "", ref: "refs/branch", owner: "Testing", repo: "fortify"}
		repoUrl, repoReference, repoCodeqlScanUrl := getRepoReferences(repoInfo)
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, repoUrl, repoReference, repoCodeqlScanUrl)

		assert.Error(t, err)
	})
}
