package eventing

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	topicPipelineTaskRunFinished = "hyperspace-pipelinetaskrun-finished"
	eventSource                  = "/default/sap.hyperspace.piper"
	eventTypeTaskRunFinished     = "sap.hyperspace.pipelineTaskRunFinished"
)

// PublishTaskRunFinishedCDEvent publishes a CDEvents TaskRunFinished event via GCP Pub/Sub.
func PublishTaskRunFinishedCDEvent(tokenProvider gcp.OIDCTokenProvider, generalConfig *config.GeneralConfigOptions, ctx EventContext) error {
	if tokenProvider == nil {
		log.Entry().Warn("event publishing is enabled but no OIDC token provider is available, skipping")
		return nil
	}

	outcome := "failure"
	if ctx.ErrorCode == "0" {
		outcome = "success"
	}

	// TODO: pass a real pipeline URL (e.g. from orchestrator config) instead of empty string
	eventData, err := newTaskRunFinishedCDEvent(eventSource, ctx.StepName, "", outcome, ctx.StageName)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	log.Entry().Debugf("publishing TaskRunFinished CDEvent to GCP Pub/Sub...")

	return publish(tokenProvider, generalConfig, topicPipelineTaskRunFinished, eventData)
}

// PublishTaskRunFinishedEvent publishes a plain CloudEvent TaskRunFinished event via GCP Pub/Sub,
// using the original event format with taskName, stageName, and outcome in the data payload.
func PublishTaskRunFinishedEvent(tokenProvider gcp.OIDCTokenProvider, generalConfig *config.GeneralConfigOptions, ctx EventContext) error {
	if tokenProvider == nil {
		log.Entry().Warn("event publishing is enabled but no OIDC token provider is available, skipping")
		return nil
	}

	outcome := "failure"
	if ctx.ErrorCode == "0" {
		outcome = "success"
	}

	var fatalError = map[string]any{}
	rawErrorDetail := log.GetFatalErrorDetail()
	if ctx.ErrorCode != "0" && rawErrorDetail != nil {
		// retrieve the error information from the logCollector
		if err := json.Unmarshal(rawErrorDetail, &fatalError); err != nil {
			log.Entry().WithError(err).Warn("could not unmarshal fatal error struct")
		}
	}

	eventData, err := newEvent(eventTypeTaskRunFinished, eventSource, map[string]interface{}{
		"taskName":      ctx.StepName,
		"stageName":     ctx.StageName,
		"outcome":       outcome,
		"pipelineRunId": ctx.PipelineID,
		"errorDetail":   fatalError,
	})
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(eventData), "", "  ")
	log.Entry().Debugf("legacy event payload:\n%s", string(prettyJSON))
	log.Entry().Debugf("publishing TaskRunFinished legacy event to GCP Pub/Sub...")

	return publish(tokenProvider, generalConfig, topicPipelineTaskRunFinished, eventData)
}

func publish(tokenProvider gcp.OIDCTokenProvider, generalConfig *config.GeneralConfigOptions, topic string, eventData []byte) error {
	cfg := generalConfig.HookConfig.GCPPubSubConfig
	publisher := gcp.NewGcpPubsubClient(
		tokenProvider,
		cfg.ProjectNumber,
		cfg.IdentityPool,
		cfg.IdentityProvider,
		generalConfig.CorrelationID,
		generalConfig.HookConfig.OIDCConfig.RoleID,
	)
	if err := publisher.Publish(topic, eventData); err != nil {
		return fmt.Errorf("event publish failed: %w", err)
	}
	return nil
}
