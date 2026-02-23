package events

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
)

type publisher interface {
	Publish(topic string, data []byte) error
}

// PublishTaskRunFinishedEvent constructs and publishes a TaskRunFinished CloudEvent
// via GCP Pub/Sub. It is a no-op if opts.Enabled is false.
func PublishTaskRunFinishedEvent(tokenProvider gcp.OIDCTokenProvider, GeneralConfig config.GeneralConfigOptions, stageName, stepName, errorCode string) error {
	if tokenProvider == nil {
		return fmt.Errorf("MSB event publishing is enabled but no OIDC token provider is available")
	}

	log.Entry().Debug("publishing event to GCP Pub/Sub...")

	payload := NewPayloadTaskRunFinished(stageName, stepName, errorCode)

	eventData, err := NewEventTaskRunFinished(
		GeneralConfig.HookConfig.GCPPubSubConfig.TypePrefix,
		GeneralConfig.HookConfig.GCPPubSubConfig.Source,
		payload,
	)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	topic := fmt.Sprintf("%spipelinetaskrun-finished", GeneralConfig.HookConfig.GCPPubSubConfig.TopicPrefix)
	err = gcp.NewGcpPubsubClient(
		tokenProvider,
		GeneralConfig.HookConfig.GCPPubSubConfig.ProjectNumber,
		GeneralConfig.HookConfig.GCPPubSubConfig.IdentityPool,
		GeneralConfig.HookConfig.GCPPubSubConfig.IdentityProvider,
		GeneralConfig.CorrelationID,
		GeneralConfig.HookConfig.OIDCConfig.RoleID,
	).Publish(topic, eventData)
	if err != nil {
		return fmt.Errorf("event publish failed: %w", err)
	}

	return nil
}
