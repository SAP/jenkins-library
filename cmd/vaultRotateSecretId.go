package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"

	"github.com/SAP/jenkins-library/pkg/ado"
	"github.com/SAP/jenkins-library/pkg/jenkins"
	"github.com/SAP/jenkins-library/pkg/vault"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type vaultRotateSecretIDUtils interface {
	GetAppRoleSecretIDTtl(secretID, roleName string) (time.Duration, error)
	GetAppRoleName() (string, error)
	GenerateNewAppRoleSecret(secretID string, roleName string) (string, error)
	UpdateSecretInStore(config *vaultRotateSecretIdOptions, secretID string) error
	GetConfig() *vaultRotateSecretIdOptions
}

type vaultRotateSecretIDUtilsBundle struct {
	*vault.Client
	config     *vaultRotateSecretIdOptions
	updateFunc func(config *vaultRotateSecretIdOptions, secretID string) error
}

func (v vaultRotateSecretIDUtilsBundle) GetConfig() *vaultRotateSecretIdOptions {
	return v.config
}

func (v vaultRotateSecretIDUtilsBundle) UpdateSecretInStore(config *vaultRotateSecretIdOptions, secretID string) error {
	return v.updateFunc(config, secretID)
}

func vaultRotateSecretId(config vaultRotateSecretIdOptions, telemetryData *telemetry.CustomData) {

	vaultConfig := &vault.Config{
		Config: &api.Config{
			Address: config.VaultServerURL,
		},
		Namespace: config.VaultNamespace,
	}
	client, err := vault.NewClientWithAppRole(vaultConfig, GeneralConfig.VaultRoleID, GeneralConfig.VaultRoleSecretID)
	if err != nil {
		log.Entry().WithError(err).Fatal("could not create vault client")
	}
	defer client.MustRevokeToken()

	utils := vaultRotateSecretIDUtilsBundle{
		Client:     &client,
		config:     &config,
		updateFunc: writeVaultSecretIDToStore,
	}

	err = runVaultRotateSecretID(utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runVaultRotateSecretID(utils vaultRotateSecretIDUtils) error {
	config := utils.GetConfig()

	roleName, err := utils.GetAppRoleName()
	if err != nil {
		log.Entry().WithError(err).Warn("Could not fetch approle role name from vault. Secret ID rotation failed!")
		return nil
	}

	ttl, err := utils.GetAppRoleSecretIDTtl(GeneralConfig.VaultRoleSecretID, roleName)

	if err != nil {
		log.Entry().WithError(err).Warn("Could not fetch secret ID TTL. Secret ID rotation failed!")
		return nil
	}

	log.Entry().Debugf("Your secret ID is about to expire in %.0f", ttl.Round(time.Hour*24).Hours()/24)

	if ttl > time.Duration(config.DaysBeforeExpiry)*24*time.Hour {
		return nil
	}

	newSecretID, err := utils.GenerateNewAppRoleSecret(GeneralConfig.VaultRoleSecretID, roleName)

	if err != nil || newSecretID == "" {
		log.Entry().WithError(err).Warn("Generating a new secret ID failed. Secret ID rotation faield!")
		return nil
	}

	if err = utils.UpdateSecretInStore(config, newSecretID); err != nil {
		log.Entry().WithError(err).Warnf("Could not write secret back to secret store %s", config.SecretStore)
		return err
	}
	log.Entry().Infof("Secret has been successfully updated in secret store %s", config.SecretStore)
	return nil

}

func writeVaultSecretIDToStore(config *vaultRotateSecretIdOptions, secretID string) error {
	switch config.SecretStore {
	case "jenkins":
		ctx := context.Background()
		instance, err := jenkins.Instance(ctx, &http.Client{}, config.JenkinsURL, config.JenkinsUsername, config.JenkinsToken)
		if err != nil {
			log.Entry().Warn("Could not write secret ID back to jenkins")
			return err
		}
		credManager := jenkins.NewCredentialsManager(instance)
		credential := jenkins.StringCredentials{ID: config.VaultAppRoleSecretTokenCredentialsID, Secret: secretID}
		return jenkins.UpdateCredential(ctx, credManager, config.JenkinsCredentialDomain, credential)
	case "ado":
		adoBuildClient, err := ado.NewBuildClient(config.AdoOrganization, config.AdoPersonalAccessToken, config.AdoProject, config.AdoPipelineID)
		if err != nil {
			log.Entry().Warn("Could not write secret ID back to jenkins")
			return err
		}
		variables := []ado.Variable{
			{
				Name:     config.VaultAppRoleSecretTokenCredentialsID,
				Value:    secretID,
				IsSecret: true,
			},
		}
		if err := adoBuildClient.UpdateVariables(variables); err != nil {
			log.Entry().Warn("Could not write secret ID back to jenkins")
			return err
		}
	default:
		return fmt.Errorf("error: invalid secret store: %s", config.SecretStore)
	}
	return nil
}
