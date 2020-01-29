package cmd

import (
	"github.com/SAP/jenkins-library/pkg/log"
)

func mtaBuild(config mtaBuildOptions, commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment) error {
	log.Entry().WithField("customKey", "customValue").Info("This is how you write a log message with a custom field ...")
	return nil
}
