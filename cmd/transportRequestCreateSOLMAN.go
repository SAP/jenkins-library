package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/transportrequest/solman"
)

type transportRequestCreateSOLMANUtils interface {
	command.ExecRunner
	GetExitCode() int

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The transportRequestCreateSOLMANUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type transportRequestCreateSOLMANUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to transportRequestCreateSOLMANUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// transportRequestCreateSOLMANUtilsBundle and forward to the implementation of the dependency.
}

func newTransportRequestCreateSOLMANUtils() transportRequestCreateSOLMANUtils {
	utils := transportRequestCreateSOLMANUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func transportRequestCreateSOLMAN(config transportRequestCreateSOLMANOptions, telemetryData *telemetry.CustomData, cpe *transportRequestCreateSOLMANCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTransportRequestCreateSOLMANUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.

	create := &solman.CreateAction{}

	err := runTransportRequestCreateSOLMAN(&config, create, telemetryData, utils, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTransportRequestCreateSOLMAN(
	config *transportRequestCreateSOLMANOptions,
	create solman.Create,
	telemetryData *telemetry.CustomData,
	utils transportRequestCreateSOLMANUtils,
	cpe *transportRequestCreateSOLMANCommonPipelineEnvironment,
) error {

	connection := solman.Connection{
		Endpoint: config.Endpoint,
		User:     config.Username,
		Password: config.Password,
	}

	create.WithConnection(connection)
	create.WithDevelopmentSystemID(config.DevelopmentSystemID)
	create.WithChangeDocumentID(config.ChangeDocumentID)

	transportRequestID, err := create.Perform(utils)

	if err != nil {
		return err
	}
	cpe.custom.transportRequestID = transportRequestID

	return err
}
