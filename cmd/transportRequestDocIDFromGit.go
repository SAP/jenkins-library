package cmd

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func transportRequestDocIDFromGit(config transportRequestDocIDFromGitOptions,
	telemetryData *telemetry.CustomData,
	commonPipelineEnvironment *transportRequestDocIDFromGitCommonPipelineEnvironment) {

	err := runTransportRequestDocIDFromGit(&config, telemetryData, &gitIDInRange{}, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTransportRequestDocIDFromGit(config *transportRequestDocIDFromGitOptions,
	telemetryData *telemetry.CustomData,
	trUtils gitIDInRangeFinder,
	commonPipelineEnvironment *transportRequestDocIDFromGitCommonPipelineEnvironment) error {

	cdID, err := getChangeDocumentID(config, trUtils)
	if err != nil {
		return err
	}

	commonPipelineEnvironment.custom.changeDocumentID = cdID

	log.Entry().Infof("Retrieved change document ID '%s' from Git.", cdID)

	return nil
}

func getChangeDocumentID(config *transportRequestDocIDFromGitOptions,
	trUtils gitIDInRangeFinder) (string, error) {

	return trUtils.FindIDInRange(config.ChangeDocumentLabel, config.GitFrom, config.GitTo)
}
