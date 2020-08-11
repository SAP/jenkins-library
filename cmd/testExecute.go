package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type testExecuteUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces,
	// for everything you need to be able to mock in tests.
	// Unit tests shall
	//  - not depend on global state,
	//  - be executable in parallel,
	//  - and don't (re-)test dependencies.
}

type testExecuteUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to testExecuteUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
}

func newTestExecuteUtils() testExecuteUtils {
	utils := testExecuteUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func testExecute(config testExecuteOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTestExecuteUtils()

	// For http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runTestExecute(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTestExecute(config *testExecuteOptions, telemetryData *telemetry.CustomData, utils testExecuteUtils) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

	exists, err := utils.FileExists("file.txt")
	if err != nil {
		return fmt.Errorf("failed to check for important file: %w", err)
	}
	if !exists {
		return fmt.Errorf("cannot run without important file")
	}

	return nil
}
