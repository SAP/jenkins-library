package cmd

import (
	"fmt"

	piperConfig "github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/events"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/pkg/errors"
)

func gcpPublishEvent(config gcpPublishEventOptions, telemetryData *telemetry.CustomData) {
	err := runGcpPublishEvent(&config, telemetryData)
	if err != nil {
		// do not fail the step
		log.Entry().WithError(err).Warnf("step execution failed")
	}
}

func runGcpPublishEvent(config *gcpPublishEventOptions, _ *telemetry.CustomData) error {
	provider, _ := orchestrator.GetOrchestratorConfigProvider(nil)

	var data []byte
	var err error

	switch config.Type {
	case string(events.PipelineRunStartedEventType):
		data, err = events.ToByteArray(event.EventContextV1{}, events.PipelineRunStartedEventData{
			URL:           provider.BuildURL(),
			CommitId:      provider.CommitSHA(),
			RepositoryURL: provider.RepoURL(),
		})
	case string(events.PipelineRunFinishedEventType):
		data, err = events.ToByteArray(event.EventContextV1{}, events.PipelineRunFinishedEventData{
			URL:           provider.BuildURL(),
			CommitId:      provider.CommitSHA(),
			RepositoryURL: provider.RepoURL(),
			Outcome:       provider.BuildStatus(),
		})
	default:
		return fmt.Errorf("event type %s not supported", config.Type)
	}
	if err != nil {
		return errors.Wrap(err, "failed to create event data")
	}

	oidcToken, err := getOidcToken(config)
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

func getOidcToken(config *gcpPublishEventOptions) (string, error) {
	vaultCreds := piperConfig.VaultCredentials{
		AppRoleID:       GeneralConfig.VaultRoleID,
		AppRoleSecretID: GeneralConfig.VaultRoleSecretID,
		VaultToken:      GeneralConfig.VaultToken,
	}
	// GeneralConfig VaultServerURL and VaultNamespace are empty swicthing to stepConfig
	var vaultConfig = map[string]interface{}{
		"vaultServerUrl": config.VaultServerURL,
		"vaultNamespace": config.VaultNamespace,
	}

	stepConfig := piperConfig.StepConfig{
		Config: vaultConfig,
	}
	// Generating vault client
	vaultClient, err := piperConfig.GetVaultClientFromConfig(stepConfig, vaultCreds)
	if err != nil {
		return "", errors.Wrap(err, "getting vault client failed")
	}
	// Getting oidc token and setting it in environment variable
	token, err := vaultClient.GetOidcTokenByValidation(GeneralConfig.HookConfig.OidcConfig.RoleID)
	if err != nil {
		return "", errors.Wrap(err, "getting oidc token failed")
	}

	return token, nil
}
