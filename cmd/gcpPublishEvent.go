package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/events"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type gcpPublishEventUtils interface {
	GetConfig() *gcpPublishEventOptions
	NewPubsubClient(projectNumber, pool, provider, key, oidcRoleId string) gcp.PubsubClient
}

type gcpPublishEventUtilsBundle struct {
	config *gcpPublishEventOptions
	config.VaultClient
}

func (g *gcpPublishEventUtilsBundle) GetConfig() *gcpPublishEventOptions {
	return g.config
}

func (g *gcpPublishEventUtilsBundle) NewPubsubClient(projectNumber, pool, provider, key, oidcRoleId string) gcp.PubsubClient {
	return gcp.NewGcpPubsubClient(g.VaultClient.GetOIDCTokenByValidation, projectNumber, pool, provider, key, oidcRoleId)
}

func gcpPublishEvent(cfg gcpPublishEventOptions, telemetryData *telemetry.CustomData) {
	utils := &gcpPublishEventUtilsBundle{
		config:      &cfg,
		VaultClient: config.GlobalVaultClient(),
	}

	if err := runGcpPublishEvent(utils); err != nil {
		// do not fail the step
		log.Entry().WithError(err).Warnf("step execution failed")
	}
}

func runGcpPublishEvent(utils gcpPublishEventUtils) error {
	provider, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		log.Entry().WithError(err).Warning("Cannot infer config from CI environment")
	}

	cfg := utils.GetConfig()
	data, err := createNewEvent(cfg)
	if err != nil {
		return fmt.Errorf("failed to create event data: %w", err)
	}

	err = utils.NewPubsubClient(
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

func createNewEvent(config *gcpPublishEventOptions) ([]byte, error) {
	event, err := events.NewEvent(config.EventType, config.EventSource, "").CreateWithJSONData(config.EventData)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create new event: %w", err)
	}

	if err = event.AddToCloudEventData(config.AdditionalEventData); err != nil {
		log.Entry().Debugf("couldn't add additionalData to cloud event data: %s", err)
	}

	eventBytes, err := event.ToBytes()
	if err != nil {
		return []byte{}, fmt.Errorf("casting event to bytes failed: %w", err)
	}
	log.Entry().Debugf("CloudEvent created: %s", string(eventBytes))
	return eventBytes, nil
}
