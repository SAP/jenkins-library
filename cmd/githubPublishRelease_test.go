package cmd

import (
	"fmt"
	"testing"

	"github.com/google/go-github/v28/github"
	"github.com/stretchr/testify/assert"
)

func TestRunGithubPublishRelease(t *testing.T) {

}

func TestIsExcluded(t *testing.T) {

	l1 := "label1"
	l2 := "label2"

	tt := []struct {
		issue         *github.Issue
		excludeLabels []string
		expected      bool
	}{
		{issue: nil, excludeLabels: nil, expected: false},
		{issue: &github.Issue{}, excludeLabels: nil, expected: false},
		{issue: &github.Issue{Labels: []github.Label{{Name: &l1}}}, excludeLabels: nil, expected: false},
		{issue: &github.Issue{Labels: []github.Label{{Name: &l1}}}, excludeLabels: []string{"label0"}, expected: false},
		{issue: &github.Issue{Labels: []github.Label{{Name: &l1}}}, excludeLabels: []string{"label1"}, expected: true},
		{issue: &github.Issue{Labels: []github.Label{{Name: &l1}, {Name: &l2}}}, excludeLabels: []string{}, expected: false},
		{issue: &github.Issue{Labels: []github.Label{{Name: &l1}, {Name: &l2}}}, excludeLabels: []string{"label1"}, expected: true},
	}

	for k, v := range tt {
		assert.Equal(t, v.expected, isExcluded(v.issue, v.excludeLabels), fmt.Sprintf("Run %v failed", k))
	}

}
