package cmd

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/pact"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func pactPublishContract(config pactPublishContractOptions, telemetryData *telemetry.CustomData) {
	utils := pact.NewUtilsBundle()

	err := runPactPublishContract(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("pactPublishContract step execution failed")
	}
}

func runPactPublishContract(config *pactPublishContractOptions, telemetryData *telemetry.CustomData, utils pact.Utils) error {
	publishConfig := pact.PublishConfig{
		PathToPactsFolder:         config.PactsFolderPath,
		PactBrokerBaseURL:         config.PactBrokerBaseURL,
		PactBrokerUsername:        config.Username,
		PactBrokerPassword:        config.Password,
		PactBrokerToken:           config.Token,
		OrgOrigin:                 config.OrgOrigin,
		OrgAlias:                  config.OrgAlias,
		GitPullID:                 config.GitPullID,
		BuildID:                   config.BuildID,
		GitTargetBranch:           config.TargetBranchName,
		GitRepo:                   config.Repository,
		GitSourceBranch:           config.SourceBranchName,
		GitCommit:                 config.CommitID,
		GitProvider:               config.GitProvider,
		EnforceAsyncAPIValidation: config.EnforceAsyncAPIValidation,
		EnforceOpenAPIValidation:  config.EnforceOpenAPIValidation,
		SystemNamespace:           config.SystemNamespace,
		Utils:                     utils,
		StdOut:                    log.Writer(),
	}

	if err := publishConfig.ExecPactPublish(); err != nil {
		return err
	}
	return nil

	// ToDo: set error categories
}
