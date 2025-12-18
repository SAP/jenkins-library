package cmd

import (
	piperConfig "github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/events"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "failed to create event data")
	}

	err = utils.NewPubsubClient(
		config.GcpProjectNumber,
		config.GcpWorkloadIDentityPool,
		config.GcpWorkloadIDentityPoolProvider,
		provider.BuildURL(),
		GeneralConfig.HookConfig.OIDCConfig.RoleID,
	).Publish(config.Topic, data)
	if err != nil {
		return errors.Wrap(err, "failed to publish event")
	}

	log.Entry().Info("event publish succeeded")
	log.Entry().Infof("  with topic %s", config.Topic)
	log.Entry().Debugf("  with data %s", string(data))
	return nil
}

func createNewEvent(config *gcpPublishEventOptions) ([]byte, error) {
	event, err := events.NewEvent(config.EventType, config.EventSource, "").CreateWithJSONData(config.EventData)
	if err != nil {
		return []byte{}, errors.Wrap(err, "failed to create new event")
	}

	err = event.AddToCloudEventData(config.AdditionalEventData)
	if err != nil {
		log.Entry().Debugf("couldn't add additionalData to cloud event data: %s", err)
	}

	eventBytes, err := event.ToBytes()
	if err != nil {
		return []byte{}, errors.Wrap(err, "casting event to bytes failed")
	}
	log.Entry().Debugf("CloudEvent created: %s", string(eventBytes))
	return eventBytes, nil
}
