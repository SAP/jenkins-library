package cmd

import (
	"github.com/SAP/jenkins-library/pkg/events"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/pkg/errors"
)

type gcpPublishEventUtils interface {
	GetConfig() *gcpPublishEventOptions
	GetOIDCTokenByValidation(roleID string) (string, error)
	GetFederatedToken(projectNumber, pool, provider, token string) (string, error)
	Publish(projectNumber string, topic string, token string, key string, data []byte) error
}

type gcpPublishEventUtilsBundle struct {
	config *gcpPublishEventOptions
}

func (g gcpPublishEventUtilsBundle) GetConfig() *gcpPublishEventOptions {
	return g.config
}

func (g gcpPublishEventUtilsBundle) GetFederatedToken(projectNumber, pool, provider, token string) (string, error) {
	return gcp.GetFederatedToken(projectNumber, pool, provider, token)
}

func (g gcpPublishEventUtilsBundle) Publish(projectNumber string, topic string, token string, key string, data []byte) error {
	return gcp.Publish(projectNumber, topic, token, key, data)
}

// to be implemented through another PR!
func (g gcpPublishEventUtilsBundle) GetOIDCTokenByValidation(roleID string) (string, error) {
	return "testToken", nil
}

func gcpPublishEvent(config gcpPublishEventOptions, telemetryData *telemetry.CustomData) {
	utils := gcpPublishEventUtilsBundle{
		config: &config,
	}

	err := runGcpPublishEvent(utils)
	if err != nil {
		// do not fail the step
		log.Entry().WithError(err).Warnf("step execution failed")
	}
}

func runGcpPublishEvent(utils gcpPublishEventUtils) error {
	config := utils.GetConfig()

	var data []byte
	var err error

	provider, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		log.Entry().WithError(err).Warning("Cannot infer config from CI environment")
	}

	data, err = events.NewEvent(config.EventType, config.EventSource).CreateWithJSONData(config.EventData).ToBytes()
	if err != nil {
		return errors.Wrap(err, "failed to create event data")
	}

	// this is currently returning a mock token. function will be implemented through another PR!
	// roleID will come from GeneralConfig.HookConfig.OIDCConfig.RoleID
	roleID := "test"
	oidcToken, err := utils.GetOIDCTokenByValidation(roleID)
	if err != nil {
		return errors.Wrap(err, "failed to get OIDC token")
	}

	token, err := utils.GetFederatedToken(config.GcpProjectNumber, config.GcpWorkloadIDentityPool, config.GcpWorkloadIDentityPoolProvider, oidcToken)
	if err != nil {
		return errors.Wrap(err, "failed to get federated token")
	}

	err = utils.Publish(config.GcpProjectNumber, config.Topic, token, provider.BuildURL(), data)
	if err != nil {
		return errors.Wrap(err, "failed to publish event")
	}

	log.Entry().Info("event published successfully!")

	return nil
}
