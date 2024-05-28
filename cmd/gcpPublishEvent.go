package cmd

import (
	piperConfig "github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/events"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/vault"

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
	*vault.Client
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

func gcpPublishEvent(config gcpPublishEventOptions, telemetryData *telemetry.CustomData) {
	vaultCreds := piperConfig.VaultCredentials{
		AppRoleID:       GeneralConfig.VaultRoleID,
		AppRoleSecretID: GeneralConfig.VaultRoleSecretID,
		VaultToken:      GeneralConfig.VaultToken,
	}
	vaultConfig := map[string]interface{}{
		"vaultNamespace": config.VaultNamespace,
		"vaultServerUrl": config.VaultServerURL,
	}

	client, err := piperConfig.GetVaultClientFromConfig(vaultConfig, vaultCreds)
	if err != nil || client == nil {
		log.Entry().WithError(err).Warnf("could not create Vault client: incomplete Vault configuration")
		return
	}
	defer client.MustRevokeToken()

	vaultClient, ok := client.(vault.Client)
	if !ok {
		log.Entry().WithError(err).Warnf("could not create Vault client")
		return
	}

	utils := gcpPublishEventUtilsBundle{
		config: &config,
		Client: &vaultClient,
	}

	err = runGcpPublishEvent(utils)
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

	oidcToken, err := utils.GetOIDCTokenByValidation(GeneralConfig.HookConfig.OIDCConfig.RoleID)
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
