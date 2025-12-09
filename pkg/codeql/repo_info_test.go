package codeql

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/stretchr/testify/assert"
)

func TestGetRepoInfo(t *testing.T) {
	t.Run("Valid URL1", func(t *testing.T) {
		repo := "https://github.hello.test/Testing/codeql.git"
		analyzedRef := "refs/heads/branch"
		commitID := "abcd1234"

		repoInfo, err := GetRepoInfo(repo, analyzedRef, commitID, "", "")
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "codeql", repoInfo.Repo)
		assert.Equal(t, "refs/heads/branch", repoInfo.AnalyzedRef)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql", repoInfo.FullUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/security/code-scanning?query=is:open+ref:refs/heads/branch", repoInfo.ScanUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/tree/branch", repoInfo.FullRef)
	})

	t.Run("Valid URL2", func(t *testing.T) {
		repo := "https://github.hello.test/Testing/codeql"
		analyzedRef := "refs/heads/branch"
		commitID := "abcd1234"

		repoInfo, err := GetRepoInfo(repo, analyzedRef, commitID, "", "")
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "codeql", repoInfo.Repo)
		assert.Equal(t, "refs/heads/branch", repoInfo.AnalyzedRef)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql", repoInfo.FullUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/security/code-scanning?query=is:open+ref:refs/heads/branch", repoInfo.ScanUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/tree/branch", repoInfo.FullRef)
	})

	t.Run("Valid url with dots URL1", func(t *testing.T) {
		repo := "https://github.hello.test/Testing/com.sap.codeql.git"
		analyzedRef := "refs/heads/branch"
		commitID := "abcd1234"

		repoInfo, err := GetRepoInfo(repo, analyzedRef, commitID, "", "")
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "com.sap.codeql", repoInfo.Repo)
		assert.Equal(t, "refs/heads/branch", repoInfo.AnalyzedRef)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "https://github.hello.test/Testing/com.sap.codeql", repoInfo.FullUrl)
		assert.Equal(t, "https://github.hello.test/Testing/com.sap.codeql/security/code-scanning?query=is:open+ref:refs/heads/branch", repoInfo.ScanUrl)
		assert.Equal(t, "https://github.hello.test/Testing/com.sap.codeql/tree/branch", repoInfo.FullRef)
	})

	t.Run("Valid url with dots URL2", func(t *testing.T) {
		repo := "https://github.hello.test/Testing/com.sap.codeql"
		analyzedRef := "refs/heads/branch"
		commitID := "abcd1234"

		repoInfo, err := GetRepoInfo(repo, analyzedRef, commitID, "", "")
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "com.sap.codeql", repoInfo.Repo)
		assert.Equal(t, "refs/heads/branch", repoInfo.AnalyzedRef)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "https://github.hello.test/Testing/com.sap.codeql", repoInfo.FullUrl)
		assert.Equal(t, "https://github.hello.test/Testing/com.sap.codeql/security/code-scanning?query=is:open+ref:refs/heads/branch", repoInfo.ScanUrl)
		assert.Equal(t, "https://github.hello.test/Testing/com.sap.codeql/tree/branch", repoInfo.FullRef)
	})

	t.Run("Valid url with username and token URL1", func(t *testing.T) {
		repo := "https://username:token@github.hello.test/Testing/codeql.git"
		analyzedRef := "refs/heads/branch"
		commitID := "abcd1234"

		repoInfo, err := GetRepoInfo(repo, analyzedRef, commitID, "", "")
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "codeql", repoInfo.Repo)
		assert.Equal(t, "refs/heads/branch", repoInfo.AnalyzedRef)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql", repoInfo.FullUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/security/code-scanning?query=is:open+ref:refs/heads/branch", repoInfo.ScanUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/tree/branch", repoInfo.FullRef)
	})

	t.Run("Valid url with username and token URL2", func(t *testing.T) {
		repo := "https://username:token@github.hello.test/Testing/codeql"
		analyzedRef := "refs/heads/branch"
		commitID := "abcd1234"

		repoInfo, err := GetRepoInfo(repo, analyzedRef, commitID, "", "")
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "codeql", repoInfo.Repo)
		assert.Equal(t, "refs/heads/branch", repoInfo.AnalyzedRef)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql", repoInfo.FullUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/security/code-scanning?query=is:open+ref:refs/heads/branch", repoInfo.ScanUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/tree/branch", repoInfo.FullRef)
	})

	t.Run("Invalid URL with no org/reponame", func(t *testing.T) {
		repo := "https://github.hello.test"
		analyzedRef := "refs/heads/branch"
		commitID := "abcd1234"

		repoInfo, err := GetRepoInfo(repo, analyzedRef, commitID, "", "")
		assert.NoError(t, err)
		_, err = orchestrator.GetOrchestratorConfigProvider(nil)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "refs/heads/branch", repoInfo.AnalyzedRef)
		if err != nil {
			assert.Equal(t, "", repoInfo.Owner)
			assert.Equal(t, "", repoInfo.Repo)
			assert.Equal(t, "", repoInfo.ServerUrl)
		}
	})

	t.Run("Non-Github SCM, TargetGithubRepo is not empty", func(t *testing.T) {
		repo := "https://gitlab.test/Testing/codeql.git"
		analyzedRef := "refs/heads/branch"
		commitID := "abcd1234"
		targetGHRepoUrl := "https://github.hello.test/Testing/codeql"

		repoInfo, err := GetRepoInfo(repo, analyzedRef, commitID, targetGHRepoUrl, "")
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "codeql", repoInfo.Repo)
		assert.Equal(t, "refs/heads/branch", repoInfo.AnalyzedRef)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql", repoInfo.FullUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/security/code-scanning?query=is:open+ref:refs/heads/branch", repoInfo.ScanUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/tree/branch", repoInfo.FullRef)

	})

	t.Run("Non-Github SCM, TargetGithubRepo and TargetGithubBranch are not empty", func(t *testing.T) {
		repo := "https://gitlab.test/Testing/codeql.git"
		analyzedRef := "refs/heads/branch"
		commitID := "abcd1234"
		targetGHRepoUrl := "https://github.hello.test/Testing/codeql"
		targetGHRepoBranch := "new-branch"

		repoInfo, err := GetRepoInfo(repo, analyzedRef, commitID, targetGHRepoUrl, targetGHRepoBranch)
		assert.NoError(t, err)
		assert.Equal(t, "abcd1234", repoInfo.CommitId)
		assert.Equal(t, "Testing", repoInfo.Owner)
		assert.Equal(t, "codeql", repoInfo.Repo)
		assert.Equal(t, "refs/heads/new-branch", repoInfo.AnalyzedRef)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql", repoInfo.FullUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/security/code-scanning?query=is:open+ref:refs/heads/new-branch", repoInfo.ScanUrl)
		assert.Equal(t, "https://github.hello.test/Testing/codeql/tree/new-branch", repoInfo.FullRef)

	})
}

func TestBuildRepoReference(t *testing.T) {
	t.Run("Valid AnalyzedRef with branch", func(t *testing.T) {
		repo := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/heads/branch"
		ref, err := buildRepoReference(repo, analyzedRef)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test/Testing/fortify/tree/branch", ref)
	})
	t.Run("Valid AnalyzedRef with PR", func(t *testing.T) {
		repo := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/pull/1/merge"
		ref, err := buildRepoReference(repo, analyzedRef)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test/Testing/fortify/pull/1", ref)
	})
	t.Run("Invalid AnalyzedRef without branch name", func(t *testing.T) {
		repo := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/heads"
		ref, err := buildRepoReference(repo, analyzedRef)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "wrong analyzedRef format")
		assert.Equal(t, "", ref)
	})
	t.Run("Invalid AnalyzedRef without PR id", func(t *testing.T) {
		repo := "https://github.hello.test/Testing/fortify"
		analyzedRef := "refs/pull/merge"
		ref, err := buildRepoReference(repo, analyzedRef)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "wrong analyzedRef format")
		assert.Equal(t, "", ref)
	})
}

func TestSetRepoInfoFromRepoUri(t *testing.T) {
	t.Run("Valid https URL1", func(t *testing.T) {
		var repoInfo RepoInfo
		err := setRepoInfoFromRepoUri("https://github.hello.test/Testing/fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Valid https URL2", func(t *testing.T) {
		var repoInfo RepoInfo
		err := setRepoInfoFromRepoUri("https://github.hello.test/Testing/fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})
	t.Run("Valid https URL1 with dots", func(t *testing.T) {
		var repoInfo RepoInfo
		err := setRepoInfoFromRepoUri("https://github.hello.test/Testing/com.sap.fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "com.sap.fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Valid https URL2 with dots", func(t *testing.T) {
		var repoInfo RepoInfo
		err := setRepoInfoFromRepoUri("https://github.hello.test/Testing/com.sap.fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "com.sap.fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})
	t.Run("Valid https URL1 with username and token", func(t *testing.T) {
		var repoInfo RepoInfo
		err := setRepoInfoFromRepoUri("https://username:token@github.hello.test/Testing/fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Valid https URL2 with username and token", func(t *testing.T) {
		var repoInfo RepoInfo
		err := setRepoInfoFromRepoUri("https://username:token@github.hello.test/Testing/fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Invalid https URL as no org/Owner passed", func(t *testing.T) {
		var repoInfo RepoInfo
		assert.Error(t, setRepoInfoFromRepoUri("https://github.com/fortify", &repoInfo))
	})

	t.Run("Invalid URL as no protocol passed", func(t *testing.T) {
		var repoInfo RepoInfo
		assert.Error(t, setRepoInfoFromRepoUri("github.hello.test/Testing/fortify", &repoInfo))
	})

	t.Run("Valid ssh URL1", func(t *testing.T) {
		var repoInfo RepoInfo
		err := setRepoInfoFromRepoUri("git@github.hello.test/Testing/fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Valid ssh URL2", func(t *testing.T) {
		var repoInfo RepoInfo
		err := setRepoInfoFromRepoUri("git@github.hello.test/Testing/fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})
	t.Run("Valid ssh URL1 with dots", func(t *testing.T) {
		var repoInfo RepoInfo
		err := setRepoInfoFromRepoUri("git@github.hello.test/Testing/com.sap.fortify.git", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "com.sap.fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Valid ssh URL2 with dots", func(t *testing.T) {
		var repoInfo RepoInfo
		err := setRepoInfoFromRepoUri("git@github.hello.test/Testing/com.sap.fortify", &repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.hello.test", repoInfo.ServerUrl)
		assert.Equal(t, "com.sap.fortify", repoInfo.Repo)
		assert.Equal(t, "Testing", repoInfo.Owner)
	})

	t.Run("Invalid ssh URL as no org/Owner passed", func(t *testing.T) {
		var repoInfo RepoInfo
		assert.Error(t, setRepoInfoFromRepoUri("git@github.com/fortify", &repoInfo))
	})
}

func TestSetTargetGithubRepoInfo(t *testing.T) {
	t.Parallel()

	t.Run("Source repo server is github", func(t *testing.T) {
		repoInfo := &RepoInfo{
			ServerUrl: "https://github.com",
			Owner:     "owner",
			Repo:      "repo",
		}
		targetRepo := "https://github.com/target/repo"
		targetBranch := "target-branch"
		err := setTargetGithubRepoInfo(targetRepo, targetBranch, repoInfo)
		assert.Error(t, err)
	})

	t.Run("Success", func(t *testing.T) {
		repoInfo := &RepoInfo{
			ServerUrl:   "https://gitlab.com",
			Owner:       "owner",
			Repo:        "repo",
			AnalyzedRef: "refs/heads/source-branch",
		}
		targetRepo := "https://github.com/target/repo"
		targetBranch := "target-branch"
		err := setTargetGithubRepoInfo(targetRepo, targetBranch, repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.com", repoInfo.ServerUrl)
		assert.Equal(t, "target", repoInfo.Owner)
		assert.Equal(t, "repo", repoInfo.Repo)
		assert.Equal(t, "refs/heads/target-branch", repoInfo.AnalyzedRef)
	})

	t.Run("Empty target branch", func(t *testing.T) {
		repoInfo := &RepoInfo{
			ServerUrl:   "https://gitlab.com",
			Owner:       "owner",
			Repo:        "repo",
			AnalyzedRef: "refs/heads/source-branch",
		}
		targetRepo := "https://github.com/target/repo"
		err := setTargetGithubRepoInfo(targetRepo, "", repoInfo)
		assert.NoError(t, err)
		assert.Equal(t, "https://github.com", repoInfo.ServerUrl)
		assert.Equal(t, "target", repoInfo.Owner)
		assert.Equal(t, "repo", repoInfo.Repo)
		assert.Equal(t, "refs/heads/source-branch", repoInfo.AnalyzedRef)
	})
}

func TestGetFullBranchName(t *testing.T) {
	t.Parallel()

	t.Run("Given short branch name", func(t *testing.T) {
		input := "branch-name"
		assert.Equal(t, "refs/heads/branch-name", getFullBranchName(input))
	})
	t.Run("Given full branch name", func(t *testing.T) {
		input := "refs/heads/branch-name"
		assert.Equal(t, "refs/heads/branch-name", getFullBranchName(input))
	})
}
