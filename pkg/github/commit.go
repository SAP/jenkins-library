package github

import (
	"fmt"

	"github.com/google/go-github/v68/github"
)

// FetchCommitOptions to configure the lookup
type FetchCommitOptions struct {
	APIURL       string   `json:"apiUrl,omitempty"`
	Owner        string   `json:"owner,omitempty"`
	Repository   string   `json:"repository,omitempty"`
	Token        string   `json:"token,omitempty"`
	SHA          string   `json:"sha,omitempty"`
	TrustedCerts []string `json:"trustedCerts,omitempty"`
}

// FetchCommitResult to handle the lookup result
type FetchCommitResult struct {
	Files     int
	Total     int
	Additions int
	Deletions int
}

// https://docs.github.com/en/rest/reference/commits#get-a-commit
// FetchCommitStatistics looks up the statistics for a certain commit SHA.
func FetchCommitStatistics(options *FetchCommitOptions) (FetchCommitResult, error) {
	// create GitHub client
	ctx, client, err := NewClientBuilder(options.Token, options.APIURL).WithTrustedCerts(options.TrustedCerts).Build()
	if err != nil {
		return FetchCommitResult{}, fmt.Errorf("failed to get GitHub client: %w", err)
	}
	// fetch commit by SAH
	result, _, err := client.Repositories.GetCommit(ctx, options.Owner, options.Repository, options.SHA, &github.ListOptions{})
	if err != nil {
		return FetchCommitResult{}, fmt.Errorf("failed to get GitHub commit: %w", err)
	}
	return FetchCommitResult{
		Files:     len(result.Files),
		Total:     result.Stats.GetTotal(),
		Additions: result.Stats.GetAdditions(),
		Deletions: result.Stats.GetDeletions(),
	}, nil
}
