package eventing

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
)

// Process is the single entry point for publishing events from generated steps.
// It takes an EventContext with step-level data and handles event creation and publishing.
func Process(tokenProvider gcp.OIDCTokenProvider, generalConfig *config.GeneralConfigOptions, ctx EventContext) error {
	if tokenProvider == nil {
		return fmt.Errorf("event publishing is enabled but no OIDC token provider is available")
	}

	cfg := generalConfig.HookConfig.GCPPubSubConfig

	outcome := "failure"
	if ctx.ErrorCode == "0" {
		outcome = "success"
	}

	// TODO: pass a real pipeline URL (e.g. from orchestrator config) instead of empty string
	eventData, err := newTaskRunFinishedCDEvent(cfg.Source, ctx.StepName, "", outcome)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	log.Entry().Debugf("publishing TaskRunFinished event to GCP Pub/Sub...")

	topic := fmt.Sprintf("%spipelinetaskrun-finished", cfg.TopicPrefix)
	publisher := gcp.NewGcpPubsubClient(
		tokenProvider,
		cfg.ProjectNumber,
		cfg.IdentityPool,
		cfg.IdentityProvider,
		generalConfig.CorrelationID,
		generalConfig.HookConfig.OIDCConfig.RoleID,
	)
	if err = publisher.Publish(topic, eventData); err != nil {
		return fmt.Errorf("event publish failed: %w", err)
	}

	return nil
}
