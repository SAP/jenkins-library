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
}

type gcpPublishEventUtilsBundle struct {
	config *gcpPublishEventOptions
	*vault.Client
}

func (g gcpPublishEventUtilsBundle) GetConfig() *gcpPublishEventOptions {
	return g.config
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
	if err != nil {
		log.Entry().WithError(err).Fatal("could not create Vault client")
	}
	defer client.MustRevokeToken()

	vaultClient, ok := client.(vault.Client)
	if !ok {
		log.Entry().WithError(err).Fatal("could not create Vault client")
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

	provider, _ := orchestrator.GetOrchestratorConfigProvider(nil)

	var data []byte
	var err error

	data, err = events.NewEvent(config.EventType, config.EventSource).CreateWithProviderData(provider).ToBytes()

	if err != nil {
		return errors.Wrap(err, "failed to create event data")
	}

	oidcToken, err := getOIDCToken(utils)
	if err != nil {
		return errors.Wrap(err, "failed to get OIDC token")
	}

	// get federated token
	token, err := gcp.GetFederatedToken(config.GcpProjectNumber, config.GcpWorkloadIDentityPool, config.GcpWorkloadIDentityPoolProvider, oidcToken)
	if err != nil {
		return errors.Wrap(err, "failed to get federated token")
	}

	// publish event
	err = gcp.Publish(config.GcpProjectNumber, config.Topic, token, data)
	if err != nil {
		return errors.Wrap(err, "failed to publish event")
	}

	log.Entry().Info("event published successfully!")

	return nil
}

func getOIDCToken(utils gcpPublishEventUtils) (string, error) {
	token, err := utils.GetOIDCTokenByValidation(GeneralConfig.HookConfig.OIDCConfig.RoleID)
	if err != nil {
		return "", errors.Wrap(err, "getting OIDC token failed")
	}

	return token, nil
}
