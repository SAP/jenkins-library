package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
	"os"
)

type testExecuteAltUtils struct {
	exec  command.ExecRunner
	files piperutils.FileUtils

	removeAll  func(path string) error
	fileExists func(path string) (bool, error)
}

func newTestExecuteAltUtils() *testExecuteAltUtils {
	utils := testExecuteAltUtils{
		exec:  &command.Command{},
		files: &piperutils.Files{},

		removeAll:  os.RemoveAll,
		fileExists: piperutils.FileExists,
	}
	// Reroute command output to logging framework
	utils.exec.Stdout(log.Writer())
	utils.exec.Stderr(log.Writer())
	return &utils
}

func testExecuteAlt(config testExecuteOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTestExecuteAltUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runTestExecuteAlt(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTestExecuteAlt(config *testExecuteOptions, telemetryData *telemetry.CustomData, utils *testExecuteAltUtils) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

	// Example of calling methods from external dependencies directly on utils:
	exists, err := utils.files.FileExists("file.txt")
	if err != nil {
		// It is good practice to set an error category.
		// Most likely you want to do this at the place where enough context is known.
		log.SetErrorCategory(log.ErrorConfiguration)
		// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
		return fmt.Errorf("failed to check for important file: %w", err)
	}

	// alternatively:
	exists, err = utils.fileExists("file.txt")

	if !exists {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("cannot run without important file")
	}

	gav, err := versioning.GetArtifact("maven", "pom.xml", nil, utils.exec)
	if err != nil {
		return fmt.Errorf("failed to get artifact descriptor: %w", err)
	}
	log.Entry().Infof("found artifact: %v", gav)

	return nil
}
