package github

import (
	"context"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/go-github/v68/github"
)

// CreateIssueOptions to configure the creation
type CreateIssueOptions struct {
	APIURL         string        `json:"apiUrl,omitempty"`
	Assignees      []string      `json:"assignees,omitempty"`
	Body           []byte        `json:"body,omitempty"`
	Owner          string        `json:"owner,omitempty"`
	Repository     string        `json:"repository,omitempty"`
	Title          string        `json:"title,omitempty"`
	UpdateExisting bool          `json:"updateExisting,omitempty"`
	Token          string        `json:"token,omitempty"`
	TrustedCerts   []string      `json:"trustedCerts,omitempty"`
	Issue          *github.Issue `json:"issue,omitempty"`
}

func CreateIssue(options *CreateIssueOptions) (*github.Issue, error) {
	ctx, client, err := NewClientBuilder(options.Token, options.APIURL).WithTrustedCerts(options.TrustedCerts).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub client: %w", err)
	}
	return createIssueLocal(ctx, options, client.Issues, client.Search, client.Issues)
}

func createIssueLocal(
	ctx context.Context,
	options *CreateIssueOptions,
	createIssueService githubCreateIssueService,
	searchIssuesService githubSearchIssuesService,
	createCommentService githubCreateCommentService,
) (*github.Issue, error) {
	issue := github.IssueRequest{
		Title: &options.Title,
	}
	var bodyString string
	if len(options.Body) > 0 {
		bodyString = string(options.Body)
	} else {
		bodyString = ""
	}
	issue.Body = &bodyString
	if len(options.Assignees) > 0 {
		issue.Assignees = &options.Assignees
	} else {
		issue.Assignees = &[]string{}
	}

	var existingIssue *github.Issue = nil

	if options.UpdateExisting {
		existingIssue = options.Issue
		if existingIssue == nil {
			queryString := fmt.Sprintf("is:open is:issue repo:%v/%v in:title %v", options.Owner, options.Repository, options.Title)
			searchResult, resp, err := searchIssuesService.Issues(ctx, queryString, nil)
			if err != nil {
				if resp != nil {
					log.Entry().Errorf("GitHub search issue returned response code %v", resp.Status)
				}
				return nil, fmt.Errorf("error occurred when looking for existing issue: %w", err)
			} else {
				for _, value := range searchResult.Issues {
					if value != nil && *value.Title == options.Title {
						existingIssue = value
					}
				}
			}
		}

		if existingIssue != nil {
			comment := &github.IssueComment{Body: issue.Body}
			_, resp, err := createCommentService.CreateComment(ctx, options.Owner, options.Repository, *existingIssue.Number, comment)
			if err != nil {
				if resp != nil {
					log.Entry().Errorf("GitHub create comment returned response code %v", resp.Status)
				}
				return nil, fmt.Errorf("error occurred when adding comment to existing issue: %w", err)
			}
		}
	}

	if existingIssue == nil {
		newIssue, resp, err := createIssueService.Create(ctx, options.Owner, options.Repository, &issue)
		if err != nil {
			if resp != nil {
				log.Entry().Errorf("GitHub create issue returned response code %v", resp.Status)
			}
			return nil, fmt.Errorf("error occurred when creating issue: %w", err)
		}
		log.Entry().Debugf("New issue created: %v", newIssue)
		existingIssue = newIssue
	}

	return existingIssue, nil
}
