package cmd

import (
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

	log.Entry().Debug("publishing event to GCP Pub/Sub...")

	// prepare event data
	payload := events.GenericEventPayload{JSONData: config.EventData}
	payload.Merge(config.AdditionalEventData)

	// create GCP Pub/Sub client
	client := utils.NewPubsubClient(
		config.GcpProjectNumber,
		config.GcpWorkloadIDentityPool,
		config.GcpWorkloadIDentityPoolProvider,
		provider.BuildURL(),
		GeneralConfig.HookConfig.OIDCConfig.RoleID,
	)
	// send event
	if events.Send(
		config.EventSource,
		config.EventType,
		config.Topic,
		&payload,
		client); err != nil {
		log.Entry().WithError(err).Warn("  failed")
	} else {
		log.Entry().Debug("  succeeded")
	}
	return nil
}
