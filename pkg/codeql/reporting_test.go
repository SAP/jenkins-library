package codeql

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

func TestCreateToolRecordCodeql(t *testing.T) {
	modulePath := "./"
	t.Run("Valid toolrun file", func(t *testing.T) {
		repoInfo := &RepoInfo{
			ServerUrl:   "https://github.hello.test",
			CommitId:    "test",
			AnalyzedRef: "refs/heads/branch",
			Owner:       "Testing",
			Repo:        "codeql",
			FullUrl:     "https://github.hello.test/Testing/codeql",
			FullRef:     "https://github.hello.test/Testing/codeql/tree/branch",
			ScanUrl:     "https://github.hello.test/Testing/codeql/security/code-scanning?query=is:open+ref:refs/heads/branch",
		}
		toolRecord, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, modulePath)
		assert.NoError(t, err)
		assert.Equal(t, toolRecord.ToolName, "codeql")
		assert.Equal(t, toolRecord.ToolInstance, "https://github.hello.test")
		assert.Equal(t, toolRecord.DisplayName, "Testing codeql - refs/heads/branch test")
		assert.Equal(t, toolRecord.DisplayURL, "https://github.hello.test/Testing/codeql/security/code-scanning?query=is:open+ref:refs/heads/branch")
	})

	t.Run("Empty repository URL", func(t *testing.T) {
		repoInfo := &RepoInfo{ServerUrl: "", CommitId: "test", AnalyzedRef: "refs/heads/branch", Owner: "Testing", Repo: "codeql"}
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, modulePath)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "Repository not set")
	})

	t.Run("Empty analyzedRef", func(t *testing.T) {
		repoInfo := &RepoInfo{ServerUrl: "https://github.hello.test", CommitId: "test", AnalyzedRef: "", Owner: "Testing", Repo: "codeql"}
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, modulePath)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "Analyzed Reference not set")
	})

	t.Run("Empty CommitId", func(t *testing.T) {
		repoInfo := &RepoInfo{ServerUrl: "https://github.hello.test", CommitId: "", AnalyzedRef: "refs/heads/branch", Owner: "Testing", Repo: "codeql"}
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, modulePath)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "CommitId not set")
	})

	t.Run("Invalid analyzedRef", func(t *testing.T) {
		repoInfo := &RepoInfo{ServerUrl: "https://github.hello.test", CommitId: "", AnalyzedRef: "refs/branch", Owner: "Testing", Repo: "codeql"}
		_, err := createToolRecordCodeql(newCodeqlExecuteScanTestsUtils(), repoInfo, modulePath)

		assert.Error(t, err)
	})
}
