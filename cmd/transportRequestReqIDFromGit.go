package cmd

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/transportrequest"
)

// mocking framework. Allows to redirect the containing methods
type gitIDInRangeFinder interface {
	FindIDInRange(label, from, to string) (string, error)
}

type gitIDInRange struct {
}

func (*gitIDInRange) FindIDInRange(label, from, to string) (string, error) {
	return transportrequest.FindIDInRange(label, from, to)
}

func transportRequestReqIDFromGit(config transportRequestReqIDFromGitOptions,
	telemetryData *telemetry.CustomData,
	commonPipelineEnvironment *transportRequestReqIDFromGitCommonPipelineEnvironment) {

	err := runTransportRequestReqIDFromGit(&config, telemetryData, &gitIDInRange{}, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTransportRequestReqIDFromGit(config *transportRequestReqIDFromGitOptions,
	telemetryData *telemetry.CustomData,
	trUtils gitIDInRangeFinder,
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
	trUtils gitIDInRangeFinder) (string, error) {

	return trUtils.FindIDInRange(config.TransportRequestLabel, config.GitFrom, config.GitTo)
}
