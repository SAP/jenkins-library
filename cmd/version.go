package cmd

import (
	"fmt"
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

// GitCommit ...
var GitCommit string

// GitTag ...
var GitTag string

// VersionCommand Returns the version of the piper binary
func VersionCommand() *cobra.Command {
	const STEP_NAME = "version"

	var createVersionCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Returns the version of the piper binary",
		Long:  `Writes the commit hash and the tag (if any) to stdout and exits with 0.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			log.SetStepName(STEP_NAME)
			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)
		},
		Run: func(cmd *cobra.Command, args []string) {
			version()
		},
	}

	return createVersionCmd
}

func version() {

	gitCommit, gitTag := "<n/a>", "<n/a>"

	if len(GitCommit) > 0 {
		gitCommit = GitCommit
	}

	if len(GitTag) > 0 {
		gitTag = GitTag
	}

	fmt.Printf("piper-version:\n    commit: \"%s\"\n    tag: \"%s\"\n", gitCommit, gitTag)
}
