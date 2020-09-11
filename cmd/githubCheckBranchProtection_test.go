package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

type ghRepoService struct {
	protection   github.Protection
	serviceError error
	owner        string
	repo         string
	branch       string
}

func (g *ghRepoService) GetBranchProtection(ctx context.Context, owner, repo, branch string) (*github.Protection, *github.Response, error) {
	g.owner = owner
	g.repo = repo
	g.branch = branch

	return &g.protection, nil, g.serviceError
}

func TestRunGithubCheckBranchProtection(t *testing.T) {
	ctx := context.Background()
	telemetryData := telemetry.CustomData{}

	t.Run("no checks active", func(t *testing.T) {
		config := githubCheckBranchProtectionOptions{Branch: "testBranch", Owner: "testOrg", Repository: "testRepo"}
		ghRepo := ghRepoService{}
		err := runGithubCheckBranchProtection(ctx, &config, &telemetryData, &ghRepo)
		assert.NoError(t, err)
		assert.Equal(t, config.Branch, ghRepo.branch)
		assert.Equal(t, config.Owner, ghRepo.owner)
		assert.Equal(t, config.Repository, ghRepo.repo)
	})

	t.Run("error calling GitHub", func(t *testing.T) {
		config := githubCheckBranchProtectionOptions{Branch: "testBranch", Owner: "testOrg", Repository: "testRepo"}
		ghRepo := ghRepoService{serviceError: fmt.Errorf("gh test error")}
		err := runGithubCheckBranchProtection(ctx, &config, &telemetryData, &ghRepo)
		assert.EqualError(t, err, "failed to read branch protection information: gh test error")
	})

	t.Run("all checks ok", func(t *testing.T) {
		config := githubCheckBranchProtectionOptions{
			Branch:                       "testBranch",
			Owner:                        "testOrg",
			Repository:                   "testRepo",
			RequiredChecks:               []string{"check1", "check2"},
			RequireEnforceAdmins:         true,
			RequiredApprovingReviewCount: 1,
		}
		ghRepo := ghRepoService{protection: github.Protection{
			RequiredStatusChecks:       &github.RequiredStatusChecks{Contexts: []string{"check0", "check1", "check2", "check3"}},
			EnforceAdmins:              &github.AdminEnforcement{Enabled: true},
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{RequiredApprovingReviewCount: 1},
		}}
		err := runGithubCheckBranchProtection(ctx, &config, &telemetryData, &ghRepo)
		assert.NoError(t, err)
		assert.Equal(t, config.Branch, ghRepo.branch)
		assert.Equal(t, config.Owner, ghRepo.owner)
		assert.Equal(t, config.Repository, ghRepo.repo)
	})

	t.Run("status check missing", func(t *testing.T) {
		config := githubCheckBranchProtectionOptions{
			RequiredChecks: []string{"check1", "check2"},
		}
		ghRepo := ghRepoService{protection: github.Protection{
			RequiredStatusChecks: &github.RequiredStatusChecks{Contexts: []string{"check0", "check1"}},
		}}
		err := runGithubCheckBranchProtection(ctx, &config, &telemetryData, &ghRepo)
		assert.Contains(t, fmt.Sprint(err), "required status check 'check2' not found")
	})

	t.Run("admin enforcement inactive", func(t *testing.T) {
		config := githubCheckBranchProtectionOptions{
			RequireEnforceAdmins: true,
		}
		ghRepo := ghRepoService{protection: github.Protection{
			EnforceAdmins: &github.AdminEnforcement{Enabled: false},
		}}
		err := runGithubCheckBranchProtection(ctx, &config, &telemetryData, &ghRepo)
		assert.Contains(t, fmt.Sprint(err), "admins are not enforced")
	})

	t.Run("not enough reviewers", func(t *testing.T) {
		config := githubCheckBranchProtectionOptions{
			RequiredApprovingReviewCount: 2,
		}
		ghRepo := ghRepoService{protection: github.Protection{
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{RequiredApprovingReviewCount: 1},
		}}
		err := runGithubCheckBranchProtection(ctx, &config, &telemetryData, &ghRepo)
		assert.Contains(t, fmt.Sprint(err), "not enough mandatory reviewers")
	})

}
