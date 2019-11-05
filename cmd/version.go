package cmd

import (
	"fmt"
)

// GitCommit ...
var GitCommit string

// GitTag ...
var GitTag string

func version(myVersionOptions versionOptions) error {

	gitCommit, gitTag := "<n/a>", "<n/a>"

	if len(GitCommit) > 0 {
		gitCommit = GitCommit
	}

	if len(GitTag) > 0 {
		gitTag = GitTag
	}

	_, err := fmt.Printf("piper-version:\n    commit: \"%s\"\n    tag: \"%s\"\n", gitCommit, gitTag)

	return err
}
