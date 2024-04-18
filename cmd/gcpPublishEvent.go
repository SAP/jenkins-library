package cmd

import (
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func gcpPublishEvent(config gcpPublishEventOptions, telemetryData *telemetry.CustomData) {
	err := runGcpPublishEvent(&config, telemetryData)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runGcpPublishEvent(config *gcpPublishEventOptions, telemetryData *telemetry.CustomData) error {
	// cdevents.NewCDEvent("")
	// create event
	// pipelineID := ""
	// pipelineURL := ""
	// pipelineName := ""
	// pipelineSource := ""

	// create event data
	data := []byte{}
	// data, err := events.CreatePipelineRunStartedCDEventAsBytes(pipelineID, pipelineName, pipelineSource, pipelineURL)
	// if err != nil {
	// 	return errors.Wrap(err, "failed to create event data")
	// }

	// get federated token
	token, err := gcp.GetFederatedToken(config.GcpProjectNumber, config.GcpWorkloadIDentityPool, config.GcpWorkloadIDentityPoolProvider, config.OIDCToken)
	if err != nil {
		return errors.Wrap(err, "failed to get federated token")
	}

	// publish event
	err = gcp.Publish(config.GcpProjectNumber, config.Topic, token, data)
	if err != nil {
		return errors.Wrap(err, "failed to publish event")
	}

	return nil
}
