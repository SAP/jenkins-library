package cmd

import (
	"context"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/google/go-github/v68/github"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

type githubIssueCommentService interface {
	CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
}

func githubCommentIssue(config githubCommentIssueOptions, telemetryData *telemetry.CustomData) {
	// TODO provide parameter for trusted certs
	ctx, client, err := piperGithub.NewClientBuilder(config.Token, config.APIURL).Build()
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client")
	}
	err = runGithubCommentIssue(ctx, &config, telemetryData, client.Issues)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to comment on issue")
	}
}

func runGithubCommentIssue(ctx context.Context, config *githubCommentIssueOptions, _ *telemetry.CustomData, ghIssueCommentService githubIssueCommentService) error {
	issueComment := github.IssueComment{
		Body: &config.Body,
	}

	newcomment, resp, err := ghIssueCommentService.CreateComment(ctx, config.Owner, config.Repository, config.Number, &issueComment)
	if err != nil {
		log.Entry().Errorf("GitHub response code %v", resp.Status)
		return fmt.Errorf("Error occurred when creating comment on issue %v: %w", config.Number, err)
	}
	log.Entry().Debugf("New issue comment created for issue %v: %v", config.Number, newcomment)

	return nil
}
