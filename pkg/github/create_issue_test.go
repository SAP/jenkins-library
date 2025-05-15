//go:build unit
// +build unit

package github

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

type ghCreateIssueMock struct {
	issue      *github.IssueRequest
	issueID    int64
	issueError error
	owner      string
	repo       string
	number     int
	assignees  []string
}

func (g *ghCreateIssueMock) Create(ctx context.Context, owner string, repo string, issue *github.IssueRequest) (*github.Issue, *github.Response, error) {
	g.issue = issue
	g.owner = owner
	g.repo = repo
	g.assignees = *issue.Assignees

	issueResponse := github.Issue{ID: &g.issueID, Title: issue.Title, Body: issue.Body}

	ghRes := github.Response{Response: &http.Response{Status: "200"}}
	if g.issueError != nil {
		ghRes.Status = "401"
	}

	return &issueResponse, &ghRes, g.issueError
}

type ghSearchIssuesMock struct {
	issueID            int64
	issueNumber        int
	issueTitle         string
	issueBody          string
	issuesSearchResult *github.IssuesSearchResult
	issuesSearchError  error
}

func (g *ghSearchIssuesMock) Issues(ctx context.Context, query string, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error) {
	regex := regexp.MustCompile(`.*in:title (?P<Title>(.*))`)
	matches := regex.FindStringSubmatch(query)

	g.issueTitle = matches[1]

	issues := []*github.Issue{
		{
			ID:     &g.issueID,
			Number: &g.issueNumber,
			Title:  &g.issueTitle,
			Body:   &g.issueBody,
		},
	}

	total := len(issues)
	incompleteResults := false

	g.issuesSearchResult = &github.IssuesSearchResult{
		Issues:            issues,
		Total:             &total,
		IncompleteResults: &incompleteResults,
	}

	ghRes := github.Response{Response: &http.Response{Status: "200"}}
	if g.issuesSearchError != nil {
		ghRes.Status = "401"
	}

	return g.issuesSearchResult, &ghRes, g.issuesSearchError
}

type ghCreateCommentMock struct {
	issueComment      *github.IssueComment
	issueNumber       int
	issueCommentError error
}

func (g *ghCreateCommentMock) CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	g.issueComment = comment
	g.issueNumber = number
	ghRes := github.Response{Response: &http.Response{Status: "200"}}
	if g.issueCommentError != nil {
		ghRes.Status = "401"
	}
	return g.issueComment, &ghRes, g.issueCommentError
}

func TestRunGithubCreateIssue(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		// init
		ghCreateIssueService := ghCreateIssueMock{
			issueID: 1,
		}
		ghSearchIssuesMock := ghSearchIssuesMock{
			issueID: 1,
		}
		ghCreateCommentMock := ghCreateCommentMock{}
		config := CreateIssueOptions{
			Owner:      "TEST",
			Repository: "test",
			Body:       []byte("This is my test body"),
			Title:      "This is my title",
			Assignees:  []string{"userIdOne", "userIdTwo"},
		}

		// test
		_, err := createIssueLocal(ctx, &config, &ghCreateIssueService, &ghSearchIssuesMock, &ghCreateCommentMock)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, config.Owner, ghCreateIssueService.owner)
		assert.Equal(t, config.Repository, ghCreateIssueService.repo)
		assert.Equal(t, "This is my test body", ghCreateIssueService.issue.GetBody())
		assert.Equal(t, config.Title, ghCreateIssueService.issue.GetTitle())
		assert.Equal(t, config.Assignees, ghCreateIssueService.issue.GetAssignees())
		assert.Nil(t, ghSearchIssuesMock.issuesSearchResult)
		assert.Nil(t, ghCreateCommentMock.issueComment)
	})

	t.Run("Success update existing", func(t *testing.T) {
		// init
		ghSearchIssuesMock := ghSearchIssuesMock{
			issueID: 1,
		}
		ghCreateCommentMock := ghCreateCommentMock{}
		config := CreateIssueOptions{
			Owner:          "TEST",
			Repository:     "test",
			Body:           []byte("This is my test body"),
			Title:          "This is my title",
			Assignees:      []string{"userIdOne", "userIdTwo"},
			UpdateExisting: true,
		}

		// test
		_, err := createIssueLocal(ctx, &config, nil, &ghSearchIssuesMock, &ghCreateCommentMock)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, ghSearchIssuesMock.issuesSearchResult)
		assert.NotNil(t, ghCreateCommentMock.issueComment)
		assert.Equal(t, config.Title, ghSearchIssuesMock.issueTitle)
		assert.Equal(t, config.Title, *ghSearchIssuesMock.issuesSearchResult.Issues[0].Title)
		assert.Equal(t, "This is my test body", ghCreateCommentMock.issueComment.GetBody())
	})

	t.Run("Success update existing based on instance", func(t *testing.T) {
		// init
		ghSearchIssuesMock := ghSearchIssuesMock{
			issueID: 1,
		}
		ghCreateCommentMock := ghCreateCommentMock{}
		var id int64 = 2
		var number int = 123
		config := CreateIssueOptions{
			Owner:          "TEST",
			Repository:     "test",
			Body:           []byte("This is my test body"),
			Title:          "This is my title",
			Assignees:      []string{"userIdOne", "userIdTwo"},
			UpdateExisting: true,
			Issue: &github.Issue{
				ID:     &id,
				Number: &number,
			},
		}

		// test
		_, err := createIssueLocal(ctx, &config, nil, &ghSearchIssuesMock, &ghCreateCommentMock)

		// assert
		assert.NoError(t, err)
		assert.Nil(t, ghSearchIssuesMock.issuesSearchResult)
		assert.NotNil(t, ghCreateCommentMock.issueComment)
		assert.Equal(t, ghCreateCommentMock.issueNumber, number)
		assert.Equal(t, "This is my test body", ghCreateCommentMock.issueComment.GetBody())
	})

	t.Run("Empty body", func(t *testing.T) {
		// init
		ghCreateIssueService := ghCreateIssueMock{
			issueID: 1,
		}
		ghSearchIssuesMock := ghSearchIssuesMock{
			issueID: 1,
		}
		ghCreateCommentMock := ghCreateCommentMock{}
		config := CreateIssueOptions{
			Owner:          "TEST",
			Repository:     "test",
			Body:           []byte(""),
			Title:          "This is my title",
			Assignees:      []string{"userIdOne", "userIdTwo"},
			UpdateExisting: true,
		}

		// test
		_, err := createIssueLocal(ctx, &config, &ghCreateIssueService, &ghSearchIssuesMock, &ghCreateCommentMock)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, ghSearchIssuesMock.issuesSearchResult)
		assert.NotNil(t, ghCreateCommentMock.issueComment)
		assert.Equal(t, config.Title, ghSearchIssuesMock.issueTitle)
		assert.Equal(t, config.Title, *ghSearchIssuesMock.issuesSearchResult.Issues[0].Title)
		assert.Equal(t, "", ghCreateCommentMock.issueComment.GetBody())
	})

	t.Run("Create error", func(t *testing.T) {
		// init
		ghCreateIssueService := ghCreateIssueMock{
			issueError: fmt.Errorf("error creating issue"),
		}
		config := CreateIssueOptions{
			Body: []byte("test content"),
		}

		// test
		_, err := createIssueLocal(ctx, &config, &ghCreateIssueService, nil, nil)

		// assert
		assert.EqualError(t, err, "error occurred when creating issue: error creating issue")
	})
}
