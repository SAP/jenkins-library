package cmd

import (
	"fmt"

	piperConfig "github.com/SAP/jenkins-library/pkg/config"
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
	piperConfig.VaultClient
}

func (g *gcpPublishEventUtilsBundle) GetConfig() *gcpPublishEventOptions {
	return g.config
}

func (g *gcpPublishEventUtilsBundle) NewPubsubClient(projectNumber, pool, provider, key, oidcRoleId string) gcp.PubsubClient {
	return gcp.NewGcpPubsubClient(g.VaultClient, projectNumber, pool, provider, key, oidcRoleId)
}

func gcpPublishEvent(config gcpPublishEventOptions, telemetryData *telemetry.CustomData) {
	utils := &gcpPublishEventUtilsBundle{
		config:      &config,
		VaultClient: piperConfig.GlobalVaultClient(),
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

	config := utils.GetConfig()
	data, err := createNewEvent(config)
	if err != nil {
		return fmt.Errorf("failed to create event data: %w", err)
	}

	err = utils.NewPubsubClient(
		config.GcpProjectNumber,
		config.GcpWorkloadIDentityPool,
		config.GcpWorkloadIDentityPoolProvider,
		provider.BuildURL(),
		GeneralConfig.HookConfig.OIDCConfig.RoleID,
	).Publish(config.Topic, data)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Entry().Infof("Event published successfully! With topic: %s", config.Topic)
	return nil
}

func createNewEvent(config *gcpPublishEventOptions) ([]byte, error) {
	event, err := events.NewEvent(config.EventType, config.EventSource, "").CreateWithJSONData(config.EventData)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create new event: %w", err)
	}

	err = event.AddToCloudEventData(config.AdditionalEventData)
	if err != nil {
		log.Entry().Debugf("couldn't add additionalData to cloud event data: %s", err)
	}

	eventBytes, err := event.ToBytes()
	if err != nil {
		return []byte{}, fmt.Errorf("casting event to bytes failed: %w", err)
	}
	log.Entry().Debugf("CloudEvent created: %s", string(eventBytes))
	return eventBytes, nil
}
