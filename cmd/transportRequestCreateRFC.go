package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/transportrequest/rfc"
	"github.com/pkg/errors"
	"io"
	"os"
)

type transportRequestCreateRFCUtils interface {
	command.ExecRunner
	GetExitCode() int

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The transportRequestCreateRFCUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type transportRequestCreateRFCUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to transportRequestCreateRFCUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// transportRequestCreateRFCUtilsBundle and forward to the implementation of the dependency.
}

func newTransportRequestCreateRFCUtils() transportRequestCreateRFCUtils {
	utils := transportRequestCreateRFCUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func transportRequestCreateRFC(config transportRequestCreateRFCOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *transportRequestCreateRFCCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTransportRequestCreateRFCUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runTransportRequestCreateRFC(&config, &rfc.CreateAction{}, telemetryData, utils, commonPipelineEnvironment, os.Stdout)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTransportRequestCreateRFC(
	config *transportRequestCreateRFCOptions,
	create rfc.Create,
	telemetryData *telemetry.CustomData,
	utils transportRequestCreateRFCUtils,
	commonPipelineEnvironment *transportRequestCreateRFCCommonPipelineEnvironment,
	stdout io.Writer,
) error {

	create.WithConnection(
		rfc.Connection{
			Endpoint: config.Endpoint,
			User:     config.Username,
			Password: config.Password,
			Client:   config.Client,
			Instance: config.Instance,
		},
	)
	create.WithTransportType(config.TransportType)
	create.WithTargetSystemID(config.TargetSystem)
	create.WithDescription(config.Description)
	transportRequestID, err := create.Perform(utils)

	if err == nil {
		if len(transportRequestID) == 0 {
			err = errors.New("No transport requestId received.")
		} else {
			commonPipelineEnvironment.custom.transportRequestID = transportRequestID
			stdout.Write([]byte(transportRequestID))
		}
	}
	return err
}
