package cmd

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

type ghPRMock struct {
	pullrequest *github.NewPullRequest
	prError     error
	owner       string
	repo        string
}

func (g *ghPRMock) Create(ctx context.Context, owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	g.pullrequest = pull
	g.owner = owner
	g.repo = repo
	prNumber := 1
	head := github.PullRequestBranch{Ref: pull.Head}
	base := github.PullRequestBranch{Ref: pull.Base}
	pr := github.PullRequest{Number: &prNumber, Title: pull.Title, Head: &head, Base: &base, Body: pull.Body}

	ghRes := github.Response{Response: &http.Response{Status: "200"}}
	if g.prError != nil {
		ghRes.Status = "401"
	}
	return &pr, &ghRes, g.prError
}

type ghIssueMock struct {
	issueRequest *github.IssueRequest
	issueError   error
	owner        string
	repo         string
	number       int
}

func (g *ghIssueMock) Edit(ctx context.Context, owner string, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error) {
	g.issueRequest = issue
	g.owner = owner
	g.repo = repo
	g.number = number
	labels := []*github.Label{}
	for _, l := range *issue.Labels {
		labels = append(labels, &github.Label{Name: &l})
	}

	assignees := []*github.User{}
	for _, a := range *issue.Assignees {
		assignees = append(assignees, &github.User{Login: &a})
	}

	updatedIssue := github.Issue{Number: &number, Labels: labels, Assignees: assignees}

	ghRes := github.Response{Response: &http.Response{Status: "200"}}
	if g.issueError != nil {
		ghRes.Status = "401"
	}

	return &updatedIssue, &ghRes, g.issueError
}

func TestRunGithubCreatePullRequest(t *testing.T) {
	ctx := context.Background()

	myGithubPROptions := githubCreatePullRequestOptions{
		Owner:      "TEST",
		Repository: "test",
		Title:      "Test Title",
		Body:       "This is the test body.",
		Head:       "head/test",
		Base:       "base/test",
		Labels:     []string{"Test1", "Test2"},
		Assignees:  []string{"User1", "User2"},
	}

	t.Run("Success", func(t *testing.T) {
		ghPRService := ghPRMock{}
		ghIssueService := ghIssueMock{}

		err := runGithubCreatePullRequest(ctx, &myGithubPROptions, &ghPRService, &ghIssueService)
		assert.NoError(t, err, "Error occurred but none expected.")

		assert.Equal(t, myGithubPROptions.Owner, ghPRService.owner, "Owner not passed correctly")
		assert.Equal(t, myGithubPROptions.Repository, ghPRService.repo, "Repository not passed correctly")
		assert.Equal(t, myGithubPROptions.Title, ghPRService.pullrequest.GetTitle(), "Title not passed correctly")
		assert.Equal(t, myGithubPROptions.Body, ghPRService.pullrequest.GetBody(), "Body not passed correctly")
		assert.Equal(t, myGithubPROptions.Head, ghPRService.pullrequest.GetHead(), "Head not passed correctly")
		assert.Equal(t, myGithubPROptions.Base, ghPRService.pullrequest.GetBase(), "Base not passed correctly")

		assert.Equal(t, myGithubPROptions.Owner, ghIssueService.owner, "Owner not passed correctly")
		assert.Equal(t, myGithubPROptions.Repository, ghIssueService.repo, "Repository not passed correctly")
		assert.Equal(t, myGithubPROptions.Labels, ghIssueService.issueRequest.GetLabels(), "Labels not passed correctly")
		assert.Equal(t, myGithubPROptions.Assignees, ghIssueService.issueRequest.GetAssignees(), "Assignees not passed correctly")
		assert.Equal(t, 1, ghIssueService.number, "PR number not passed correctly")
	})

	t.Run("Create error", func(t *testing.T) {
		ghPRService := ghPRMock{prError: fmt.Errorf("Authentication failed")}
		ghIssueService := ghIssueMock{}

		err := runGithubCreatePullRequest(ctx, &myGithubPROptions, &ghPRService, &ghIssueService)
		assert.EqualError(t, err, "Error occurred when creating pull request: Authentication failed", "Wrong error returned")
	})

	t.Run("Edit error", func(t *testing.T) {
		ghPRService := ghPRMock{}
		ghIssueService := ghIssueMock{issueError: fmt.Errorf("Authentication failed")}

		err := runGithubCreatePullRequest(ctx, &myGithubPROptions, &ghPRService, &ghIssueService)
		assert.EqualError(t, err, "Error occurred when editing pull request: Authentication failed", "Wrong error returned")
	})
}
