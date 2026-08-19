package cmd

import (
	"context"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/google/go-github/v68/github"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

type githubPRService interface {
	Create(ctx context.Context, owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
}

type githubIssueService interface {
	Edit(ctx context.Context, owner string, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
}

func githubCreatePullRequest(config githubCreatePullRequestOptions, telemetryData *telemetry.CustomData) {
	// TODO provide parameter for trusted certs
	ctx, client, err := piperGithub.NewClientBuilder(config.Token, config.APIURL).Build()
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client")
	}

	err = runGithubCreatePullRequest(ctx, &config, client.PullRequests, client.Issues)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to create GitHub pull request")
	}
}

func runGithubCreatePullRequest(ctx context.Context, config *githubCreatePullRequestOptions, ghPRService githubPRService, ghIssueService githubIssueService) error {
	prRequest := github.NewPullRequest{
		Title: &config.Title,
		Head:  &config.Head,
		Base:  &config.Base,
		Body:  &config.Body,
	}

	newPR, resp, err := ghPRService.Create(ctx, config.Owner, config.Repository, &prRequest)
	if err != nil {
		log.Entry().Errorf("GitHub response code %v", resp.Status)
		return fmt.Errorf("Error occurred when creating pull request: %w", err)
	}
	log.Entry().Debugf("New pull request created: %v", newPR)

	issueRequest := github.IssueRequest{
		Labels:    &config.Labels,
		Assignees: &config.Assignees,
	}

	updatedPr, resp, err := ghIssueService.Edit(ctx, config.Owner, config.Repository, newPR.GetNumber(), &issueRequest)
	if err != nil {
		log.Entry().Errorf("GitHub response code %v", resp.Status)
		return fmt.Errorf("Error occurred when editing pull request: %w", err)
	}
	log.Entry().Debugf("Updated pull request: %v", updatedPr)

	return nil
}
