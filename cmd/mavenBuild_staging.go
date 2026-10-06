package cmd

import (
	"context"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/stagingcredentials"
)

// acquireMavenBuildStagingCredentials fetches just-in-time Nexus credentials
// and populates the repository fields of config when stagingCredentialMode is
// "justInTimeV1". It is a no-op for any other mode or when publish is false.
func acquireMavenBuildStagingCredentials(ctx context.Context, config *mavenBuildOptions) error {
	if !config.Publish || config.StagingCredentialMode != stagingcredentials.ModeJustInTimeV1 {
		return nil
	}

	log.Entry().Info("Acquiring JIT staging credentials for Maven build")

	req := stagingcredentials.Request{
		SystemTrustURL:          GeneralConfig.HookConfig.SystemTrustConfig.ServerURL,
		SystemTrustSessionToken: GeneralConfig.SystemTrustToken,
		StagingServiceURL:       config.StagingServiceURL,
		GroupID:                 config.StagingGroupID,
		RepositoryID:            config.StagingRepositoryID,
		Operation:               stagingcredentials.OperationWrite,
	}

	credential, err := stagingcredentials.NewProvider().Acquire(ctx, req)
	if err != nil {
		return err
	}

	config.AltDeploymentRepositoryID = credential.RepositoryID
	config.AltDeploymentRepositoryURL = credential.RepositoryURL
	config.AltDeploymentRepositoryUser = credential.Username
	config.AltDeploymentRepositoryPassword = credential.Password
	log.Entry().Infof("JIT staging credentials acquired for repository %s", credential.RepositoryID)
	return nil
}
