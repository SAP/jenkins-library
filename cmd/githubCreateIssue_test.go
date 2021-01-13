package cmd

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

type ghCreateIssueMock struct {
	issue      *github.IssueRequest
	issueID    int64
	issueError error
	owner      string
	repo       string
	number     int
}

func (g *ghCreateIssueMock) Create(ctx context.Context, owner string, repo string, issue *github.IssueRequest) (*github.Issue, *github.Response, error) {
	g.issue = issue
	g.owner = owner
	g.repo = repo

	issueResponse := github.Issue{ID: &g.issueID, Title: issue.Title, Body: issue.Body}

	ghRes := github.Response{Response: &http.Response{Status: "200"}}
	if g.issueError != nil {
		ghRes.Status = "401"
	}

	return &issueResponse, &ghRes, g.issueError
}

func TestRunGithubCreateIssue(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		// init
		ghCreateIssueService := ghCreateIssueMock{
			issueID: 1,
		}
		config := githubCreateIssueOptions{
			Owner:      "TEST",
			Repository: "test",
			Body:       "This is my test body",
			Title:      "This is my title",
		}

		// test
		err := runGithubCreateIssue(ctx, &config, nil, &ghCreateIssueService)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, config.Owner, ghCreateIssueService.owner)
		assert.Equal(t, config.Repository, ghCreateIssueService.repo)
		assert.Equal(t, config.Body, ghCreateIssueService.issue.GetBody())
		assert.Equal(t, config.Title, ghCreateIssueService.issue.GetTitle())
	})

	t.Run("Error", func(t *testing.T) {
		// init
		ghCreateIssueService := ghCreateIssueMock{
			issueError: fmt.Errorf("error creating issue"),
		}
		config := githubCreateIssueOptions{}

		// test
		err := runGithubCreateIssue(ctx, &config, nil, &ghCreateIssueService)

		// assert
		assert.EqualError(t, err, "Error occurred when creating issue: error creating issue")
	})
}
