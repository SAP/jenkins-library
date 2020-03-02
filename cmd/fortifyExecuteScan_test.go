package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeterminePullRequestMerge(t *testing.T) {
	config := fortifyExecuteScanOptions{CommitMessage: "Merge pull request #2462 from branch f-test", PullRequestMessageRegex: `(?m).*Merge pull request #(\d+) from.*`, PullRequestMessageRegexGroup: 1}

	t.Run("success", func(t *testing.T) {
		match := determinePullRequestMerge(config)
		assert.Equal(t, "2462", match, "Expected different result")
	})

	t.Run("no match", func(t *testing.T) {
		config.CommitMessage = "Some test commit"
		match := determinePullRequestMerge(config)
		assert.Equal(t, "", match, "Expected different result")
	})
}
