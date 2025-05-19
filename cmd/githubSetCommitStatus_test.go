//go:build unit

package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

type ghSetCommitRepoService struct {
	serviceError error
	owner        string
	ref          string
	repo         string
	status       *github.RepoStatus
}

func (g *ghSetCommitRepoService) CreateStatus(ctx context.Context, owner, repo, ref string, status *github.RepoStatus) (*github.RepoStatus, *github.Response, error) {
	g.owner = owner
	g.repo = repo
	g.ref = ref
	g.status = status

	return nil, nil, g.serviceError
}

func TestRunGithubSetCommitStatus(t *testing.T) {
	ctx := context.Background()
	telemetryData := telemetry.CustomData{}

	t.Run("success case", func(t *testing.T) {
		config := githubSetCommitStatusOptions{CommitID: "testSha", Context: "test /context", Description: "testDescription", Owner: "testOrg", Repository: "testRepo", Status: "success", TargetURL: "https://test.url"}
		ghRepo := ghSetCommitRepoService{}
		err := runGithubSetCommitStatus(ctx, &config, &telemetryData, &ghRepo)
		expectedStatus := github.RepoStatus{Context: &config.Context, Description: &config.Description, State: &config.Status, TargetURL: &config.TargetURL}
		assert.NoError(t, err)
		assert.Equal(t, config.CommitID, ghRepo.ref)
		assert.Equal(t, config.Owner, ghRepo.owner)
		assert.Equal(t, config.Repository, ghRepo.repo)
		assert.Equal(t, &expectedStatus, ghRepo.status)
	})

	t.Run("error calling GitHub", func(t *testing.T) {
		config := githubSetCommitStatusOptions{CommitID: "testSha", Owner: "testOrg", Repository: "testRepo", Status: "pending"}
		ghRepo := ghSetCommitRepoService{serviceError: fmt.Errorf("gh test error")}
		err := runGithubSetCommitStatus(ctx, &config, &telemetryData, &ghRepo)
		assert.EqualError(t, err, "failed to set status 'pending' on commitId 'testSha': gh test error")
	})
}
