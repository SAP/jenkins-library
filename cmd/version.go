package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/telemetry"
)

// GitCommit ...
var GitCommit string

// GitTag ...
var GitTag string

func version(config versionOptions, telemetryData *telemetry.CustomData) {

	gitCommit, gitTag := "<n/a>", "<n/a>"

	if len(GitCommit) > 0 {
		gitCommit = GitCommit
	}

	if len(GitTag) > 0 {
		gitTag = GitTag
	}

	fmt.Printf("piper-version:\n    commit: \"%s\"\n    tag: \"%s\"\n", gitCommit, gitTag)
}
