package cmd

import (
	"encoding/json"
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
	if vaultClient == nil {
		log.Entry().Info("Vault not configured, event publishing will be disabled")
		return
	}

	provider := orchestrator.GetOrchestratorConfigProvider(nil)

	publisher := gcp.NewGcpPubsubClient(
		vaultClient.GetOIDCTokenByValidation,
		cfg.GcpProjectNumber,
		cfg.GcpWorkloadIDentityPool,
		cfg.GcpWorkloadIDentityPoolProvider,
		provider.BuildURL(),
		GeneralConfig.HookConfig.OIDCConfig.RoleID,
	)

	if err := runGcpPublishEvent(publisher, &cfg); err != nil {
		// do not fail the step
		log.Entry().WithError(err).Warnf("step execution failed")
	}
}

func runGcpPublishEvent(publisher gcp.PubsubClient, cfg *gcpPublishEventOptions) error {
	eventData, err := eventing.NewEventFromJSON(cfg.EventType, cfg.EventSource, cfg.EventData, cfg.AdditionalEventData)
	if err != nil {
		return fmt.Errorf("failed to create event data: %w", err)
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(eventData), "", "  ")
	
	log.Entry().Debugf("CloudEvent event created: %s", prettyJSON)

	if err = publisher.Publish(cfg.Topic, eventData); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Entry().Infof("Event published successfully! With topic: %s", cfg.Topic)
	return nil
}
