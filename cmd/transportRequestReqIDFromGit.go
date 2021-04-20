package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/transportrequest"
)

type transportRequestReqIDFromGitUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The transportRequestReqIDFromGitUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type transportRequestReqIDFromGitUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to transportRequestReqIDFromGitUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// transportRequestReqIDFromGitUtilsBundle and forward to the implementation of the dependency.
}

func newTransportRequestReqIDFromGitUtils() transportRequestReqIDFromGitUtils {
	utils := transportRequestReqIDFromGitUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

// mocking framework. Allows to redirect the containing methods
type iTransportRequestUtils interface {
	FindIDInRange(label, from, to string) (string, error)
}
type transportRequestUtils struct {
}

func (*transportRequestUtils) FindIDInRange(label, from, to string) (string, error) {
	return transportrequest.FindIDInRange(label, from, to)
}

func transportRequestReqIDFromGit(config transportRequestReqIDFromGitOptions,
	telemetryData *telemetry.CustomData,
	commonPipelineEnvironment *transportRequestReqIDFromGitCommonPipelineEnvironment) {

	err := runTransportRequestReqIDFromGit(&config, telemetryData, &transportRequestUtils{}, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTransportRequestReqIDFromGit(config *transportRequestReqIDFromGitOptions,
	telemetryData *telemetry.CustomData,
	trUtils iTransportRequestUtils,
	commonPipelineEnvironment *transportRequestReqIDFromGitCommonPipelineEnvironment) error {

	trID, err := getTransportRequestID(config, trUtils)
	if err != nil {
		return err
	}

	commonPipelineEnvironment.custom.transportRequestID = trID

	log.Entry().Infof("Retrieved transport request ID '%s' from Git.", trID)

	return nil
}

func getTransportRequestID(config *transportRequestReqIDFromGitOptions,
	trUtils iTransportRequestUtils) (string, error) {

	return trUtils.FindIDInRange(config.TransportRequestLabel, config.GitFrom, config.GitTo)
}
