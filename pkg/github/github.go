package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

//NewClient creates a new GitHub client using an OAuth token for authentication
func NewClient(token, apiURL, uploadURL string) (context.Context, *github.Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
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

func CreateIssue(token, APIURL, owner, repository, title string, body []byte, assignees []string, updateExisting bool) error {

	ctx, client, err := NewClient(token, APIURL, "")
	if err != nil {
		return errors.Wrap(err, "failed to get GitHub client")
	}

	issue := github.IssueRequest{
		Title: &title,
	}
	var bodyString string
	if len(body) > 0 {
		bodyString = string(body)
	} else {
		bodyString = ""
	}
	issue.Body = &bodyString
	if len(assignees) > 0 {
		issue.Assignees = &assignees
	} else {
		issue.Assignees = &[]string{}
	}

	var existingIssue *github.Issue = nil

	if updateExisting {
		queryString := fmt.Sprintf("is:open is:issue repo:%v/%v in:title %v", owner, repository, title)
		searchResult, resp, err := client.Search.Issues(ctx, queryString, nil)
		if err != nil {
			if resp != nil {
				log.Entry().Errorf("GitHub response code %v", resp.Status)
			}
			return errors.Wrap(err, "error occurred when looking for existing issue")
		} else {
			for _, value := range searchResult.Issues {
				if value != nil && *value.Title == title {
					existingIssue = value
				}
			}
		}

		if existingIssue != nil {
			comment := &github.IssueComment{Body: issue.Body}
			_, resp, err := client.Issues.CreateComment(ctx, owner, repository, *existingIssue.Number, comment)
			if err != nil {
				if resp != nil {
					log.Entry().Errorf("GitHub response code %v", resp.Status)
				}
				return errors.Wrap(err, "error occurred when looking for existing issue")
			}
		}
	}

	if existingIssue == nil {
		newIssue, resp, err := client.Issues.Create(ctx, owner, repository, &issue)
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
