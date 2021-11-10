package cmd

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

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
	issueCommentError error
}

func (g *ghCreateCommentMock) CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	g.issueComment = comment
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
		filesMock := mock.FilesMock{}
		ghCreateIssueService := ghCreateIssueMock{
			issueID: 1,
		}
		ghSearchIssuesMock := ghSearchIssuesMock{
			issueID: 1,
		}
		ghCreateCommentMock := ghCreateCommentMock{}
		config := githubCreateIssueOptions{
			Owner:      "TEST",
			Repository: "test",
			Body:       "This is my test body",
			Title:      "This is my title",
			Assignees:  []string{"userIdOne", "userIdTwo"},
		}

		// test
		err := runGithubCreateIssue(ctx, &config, nil, &ghCreateIssueService, &ghSearchIssuesMock, &ghCreateCommentMock, filesMock.FileRead)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, config.Owner, ghCreateIssueService.owner)
		assert.Equal(t, config.Repository, ghCreateIssueService.repo)
		assert.Equal(t, config.Body, ghCreateIssueService.issue.GetBody())
		assert.Equal(t, config.Title, ghCreateIssueService.issue.GetTitle())
		assert.Equal(t, config.Assignees, ghCreateIssueService.issue.GetAssignees())
		assert.Nil(t, ghSearchIssuesMock.issuesSearchResult)
		assert.Nil(t, ghCreateCommentMock.issueComment)
	})

	t.Run("Success - body from file", func(t *testing.T) {
		// init
		filesMock := mock.FilesMock{}
		filesMock.AddFile("test.md", []byte("Test markdown"))
		ghCreateIssueService := ghCreateIssueMock{
			issueID: 1,
		}
		config := githubCreateIssueOptions{
			Owner:        "TEST",
			Repository:   "test",
			BodyFilePath: "test.md",
			Title:        "This is my title",
		}

		// test
		err := runGithubCreateIssue(ctx, &config, nil, &ghCreateIssueService, nil, nil, filesMock.FileRead)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, config.Owner, ghCreateIssueService.owner)
		assert.Equal(t, config.Repository, ghCreateIssueService.repo)
		assert.Equal(t, "Test markdown", ghCreateIssueService.issue.GetBody())
		assert.Equal(t, config.Title, ghCreateIssueService.issue.GetTitle())
		assert.Empty(t, ghCreateIssueService.issue.GetAssignees())
	})

	t.Run("Success update existing", func(t *testing.T) {
		// init
		filesMock := mock.FilesMock{}
		ghSearchIssuesMock := ghSearchIssuesMock{
			issueID: 1,
		}
		ghCreateCommentMock := ghCreateCommentMock{}
		config := githubCreateIssueOptions{
			Owner:          "TEST",
			Repository:     "test",
			Body:           "This is my test body",
			Title:          "This is my title",
			Assignees:      []string{"userIdOne", "userIdTwo"},
			UpdateExisting: true,
		}

		// test
		err := runGithubCreateIssue(ctx, &config, nil, nil, &ghSearchIssuesMock, &ghCreateCommentMock, filesMock.FileRead)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, ghSearchIssuesMock.issuesSearchResult)
		assert.NotNil(t, ghCreateCommentMock.issueComment)
		assert.Equal(t, config.Title, ghSearchIssuesMock.issueTitle)
		assert.Equal(t, config.Title, *ghSearchIssuesMock.issuesSearchResult.Issues[0].Title)
		assert.Equal(t, config.Body, ghCreateCommentMock.issueComment.GetBody())
	})

	t.Run("Error", func(t *testing.T) {
		// init
		filesMock := mock.FilesMock{}
		ghCreateIssueService := ghCreateIssueMock{
			issueError: fmt.Errorf("error creating issue"),
		}
		config := githubCreateIssueOptions{
			Body: "test content",
		}

		// test
		err := runGithubCreateIssue(ctx, &config, nil, &ghCreateIssueService, nil, nil, filesMock.FileRead)

		// assert
		assert.EqualError(t, err, "error occurred when creating issue: error creating issue")
	})

	t.Run("Error - missing issue body", func(t *testing.T) {
		// init
		filesMock := mock.FilesMock{}
		ghCreateIssueService := ghCreateIssueMock{}
		ghSearchIssuesMock := ghSearchIssuesMock{}
		ghCreateCommentMock := ghCreateCommentMock{}
		config := githubCreateIssueOptions{}

		// test
		err := runGithubCreateIssue(ctx, &config, nil, &ghCreateIssueService, &ghSearchIssuesMock, &ghCreateCommentMock, filesMock.FileRead)

		// assert
		assert.EqualError(t, err, "either parameter `body` or parameter `bodyFilePath` is required")
	})

	t.Run("Error - missing body file", func(t *testing.T) {
		// init
		filesMock := mock.FilesMock{}
		ghCreateIssueService := ghCreateIssueMock{}
		config := githubCreateIssueOptions{
			BodyFilePath: "test.md",
		}

		// test
		err := runGithubCreateIssue(ctx, &config, nil, &ghCreateIssueService, nil, nil, filesMock.FileRead)

		// assert
		assert.Contains(t, fmt.Sprint(err), "failed to read file 'test.md'")
	})
}
