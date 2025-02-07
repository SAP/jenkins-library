//go:build unit
// +build unit

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

type ghIssueCommentMock struct {
	issueComment *github.IssueComment
	issueID      int64
	issueError   error
	owner        string
	repo         string
	number       int
}

func (g *ghIssueCommentMock) CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	g.issueComment = comment
	g.owner = owner
	g.repo = repo
	g.number = number

	issueComment := github.IssueComment{ID: &g.issueID, Body: comment.Body}

	ghRes := github.Response{Response: &http.Response{Status: "200"}}
	if g.issueError != nil {
		ghRes.Status = "401"
	}

	return &issueComment, &ghRes, g.issueError
}

func TestRunGithubCommentIssue(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		// init
		ghIssueCommentService := ghIssueCommentMock{
			issueID: 1,
		}
		config := githubCommentIssueOptions{
			Owner:      "TEST",
			Repository: "test",
			Body:       "This is my test body",
			Number:     1,
		}

		// test
		err := runGithubCommentIssue(ctx, &config, nil, &ghIssueCommentService)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, config.Owner, ghIssueCommentService.owner)
		assert.Equal(t, config.Repository, ghIssueCommentService.repo)
		assert.Equal(t, config.Body, ghIssueCommentService.issueComment.GetBody())
		assert.Equal(t, config.Number, ghIssueCommentService.number)
	})

	t.Run("Error", func(t *testing.T) {
		// init
		ghIssueCommentService := ghIssueCommentMock{
			issueError: fmt.Errorf("error creating comment"),
		}
		config := githubCommentIssueOptions{
			Number: 1,
		}

		// test
		err := runGithubCommentIssue(ctx, &config, nil, &ghIssueCommentService)

		// assert
		assert.EqualError(t, err, "Error occurred when creating comment on issue 1: error creating comment")
	})
}
