package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"

	"github.com/SAP/jenkins-library/pkg/ado"
	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/jenkins"
	"github.com/SAP/jenkins-library/pkg/vault"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const automaticdTTLThreshold = 18 * 24 * time.Hour // Threshold for automaticd service to rotate secrets

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

	vaultConfig := &vault.ClientConfig{
		Config: &api.Config{
			Address: config.VaultServerURL,
		},
		Namespace: config.VaultNamespace,
		RoleID:    GeneralConfig.VaultRoleID,
		SecretID:  GeneralConfig.VaultRoleSecretID,
	}
	client, err := vault.NewClient(vaultConfig)
	if err != nil {
		log.Entry().WithError(err).Fatal("could not create Vault client")
	}
	defer client.MustRevokeToken()

	utils := vaultRotateSecretIDUtilsBundle{
		Client:     client,
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
		log.Entry().WithError(err).Warn("Could not fetch Vault AppRole role name from Vault. Secret ID rotation failed!")
		return nil
	}

	ttl, err := utils.GetAppRoleSecretIDTtl(GeneralConfig.VaultRoleSecretID, roleName)
	if err != nil {
		log.Entry().WithError(err).Warn("Could not fetch secret ID TTL. Secret ID rotation failed!")
		return nil
	}

	if ttl == 0 {
		log.Entry().Warn("Secret ID expired")
	} else {
		log.Entry().Infof("Your secret ID is about to expire in %.0f days", ttl.Round(time.Hour*24).Hours()/24)
	}

	if ttl > time.Duration(config.DaysBeforeExpiry)*24*time.Hour {
		log.Entry().Info("Secret ID TTL valid.")
		return nil
	}

	// Check if the secret store is ADO and apply the TTL condition
	if config.SecretStore == "ado" {
		warnMessage := "ADO Personal Access Token is required but not provided. Secret ID rotation cannot proceed for Azure DevOps."
		// Check if the secret ID TTL is less than 18 days and greater than or equal to the configured days before expiry
		if ttl < automaticdTTLThreshold && ttl >= time.Duration(config.DaysBeforeExpiry)*24*time.Hour {
			log.Entry().Warn("automaticd service did not update Vault secrets. Attempting to update the secret with PAT.")
			// Check if ADO Personal Access Token is missing
			if config.AdoPersonalAccessToken == "" {
				log.Entry().Warn(warnMessage)
				return fmt.Errorf("ADO Personal Access Token is missing")
			}
		}
	}

	log.Entry().Info("Rotating secret ID...")

	newSecretID, err := utils.GenerateNewAppRoleSecret(GeneralConfig.VaultRoleSecretID, roleName)
	if err != nil || newSecretID == "" {
		log.Entry().WithError(err).Warn("Generating a new secret ID failed. Secret ID rotation failed!")
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
			log.Entry().Warn("Could not write secret ID back to Jenkins")
			return err
		}
		credManager := jenkins.NewCredentialsManager(instance)
		credential := jenkins.StringCredentials{ID: config.VaultAppRoleSecretTokenCredentialsID, Secret: secretID}
		return jenkins.UpdateCredential(ctx, credManager, config.JenkinsCredentialDomain, credential)
	case "ado":
		adoBuildClient, err := ado.NewBuildClient(config.AdoOrganization, config.AdoPersonalAccessToken, config.AdoProject, config.AdoPipelineID)
		if err != nil {
			log.Entry().Warn("Could not write secret ID back to Azure DevOps")
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
			log.Entry().Warn("Could not write secret ID back to Azure DevOps")
			return err
		}
	case "github":
		// Additional info:
		// https://github.com/google/go-github/blob/master/example/newreposecretwithxcrypto/main.go

		ctx, client, err := piperGithub.NewClientBuilder(config.GithubToken, config.GithubAPIURL).Build()
		if err != nil {
			log.Entry().Warnf("Could not write secret ID back to GitHub Actions: GitHub client not created: %v", err)
			return err
		}

		publicKey, _, err := client.Actions.GetRepoPublicKey(ctx, config.Owner, config.Repository)
		if err != nil {
			log.Entry().Warnf("Could not write secret ID back to GitHub Actions: repository's public key not retrieved: %v", err)
			return err
		}

		encryptedSecret, err := piperGithub.CreateEncryptedSecret(config.VaultAppRoleSecretTokenCredentialsID, secretID, publicKey)
		if err != nil {
			log.Entry().Warnf("Could not write secret ID back to GitHub Actions: secret encryption failed: %v", err)
			return err
		}

		_, err = client.Actions.CreateOrUpdateRepoSecret(ctx, config.Owner, config.Repository, encryptedSecret)
		if err != nil {
			log.Entry().Warnf("Could not write secret ID back to GitHub Actions: submission to GitHub failed: %v", err)
			return err
		}
	default:
		return fmt.Errorf("error: invalid secret store: %s", config.SecretStore)
	}
	return nil
}
