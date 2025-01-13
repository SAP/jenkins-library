package reporting

import (
	"context"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/go-github/v68/github"
)

type githubIssueService interface {
	Create(ctx context.Context, owner string, repo string, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
	CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
	Edit(ctx context.Context, owner string, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
}

type githubSearchService interface {
	Issues(ctx context.Context, query string, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error)
}

// GitHub contains metadata for reporting towards GitHub
type GitHub struct {
	Owner         *string
	Repository    *string
	Assignees     *[]string
	IssueService  githubIssueService
	SearchService githubSearchService
}

// UploadSingleReport uploads a single report to GitHub
func (g *GitHub) UploadSingleReport(ctx context.Context, scanReport IssueDetail) error {
	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	title := scanReport.Title()
	markdownReport, _ := scanReport.ToMarkdown()

	log.Entry().Debugf("Creating/updating GitHub issue with title %v in org %v and repo %v", title, &g.Owner, &g.Repository)
	if err := g.createIssueOrUpdateIssueComment(ctx, title, string(markdownReport)); err != nil {
		return fmt.Errorf("failed to upload results for '%v' into GitHub issue: %w", title, err)
	}
	return nil
}

// UploadMultipleReports uploads a number of reports to GitHub, one per IssueDetail to create transparency
func (g *GitHub) UploadMultipleReports(ctx context.Context, scanReports *[]IssueDetail) error {
	for _, scanReport := range *scanReports {
		if err := g.UploadSingleReport(ctx, scanReport); err != nil {
			return err
		}
	}
	return nil
}

func (g *GitHub) createIssueOrUpdateIssueComment(ctx context.Context, title, issueContent string) error {
	// check if issue is existing
	issueNumber, issueBody, err := g.findExistingIssue(ctx, title)
	if err != nil {
		return fmt.Errorf("error when looking up issue: %w", err)
	}

	if issueNumber == 0 {
		// issue not existing need to create it
		issue := github.IssueRequest{Title: &title, Body: &issueContent, Assignees: g.Assignees}
		if _, _, err := g.IssueService.Create(ctx, *g.Owner, *g.Repository, &issue); err != nil {
			return fmt.Errorf("failed to create issue: %w", err)
		}
		return nil
	}

	// let's compare and only update in case an update is required
	if issueContent != issueBody {
		// update of issue required
		issueRequest := github.IssueRequest{Body: &issueContent}
		if _, _, err := g.IssueService.Edit(ctx, *g.Owner, *g.Repository, issueNumber, &issueRequest); err != nil {
			return fmt.Errorf("failed to edit issue: %w", err)
		}

		// now add a small comment that the issue content has been updated
		updateText := "issue content has been updated"
		updateComment := github.IssueComment{Body: &updateText}
		if _, _, err := g.IssueService.CreateComment(ctx, *g.Owner, *g.Repository, issueNumber, &updateComment); err != nil {
			return fmt.Errorf("failed to create comment: %w", err)
		}
	}
	return nil
}

func (g *GitHub) findExistingIssue(ctx context.Context, title string) (int, string, error) {
	queryString := fmt.Sprintf("is:issue repo:%v/%v in:title %v", *g.Owner, *g.Repository, title)
	searchResult, _, err := g.SearchService.Issues(ctx, queryString, nil)
	if err != nil {
		return 0, "", fmt.Errorf("error occurred when looking for existing issue: %w", err)
	}
	for _, i := range searchResult.Issues {
		if i != nil && *i.Title == title {
			if i.GetState() == "closed" {
				// reopen issue
				open := "open"
				ir := github.IssueRequest{State: &open}
				if _, _, err := g.IssueService.Edit(ctx, *g.Owner, *g.Repository, i.GetNumber(), &ir); err != nil {
					return i.GetNumber(), "", fmt.Errorf("failed to re-open issue: %w", err)
				}
			}
			return i.GetNumber(), i.GetBody(), nil
		}
	}
	return 0, "", nil
}
