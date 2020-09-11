package github

import (
	"context"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

//NewClient creates a new GitHub client using an OAuth token for authentication
func NewClient(token, apiURL, uploadURL string) (context.Context, *github.Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client, err := github.NewEnterpriseClient(apiURL, uploadURL, tc)
	if err != nil {
		return ctx, nil, err
	}
	return ctx, client, nil
}
