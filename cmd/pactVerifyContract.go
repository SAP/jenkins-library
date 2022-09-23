package cmd

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/pact"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func pactVerifyContract(config pactVerifyContractOptions, telemetryData *telemetry.CustomData) {
	utils := pact.NewUtilsBundle()

	err := runPactVerifyContract(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("pactPublishContract step execution failed")
	}
}

func runPactVerifyContract(config *pactVerifyContractOptions, telemetryData *telemetry.CustomData, utils pact.Utils) error {
	verifyConfig := pact.VerifyConfig{
		PathToAsyncFile:           config.AsynchAPIFilePath,
		PathToSwaggerFile:         config.SwaggerFilePath,
		PactBrokerBaseURL:         config.PactBrokerBaseURL,
		PactBrokerUsername:        config.Username,
		PactBrokerPassword:        config.Password,
		OrgOrigin:                 config.OrgOrigin,
		OrgAlias:                  config.OrgAlias,
		GitProvider:               config.GitProvider,
		GitRepo:                   config.Repository,
		GitPullID:                 config.GitPullID,
		BuildID:                   config.BuildID,
		GitTargetBranch:           config.TargetBranchName,
		GitSourceBranch:           config.SourceBranchName,
		GitCommit:                 config.CommitID,
		EnforceAsyncAPIValidation: config.EnforceAsyncAPIValidation,
		EnforceOpenAPIValidation:  config.EnforceOpenAPIValidation,
		SystemNamespace:           config.SystemNamespace,
		Provider:                  config.Provider,
		Utils:                     utils,
	}

	if err := verifyConfig.ExecPactVerify(); err != nil {
		return err
	}
	return nil
}
