package cmd

import (
	"context"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/stagingcredentials"
)

// acquirePythonBuildStagingCredentials fetches just-in-time Nexus credentials
// and populates the repository fields of config when stagingCredentialMode is
// "justInTimeV1". It is a no-op for any other mode or when publish is false.
func acquirePythonBuildStagingCredentials(ctx context.Context, config *pythonBuildOptions) error {
	if !config.Publish || config.StagingCredentialMode != stagingcredentials.ModeJustInTimeV1 {
		return nil
	}

	log.Entry().Info("Acquiring JIT staging credentials for Python build")

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

	config.TargetRepositoryURL = credential.RepositoryURL
	config.TargetRepositoryUser = credential.Username
	config.TargetRepositoryPassword = credential.Password
	log.Entry().Infof("JIT staging credentials acquired for repository %s", credential.RepositoryID)
	return nil
}
