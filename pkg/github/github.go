package github

import (
	"context"
	"net/url"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/google/go-github/v45/github"
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

type ClientBuilder struct {
	token        string   // GitHub token, required
	baseURL      string   // GitHub API URL, required
	uploadURL    string   // Base URL for uploading files, optional
	trustedCerts []string // Trusted TLS certificates, optional
}

func NewClientBuilder(token, baseURL string) *ClientBuilder {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &ClientBuilder{
		token:        token,
		baseURL:      baseURL,
		uploadURL:    "",
		trustedCerts: nil,
	}
}

func (b *ClientBuilder) WithTrustedCerts(trustedCerts []string) *ClientBuilder {
	b.trustedCerts = trustedCerts
	return b
}

func (b *ClientBuilder) WithUploadURL(uploadURL string) *ClientBuilder {
	if !strings.HasSuffix(uploadURL, "/") {
		uploadURL += "/"
	}

	b.uploadURL = uploadURL
	return b
}

func (b *ClientBuilder) Build() (context.Context, *github.Client, error) {
	baseURL, err := url.Parse(b.baseURL)
	if err != nil {
		return nil, nil, err
	}

	uploadTargetURL, err := url.Parse(b.uploadURL)
	if err != nil {
		return nil, nil, err
	}

	piperHttp := piperhttp.Client{}
	piperHttp.SetOptions(piperhttp.ClientOptions{
		TrustedCerts:             b.trustedCerts,
		DoLogRequestBodyOnDebug:  true,
		DoLogResponseBodyOnDebug: true,
	})
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, piperHttp.StandardClient())
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: b.token, TokenType: "Bearer"})

	client := github.NewClient(oauth2.NewClient(ctx, tokenSource))
	client.BaseURL = baseURL
	client.UploadURL = uploadTargetURL
	return ctx, client, nil
}
