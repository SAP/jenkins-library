package codeql

import (
	"fmt"
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

func TestBuildRepoReference(t *testing.T) {
	t.Run("Valid Ref with branch", func(t *testing.T) {
		repository := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/head/branch"
		ref, err := BuildRepoReference(repository, analyzedRef)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test/Testing/fortify/tree/branch", ref)
	})
	t.Run("Valid Ref with PR", func(t *testing.T) {
		repository := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/pull/1/merge"
		ref, err := BuildRepoReference(repository, analyzedRef)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test/Testing/fortify/pull/1", ref)
	})
	t.Run("Invalid Ref without branch name", func(t *testing.T) {
		repository := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/head"
		ref, err := BuildRepoReference(repository, analyzedRef)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Wrong analyzedRef format")
		assert.Equal(t, "", ref)
	})
	t.Run("Invalid Ref without PR id", func(t *testing.T) {
		repository := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/pull/merge"
		ref, err := BuildRepoReference(repository, analyzedRef)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Wrong analyzedRef format")
		assert.Equal(t, "", ref)
	})
}

func getRepoReferences(repoInfo RepoInfo) (string, string) {
	repoUrl := fmt.Sprintf("%s/%s/%s", repoInfo.ServerUrl, repoInfo.Owner, repoInfo.Repo)
	repoReference, _ := BuildRepoReference(repoUrl, repoInfo.Ref)
	return repoUrl, repoReference
}

func TestCreateToolRecordCodeql(t *testing.T) {
	modulePath := "./"
	t.Run("Valid toolrun file", func(t *testing.T) {
		repoInfo := RepoInfo{ServerUrl: "https://github.hello.test", CommitId: "test", Ref: "refs/head/branch", Owner: "Testing", Repo: "fortify"}
		repoUrl, repoReference := getRepoReferences(repoInfo)
		toolRecord, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, repoUrl, repoReference, modulePath)
		assert.NoError(t, err)
		assert.Equal(t, toolRecord.ToolName, "codeql")
		assert.Equal(t, toolRecord.ToolInstance, "https://github.hello.test")
		assert.Equal(t, toolRecord.DisplayName, "Testing fortify - refs/head/branch test")
		assert.Equal(t, toolRecord.DisplayURL, "https://github.hello.test/Testing/fortify/security/code-scanning?query=is:open+ref:refs/head/branch")
	})
	t.Run("Empty repository URL", func(t *testing.T) {
		repoInfo := RepoInfo{ServerUrl: "", CommitId: "test", Ref: "refs/head/branch", Owner: "Testing", Repo: "fortify"}
		repoUrl, repoReference := getRepoReferences(repoInfo)
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, repoUrl, repoReference, modulePath)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "Repository not set")
	})

	t.Run("Empty analyzedRef", func(t *testing.T) {
		repoInfo := RepoInfo{ServerUrl: "https://github.hello.test", CommitId: "test", Ref: "", Owner: "Testing", Repo: "fortify"}
		repoUrl, repoReference := getRepoReferences(repoInfo)
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, repoUrl, repoReference, modulePath)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "Analyzed Reference not set")
	})

	t.Run("Empty CommitId", func(t *testing.T) {
		repoInfo := RepoInfo{ServerUrl: "https://github.hello.test", CommitId: "", Ref: "refs/head/branch", Owner: "Testing", Repo: "fortify"}
		repoUrl, repoReference := getRepoReferences(repoInfo)
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, repoUrl, repoReference, modulePath)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "CommitId not set")
	})
	t.Run("Invalid analyzedRef", func(t *testing.T) {
		repoInfo := RepoInfo{ServerUrl: "https://github.hello.test", CommitId: "", Ref: "refs/branch", Owner: "Testing", Repo: "fortify"}
		repoUrl, repoReference := getRepoReferences(repoInfo)
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, repoUrl, repoReference, modulePath)

		assert.Error(t, err)
	})
}
