package cmd

import (
	"context"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

type githubCreateIssueService interface {
	Create(ctx context.Context, owner string, repo string, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
}

func githubCreateIssue(config githubCreateIssueOptions, telemetryData *telemetry.CustomData) {
	ctx, client, err := piperGithub.NewClient(config.Token, config.APIURL, "")
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client")
	}
	err = runGithubCreateIssue(ctx, &config, telemetryData, client.Issues)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to comment on issue")
	}
}

func runGithubCreateIssue(ctx context.Context, config *githubCreateIssueOptions, _ *telemetry.CustomData, ghCreateIssueService githubCreateIssueService) error {
	issue := github.IssueRequest{
		Body:  &config.Body,
		Title: &config.Title,
	}

	newIssue, resp, err := ghCreateIssueService.Create(ctx, config.Owner, config.Repository, &issue)
	if err != nil {
		log.Entry().Errorf("GitHub response code %v", resp.Status)
		return errors.Wrap(err, "Error occurred when creating issue")
	}
	log.Entry().Debugf("New issue created: %v", newIssue)

	return nil
}
