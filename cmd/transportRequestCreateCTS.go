package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/transportrequest/cts"
	"github.com/pkg/errors"
)

type transportRequestCreateCTSUtils interface {
	command.ExecRunner
	GetExitCode() int

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The transportRequestCreateCTSUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type transportRequestCreateCTSUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to transportRequestCreateCTSUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// transportRequestCreateCTSUtilsBundle and forward to the implementation of the dependency.
}

func newTransportRequestCreateCTSUtils() transportRequestCreateCTSUtils {
	utils := transportRequestCreateCTSUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func transportRequestCreateCTS(config transportRequestCreateCTSOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *transportRequestCreateCTSCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTransportRequestCreateCTSUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runTransportRequestCreateCTS(&config, telemetryData, utils, commonPipelineEnvironment, &cts.CreateAction{})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTransportRequestCreateCTS(
	config *transportRequestCreateCTSOptions,
	telemetryData *telemetry.CustomData,
	utils transportRequestCreateCTSUtils,
	commonPipelineEnvironment *transportRequestCreateCTSCommonPipelineEnvironment,
	create cts.Create,
) error {
	log.Entry().Infof("Creating transport request at '%s'", config.Endpoint)

	create.WithConnection(
		cts.Connection{
			Endpoint: config.Endpoint,
			User:     config.Username,
			Password: config.Password,
		},
	)

	// "W" will create a customizing request
	// "K" will create a workbench request
	// Not sure what else we have expect, even not sure about two ore more char values ...
	create.WithTransportType(config.TransportType)
	create.WithTargetSystemID(config.TargetSystem)
	create.WithDescription(config.Description)
	create.WithCMOpts(config.CmClientOpts)

	transportRequestID, err := create.Perform(utils)
	if err == nil {
		log.Entry().Infof("Transport request '%s' has been created.", transportRequestID)
		commonPipelineEnvironment.custom.transportRequestID = transportRequestID
	} else {
		log.Entry().Warnf("Creating transport request at '%s' failed", config.Endpoint)
	}

	return errors.Wrapf(err, "cannot create transport at '%s'", config.Endpoint)
}
