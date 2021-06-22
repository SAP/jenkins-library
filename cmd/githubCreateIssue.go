package cmd

import (
	"context"
	"fmt"
	"io/ioutil"

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
	err = runGithubCreateIssue(ctx, &config, telemetryData, client.Issues, ioutil.ReadFile)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to comment on issue")
	}
}

func runGithubCreateIssue(ctx context.Context, config *githubCreateIssueOptions, _ *telemetry.CustomData, ghCreateIssueService githubCreateIssueService, readFile func(string) ([]byte, error)) error {

	if len(config.Body)+len(config.BodyFilePath) == 0 {
		return fmt.Errorf("either parameter `body` or parameter `bodyFilePath` is required")
	}

	issue := github.IssueRequest{
		Title: &config.Title,
	}

	if len(config.Body) > 0 {
		issue.Body = &config.Body
	} else {
		issueContent, err := readFile(config.BodyFilePath)
		if err != nil {
			return errors.Wrapf(err, "failed to read file '%v'", config.BodyFilePath)
		}
		body := string(issueContent)
		issue.Body = &body
	}

	newIssue, resp, err := ghCreateIssueService.Create(ctx, config.Owner, config.Repository, &issue)
	if err != nil {
		if resp != nil {
			log.Entry().Errorf("GitHub response code %v", resp.Status)
		}
		return errors.Wrap(err, "error occurred when creating issue")
	}
	log.Entry().Debugf("New issue created: %v", newIssue)

	return nil
}
