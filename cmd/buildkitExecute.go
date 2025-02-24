package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func buildkitExecute(config buildkitExecuteOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{
		ErrorCategoryMapping: map[string][]string{
			log.ErrorBuild.String(): {
				"failed to execute buildctl",
			},
		},
		StepName: "buildkitExecute",
	}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	log.Entry().Info("Starting buildkit execution...")
	log.Entry().Infof("Using Dockerfile at: %s", config.DockerfilePath)

	// Test buildctl command availability
	err := c.RunExecutable("buildctl", "--version")
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to execute buildctl command")
	}

	log.Entry().Info("Buildkit execution completed")
}
