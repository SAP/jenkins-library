package cmd

import (
	"context"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/go-github/v28/github"
	"github.com/pkg/errors"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

type githubPRService interface {
	Create(ctx context.Context, owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
}

type githubIssueService interface {
	Edit(ctx context.Context, owner string, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
}

func githubCreatePullRequest(myGithubCreatePullRequestOptions githubCreatePullRequestOptions) error {
	ctx, client, err := piperGithub.NewClient(myGithubPublishReleaseOptions.Token, myGithubPublishReleaseOptions.APIURL, myGithubPublishReleaseOptions.UploadURL)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client")
	}

	err = runGithubCreatePullRequest(ctx, &myGithubCreatePullRequestOptions, client.PullRequests, client.Issues)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to publish GitHub release")
	}

	return nil
}

func runGithubCreatePullRequest(ctx context.Context, myGithubCreatePullRequestOptions *githubCreatePullRequestOptions, ghPRService githubPRService, ghIssueService githubIssueService) error {

	prRequest := github.NewPullRequest{
		Title: &myGithubCreatePullRequestOptions.Title,
		Head:  &myGithubCreatePullRequestOptions.Head,
		Base:  &myGithubCreatePullRequestOptions.Base,
		Body:  &myGithubCreatePullRequestOptions.Body,
	}

	newPR, resp, err := ghPRService.Create(ctx, myGithubCreatePullRequestOptions.Owner, myGithubCreatePullRequestOptions.Repository, &prRequest)
	if err != nil {
		log.Entry().Errorf("GitHub response code %v", resp.Status)
		return errors.Wrap(err, "Error occured when creating pull request")
	}
	log.Entry().Debugf("New pull request created: %v", newPR)

	issueRequest := github.IssueRequest{
		Labels:    &myGithubCreatePullRequestOptions.Labels,
		Assignees: &myGithubCreatePullRequestOptions.Assignees,
	}

	updatedPr, resp, err := ghIssueService.Edit(ctx, myGithubCreatePullRequestOptions.Owner, myGithubCreatePullRequestOptions.Repository, newPR.GetNumber(), &issueRequest)
	if err != nil {
		log.Entry().Errorf("GitHub response code %v", resp.Status)
		return errors.Wrap(err, "Error occured when editing pull request")
	}
	log.Entry().Debugf("Updated pull request: %v", updatedPr)

	return nil
}
