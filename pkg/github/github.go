package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
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

// NewClient creates a new GitHub client using an OAuth token for authentication
func NewClient(token, apiURL, uploadURL string, trustedCerts []string) (context.Context, *github.Client, error) {
	httpClient := piperhttp.Client{}
	httpClient.SetOptions(piperhttp.ClientOptions{
		TrustedCerts:             trustedCerts,
		DoLogRequestBodyOnDebug:  true,
		DoLogResponseBodyOnDebug: true,
	})
	stdClient := httpClient.StandardClient()
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, stdClient)
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token, TokenType: "Bearer"})
	tc := oauth2.NewClient(ctx, ts)

	if !strings.HasSuffix(apiURL, "/") {
		apiURL += "/"
	}
	baseURL, err := url.Parse(apiURL)
	if err != nil {
		return ctx, nil, err
	}

	if !strings.HasSuffix(uploadURL, "/") {
		uploadURL += "/"
	}
	uploadTargetURL, err := url.Parse(uploadURL)
	if err != nil {
		return ctx, nil, err
	}

	client := github.NewClient(tc)

	client.BaseURL = baseURL
	client.UploadURL = uploadTargetURL
	return ctx, client, nil
}

func CreateIssue(ghCreateIssueOptions *CreateIssueOptions) (*github.Issue, error) {
	ctx, client, err := NewClient(ghCreateIssueOptions.Token, ghCreateIssueOptions.APIURL, "", ghCreateIssueOptions.TrustedCerts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get GitHub client")
	}
	return createIssueLocal(ctx, ghCreateIssueOptions, client.Issues, client.Search, client.Issues)
}

func createIssueLocal(ctx context.Context, ghCreateIssueOptions *CreateIssueOptions, ghCreateIssueService githubCreateIssueService, ghSearchIssuesService githubSearchIssuesService, ghCreateCommentService githubCreateCommentService) (*github.Issue, error) {
	issue := github.IssueRequest{
		Title: &ghCreateIssueOptions.Title,
	}
	var bodyString string
	if len(ghCreateIssueOptions.Body) > 0 {
		bodyString = string(ghCreateIssueOptions.Body)
	} else {
		bodyString = ""
	}
	issue.Body = &bodyString
	if len(ghCreateIssueOptions.Assignees) > 0 {
		issue.Assignees = &ghCreateIssueOptions.Assignees
	} else {
		issue.Assignees = &[]string{}
	}

	var existingIssue *github.Issue = nil

	if ghCreateIssueOptions.UpdateExisting {
		existingIssue = ghCreateIssueOptions.Issue
		if existingIssue == nil {
			queryString := fmt.Sprintf("is:open is:issue repo:%v/%v in:title %v", ghCreateIssueOptions.Owner, ghCreateIssueOptions.Repository, ghCreateIssueOptions.Title)
			searchResult, resp, err := ghSearchIssuesService.Issues(ctx, queryString, nil)
			if err != nil {
				if resp != nil {
					log.Entry().Errorf("GitHub search issue returned response code %v", resp.Status)
				}
				return nil, errors.Wrap(err, "error occurred when looking for existing issue")
			} else {
				for _, value := range searchResult.Issues {
					if value != nil && *value.Title == ghCreateIssueOptions.Title {
						existingIssue = value
					}
				}
			}
		}

		if existingIssue != nil {
			comment := &github.IssueComment{Body: issue.Body}
			_, resp, err := ghCreateCommentService.CreateComment(ctx, ghCreateIssueOptions.Owner, ghCreateIssueOptions.Repository, *existingIssue.Number, comment)
			if err != nil {
				if resp != nil {
					log.Entry().Errorf("GitHub create comment returned response code %v", resp.Status)
				}
				return nil, errors.Wrap(err, "error occurred when adding comment to existing issue")
			}
		}
	}

	if existingIssue == nil {
		newIssue, resp, err := ghCreateIssueService.Create(ctx, ghCreateIssueOptions.Owner, ghCreateIssueOptions.Repository, &issue)
		if err != nil {
			if resp != nil {
				log.Entry().Errorf("GitHub create issue returned response code %v", resp.Status)
			}
			return nil, errors.Wrap(err, "error occurred when creating issue")
		}
		log.Entry().Debugf("New issue created: %v", newIssue)
		existingIssue = newIssue
	}

	return existingIssue, nil
}
