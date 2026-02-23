package eventing

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
)

// PublishTaskRunFinishedEvent creates and publishes a TaskRunFinished CloudEvent via GCP Pub/Sub.
func PublishTaskRunFinishedEvent(tokenProvider gcp.OIDCTokenProvider, generalConfig config.GeneralConfigOptions, stageName, stepName, errorCode string) error {
	if tokenProvider == nil {
		return fmt.Errorf("event publishing is enabled but no OIDC token provider is available")
	}

	cfg := generalConfig.HookConfig.GCPPubSubConfig

	outcome := "failure"
	if errorCode == "0" {
		outcome = "success"
	}

	// TODO: pass a real pipeline URL (e.g. from orchestrator config) instead of empty string
	eventData, err := NewTaskRunFinishedCDEvent(cfg.Source, stepName, "", outcome)
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
