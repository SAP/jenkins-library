package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/events"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/pkg/errors"
)

func gcpPublishEvent(config gcpPublishEventOptions, telemetryData *telemetry.CustomData) {
	err := runGcpPublishEvent(&config, telemetryData)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runGcpPublishEvent(config *gcpPublishEventOptions, _ *telemetry.CustomData) error {
	provider, _ := orchestrator.GetOrchestratorConfigProvider(nil)

	var data []byte
	var err error

	switch config.Type {
	case string(events.PipelineRunStartedEventType):
		data, err = events.ToByteArray(event.EventContextV1{}, events.PipelineRunStartedEventData{
			URL:           provider.BuildURL(),
			CommitId:      provider.CommitSHA(),
			RepositoryURL: provider.RepoURL(),
		})
	case string(events.PipelineRunFinishedEventType):
		data, err = events.ToByteArray(event.EventContextV1{}, events.PipelineRunFinishedEventData{
			URL:           provider.BuildURL(),
			CommitId:      provider.CommitSHA(),
			RepositoryURL: provider.RepoURL(),
			Outcome:       provider.BuildStatus(),
		})
	default:
		return fmt.Errorf("event type %s not supported", config.Type)
	}
	if err != nil {
		return errors.Wrap(err, "failed to create event data")
	}

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
