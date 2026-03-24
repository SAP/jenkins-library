package eventing

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
)

// ProcessCDE publishes a CDEvents TaskRunFinished event via GCP Pub/Sub.
func ProcessCDE(tokenProvider gcp.OIDCTokenProvider, generalConfig *config.GeneralConfigOptions, ctx EventContext) error {
	if tokenProvider == nil {
		log.Entry().Warn("event publishing is enabled but no OIDC token provider is available, skipping")
		return nil
	}

	cfg := generalConfig.HookConfig.GCPPubSubConfig

	outcome := "failure"
	if ctx.ErrorCode == "0" {
		outcome = "success"
	}

	// TODO: pass a real pipeline URL (e.g. from orchestrator config) instead of empty string
	eventData, err := newTaskRunFinishedCDEvent(cfg.Source, ctx.StepName, "", outcome, ctx.StageName)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	log.Entry().Debugf("publishing TaskRunFinished CDEvent to GCP Pub/Sub...")

	return publish(tokenProvider, generalConfig, fmt.Sprintf("%spipelinetaskrun-finished", cfg.TopicPrefix), eventData)
}

// Process publishes a plain CloudEvent TaskRunFinished event via GCP Pub/Sub,
// using the original event format with taskName, stageName, and outcome in the data payload.
func Process(tokenProvider gcp.OIDCTokenProvider, generalConfig *config.GeneralConfigOptions, ctx EventContext) error {
	if tokenProvider == nil {
		log.Entry().Warn("event publishing is enabled but no OIDC token provider is available, skipping")
		return nil
	}

	cfg := generalConfig.HookConfig.GCPPubSubConfig

	outcome := "failure"
	if ctx.ErrorCode == "0" {
		outcome = "success"
	}

	eventType := fmt.Sprintf("%seventTypeTaskRunFinished", cfg.TypePrefix)
	eventData, err := newEvent(eventType, cfg.Source, map[string]string{
		"taskName":      ctx.StepName,
		"stageName":     ctx.StageName,
		"outcome":       outcome,
		"pipelineRunId": ctx.PipelineId,
	})
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(eventData), "", "  ")
	log.Entry().Debugf("legacy event payload:\n%s", string(prettyJSON))
	log.Entry().Debugf("publishing TaskRunFinished legacy event to GCP Pub/Sub...")

	return publish(tokenProvider, generalConfig, fmt.Sprintf("%spipelinetaskrun-finished", cfg.TopicPrefix), eventData)
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
