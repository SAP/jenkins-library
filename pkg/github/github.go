package github

import (
	"context"
	"net/url"
	"strings"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

// RepositoriesService provides interface for testing
type RepositoriesService interface {
	CreateStatus(ctx context.Context, owner, repo, ref string, status *github.RepoStatus) (*github.RepoStatus, *github.Response, error)
	GetBranchProtection(ctx context.Context, owner, repo, branch string) (*github.Protection, *github.Response, error)
}

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
