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

	if err := runGcpPublishEvent(tokenProvider, &cfg); err != nil {
		// do not fail the step
		log.Entry().WithError(err).Warnf("step execution failed")
	}
}

func runGcpPublishEvent(tokenProvider gcp.OIDCTokenProvider, cfg *gcpPublishEventOptions) error {
	provider, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		log.Entry().WithError(err).Warning("Cannot infer config from CI environment")
	}

	data, err := eventing.NewEventFromJSON(cfg.EventType, cfg.EventSource, cfg.EventData, cfg.AdditionalEventData)
	if err != nil {
		return fmt.Errorf("failed to create event data: %w", err)
	}
	log.Entry().Debugf("CloudEvent created: %s", string(data))

	err = gcp.NewGcpPubsubClient(
		tokenProvider,
		cfg.GcpProjectNumber,
		cfg.GcpWorkloadIDentityPool,
		cfg.GcpWorkloadIDentityPoolProvider,
		provider.BuildURL(),
		GeneralConfig.HookConfig.OIDCConfig.RoleID,
	).Publish(cfg.Topic, data)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Entry().Infof("Event published successfully! With topic: %s", cfg.Topic)
	return nil
}
