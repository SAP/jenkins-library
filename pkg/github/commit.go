package github

import (
	"github.com/pkg/errors"
)

// FetchCommitOptions to configure the lookup
type FetchCommitOptions struct {
	APIURL     string `json:"apiUrl,omitempty"`
	Owner      string `json:"owner,omitempty"`
	Repository string `json:"repository,omitempty"`
	Token      string `json:"token,omitempty"`
	SHA        string `json:"sha,omitempty"`
}

// https://docs.github.com/en/rest/reference/commits#get-a-commit
func FetchCommitStatistics(options *FetchCommitOptions) (int, int, int, int, error) {
	// create GitHub client
	ctx, client, err := NewClient(options.Token, options.APIURL, "")
	if err != nil {
		return 0, 0, 0, 0, errors.Wrap(err, "failed to get GitHub client")
	}
	// fetch commit by SAH
	result, _, err := client.Repositories.GetCommit(ctx, options.Owner, options.Repository, options.SHA)
	if err != nil {
		return 0, 0, 0, 0, errors.Wrap(err, "failed to get GitHub commit")
	}
	return result.Stats.GetAdditions(), result.Stats.GetDeletions(), result.Stats.GetTotal(), len(result.Files), nil
}
