package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/eventing"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func gcpPublishEvent(cfg gcpPublishEventOptions, telemetryData *telemetry.CustomData) {
	vaultClient := config.GlobalVaultClient()
	var tokenProvider gcp.OIDCTokenProvider
	if vaultClient != nil {
		tokenProvider = vaultClient.GetOIDCTokenByValidation
	}

	provider, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		// do not fail the step
		log.Entry().WithError(err).Warning("Cannot infer config from CI environment")
	}

	publisher := gcp.NewGcpPubsubClient(
		tokenProvider,
		cfg.GcpProjectNumber,
		cfg.GcpWorkloadIDentityPool,
		cfg.GcpWorkloadIDentityPoolProvider,
		provider.BuildURL(),
		GeneralConfig.HookConfig.OIDCConfig.RoleID,
	)

	if err = runGcpPublishEvent(publisher, &cfg); err != nil {
		// do not fail the step
		log.Entry().WithError(err).Warnf("step execution failed")
	}
}

func runGcpPublishEvent(publisher gcp.PubsubClient, cfg *gcpPublishEventOptions) error {
	data, err := eventing.NewEventFromJSON(cfg.EventType, cfg.EventSource, cfg.EventData, cfg.AdditionalEventData)
	if err != nil {
		return fmt.Errorf("failed to create event data: %w", err)
	}
	log.Entry().Debugf("CloudEvent created: %s", string(data))

	if err = publisher.Publish(cfg.Topic, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Entry().Infof("Event published successfully! With topic: %s", cfg.Topic)
	return nil
}
