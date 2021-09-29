package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/transportrequest/rfc"
)

type transportRequestUploadRFCUtils interface {
	rfc.Exec
	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The transportRequestUploadRFCUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type transportRequestUploadRFCUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to transportRequestUploadRFCUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// transportRequestUploadRFCUtilsBundle and forward to the implementation of the dependency.
}

func newTransportRequestUploadRFCUtils() transportRequestUploadRFCUtils {
	utils := transportRequestUploadRFCUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func transportRequestUploadRFC(config transportRequestUploadRFCOptions,
	telemetryData *telemetry.CustomData,
	commonPipelineEnvironment *transportRequestUploadRFCCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTransportRequestUploadRFCUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runTransportRequestUploadRFC(&config, &rfc.UploadAction{}, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTransportRequestUploadRFC(config *transportRequestUploadRFCOptions,
	action rfc.Upload,
	telemetryData *telemetry.CustomData,
	utils rfc.Exec,
	commonPipelineEnvironment *transportRequestUploadRFCCommonPipelineEnvironment) error {

	action.WithConnection(
		rfc.Connection{
			Endpoint: config.Endpoint,
			Client:   config.Client,
			Instance: config.Instance,
			User:     config.Username,
			Password: config.Password,
		},
	)
	action.WithApplication(
		rfc.Application{
			Name:        config.ApplicationName,
			Description: config.ApplicationDescription,
			AbapPackage: config.AbapPackage,
		},
	)
	action.WithConfiguration(
		rfc.UploadConfig{
			AcceptUnixStyleEndOfLine: config.AcceptUnixStyleLineEndings,
			CodePage:                 config.CodePage,
			FailUploadOnWarning:      config.FailUploadOnWarning,
			Verbose:                  GeneralConfig.Verbose,
		},
	)
	action.WithTransportRequestID(config.TransportRequestID)
	action.WithApplicationURL(config.ApplicationURL)

	commonPipelineEnvironment.custom.transportRequestID = config.TransportRequestID

	err := action.Perform(utils)

	if err == nil {
		log.Entry().Infof("Upload of artifact '%s' to ABAP backend succeeded (TransportRequestId: '%s').",
			config.ApplicationURL,
			config.TransportRequestID,
		)
	}
	return err
}
