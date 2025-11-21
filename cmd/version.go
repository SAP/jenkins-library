package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

// TODO: deprecated, remove in future releases
// GitCommit ...
var GitCommit string

// TODO: deprecated, remove in future releases
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
	fmt.Printf("piper-version: %s\n", getVersion())
}

func getVersion() string {
	if build, ok := debug.ReadBuildInfo(); ok && build != nil {
		return build.Main.Version
	}
	return "n/a"
}
