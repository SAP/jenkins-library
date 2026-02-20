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

// GCPPubSubOptions holds the configuration needed to publish task run events via GCP Pub/Sub.
type GCPPubSubOptions struct {
	ProjectNumber    string
	IdentityPool     string
	IdentityProvider string
	Source           string
	TopicPrefix      string
	TypePrefix       string
	CorrelationID    string
	OIDCRoleID       string
}

// PublishTaskRunFinishedEvent constructs and publishes a TaskRunFinished CloudEvent
// via GCP Pub/Sub. It is a no-op if opts.Enabled is false.
func PublishTaskRunFinishedEvent(tokenProvider gcp.OIDCTokenProvider, GeneralConfig config.GeneralConfigOptions, stageName, stepName, errorCode string) error {
	if !GeneralConfig.HookConfig.GCPPubSubConfig.Enabled {
		return nil
	}
	opts := GCPPubSubOptions{
		ProjectNumber:    GeneralConfig.HookConfig.GCPPubSubConfig.ProjectNumber,
		IdentityPool:     GeneralConfig.HookConfig.GCPPubSubConfig.IdentityPool,
		IdentityProvider: GeneralConfig.HookConfig.GCPPubSubConfig.IdentityProvider,
		Source:           GeneralConfig.HookConfig.GCPPubSubConfig.Source,
		TopicPrefix:      GeneralConfig.HookConfig.GCPPubSubConfig.TopicPrefix,
		TypePrefix:       GeneralConfig.HookConfig.GCPPubSubConfig.TypePrefix,
		CorrelationID:    GeneralConfig.CorrelationID,
		OIDCRoleID:       GeneralConfig.HookConfig.OIDCConfig.RoleID,
	}

	if tokenProvider == nil {
		return fmt.Errorf("MSB event publishing is enabled but no OIDC token provider is available")
	}

	log.Entry().Debug("publishing event to GCP Pub/Sub...")

	payload := NewPayloadTaskRunFinished(stageName, stepName, errorCode)

	eventData, err := NewEventTaskRunFinished(opts.TypePrefix, opts.Source, payload)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	topic := fmt.Sprintf("%spipelinetaskrun-finished", opts.TopicPrefix)
	err = gcp.NewGcpPubsubClient(
		tokenProvider,
		opts.ProjectNumber,
		opts.IdentityPool,
		opts.IdentityProvider,
		opts.CorrelationID,
		opts.OIDCRoleID,
	).Publish(topic, eventData)
	if err != nil {
		return fmt.Errorf("event publish failed: %w", err)
	}

	return nil
}
