package cmd

import (
	"context"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/stagingcredentials"
)

// acquireNpmExecuteScriptsStagingCredentials fetches just-in-time Nexus credentials
// and populates the repository fields of config when stagingCredentialMode is
// "justInTimeV1". It is a no-op for any other mode or when publish is false.
func acquireNpmExecuteScriptsStagingCredentials(ctx context.Context, config *npmExecuteScriptsOptions) error {
	if !config.Publish || config.StagingCredentialMode != stagingcredentials.ModeJustInTimeV1 {
		return nil
	}

	log.Entry().Info("Acquiring JIT staging credentials for npm publish")

	req := stagingcredentials.Request{
		SystemTrustURL:          GeneralConfig.HookConfig.SystemTrustConfig.ServerURL,
		SystemTrustSessionToken: GeneralConfig.SystemTrustToken,
		StagingServiceURL:       config.StagingServiceURL,
		GroupID:                 config.StagingGroupId,
		RepositoryID:            config.StagingRepositoryId,
		Operation:               stagingcredentials.OperationWrite,
	}

	credential, err := stagingcredentials.NewProvider().Acquire(ctx, req)
	if err != nil {
		return err
	}

	config.RepositoryURL = credential.RepositoryURL
	config.RepositoryUsername = credential.Username
	config.RepositoryPassword = credential.Password
	log.Entry().Infof("JIT staging credentials acquired for repository %s", credential.RepositoryID)
	return nil
}
