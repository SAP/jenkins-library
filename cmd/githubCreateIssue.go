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

type githubSearchIssuesService interface {
	Issues(ctx context.Context, query string, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error)
}

type githubCreateCommentService interface {
	CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
}

func githubCreateIssue(config githubCreateIssueOptions, telemetryData *telemetry.CustomData) {
	ctx, client, err := piperGithub.NewClient(config.Token, config.APIURL, "")
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client")
	}
	err = runGithubCreateIssue(ctx, &config, telemetryData, client.Issues, client.Search, client.Issues, ioutil.ReadFile)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to comment on issue")
	}
}

func runGithubCreateIssue(ctx context.Context, config *githubCreateIssueOptions, _ *telemetry.CustomData, ghCreateIssueService githubCreateIssueService, ghSearchIssuesService githubSearchIssuesService, ghCreateCommentService githubCreateCommentService, readFile func(string) ([]byte, error)) error {

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

	if len(config.Assignees) > 0 {
		issue.Assignees = &config.Assignees
	} else {
		issue.Assignees = &[]string{}
	}

	var existingIssue *github.Issue = nil

	if config.UpdateExisting {
		queryString := fmt.Sprintf("is:open is:issue repo:%v/%v in:title %v", config.Owner, config.Repository, config.Title)
		searchResult, resp, err := ghSearchIssuesService.Issues(ctx, queryString, nil)
		if err != nil {
			if resp != nil {
				log.Entry().Errorf("GitHub response code %v", resp.Status)
			}
			return errors.Wrap(err, "error occurred when looking for existing issue")
		} else {
			for _, value := range searchResult.Issues {
				if value != nil && *value.Title == config.Title {
					existingIssue = value
				}
			}
		}

		if existingIssue != nil {
			comment := &github.IssueComment{Body: issue.Body}
			_, resp, err := ghCreateCommentService.CreateComment(ctx, config.Owner, config.Repository, *existingIssue.Number, comment)
			if err != nil {
				if resp != nil {
					log.Entry().Errorf("GitHub response code %v", resp.Status)
				}
				return errors.Wrap(err, "error occurred when looking for existing issue")
			}
		}
	}

	if existingIssue == nil {
		newIssue, resp, err := ghCreateIssueService.Create(ctx, config.Owner, config.Repository, &issue)
		if err != nil {
			if resp != nil {
				log.Entry().Errorf("GitHub response code %v", resp.Status)
			}
			return errors.Wrap(err, "error occurred when creating issue")
		}
		log.Entry().Debugf("New issue created: %v", newIssue)
	}

	return nil
}
