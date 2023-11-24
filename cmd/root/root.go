package root

import (
	"os"

	"github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/cmd/cnb"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "piper",
	Short: "Executes CI/CD steps from project 'Piper' ",
	Long: `
This project 'Piper' binary provides a CI/CD step library.
It contains many steps which can be used within CI/CD systems as well as directly on e.g. a developer's machine.
`,
}

// Execute is the starting point of the piper command line tool
func Execute() {
	log.Entry().Infof("Version %s", cmd.GitCommit)

	rootCmd.AddCommand(cmd.ArtifactPrepareVersionCommand())
	rootCmd.AddCommand(cmd.ConfigCommand())
	rootCmd.AddCommand(cmd.DefaultsCommand())
	rootCmd.AddCommand(cmd.ContainerSaveImageCommand())
	rootCmd.AddCommand(cmd.CommandLineCompletionCommand())
	rootCmd.AddCommand(cmd.VersionCommand())
	rootCmd.AddCommand(cmd.DetectExecuteScanCommand())
	rootCmd.AddCommand(cmd.HadolintExecuteCommand())
	rootCmd.AddCommand(cmd.KarmaExecuteTestsCommand())
	rootCmd.AddCommand(cmd.UiVeri5ExecuteTestsCommand())
	rootCmd.AddCommand(cmd.SonarExecuteScanCommand())
	rootCmd.AddCommand(cmd.KubernetesDeployCommand())
	rootCmd.AddCommand(cmd.HelmExecuteCommand())
	rootCmd.AddCommand(cmd.XsDeployCommand())
	rootCmd.AddCommand(cmd.GithubCheckBranchProtectionCommand())
	rootCmd.AddCommand(cmd.GithubCommentIssueCommand())
	rootCmd.AddCommand(cmd.GithubCreateIssueCommand())
	rootCmd.AddCommand(cmd.GithubCreatePullRequestCommand())
	rootCmd.AddCommand(cmd.GithubPublishReleaseCommand())
	rootCmd.AddCommand(cmd.GithubSetCommitStatusCommand())
	rootCmd.AddCommand(cmd.GitopsUpdateDeploymentCommand())
	rootCmd.AddCommand(cmd.CloudFoundryDeleteServiceCommand())
	rootCmd.AddCommand(cmd.AbapEnvironmentPullGitRepoCommand())
	rootCmd.AddCommand(cmd.AbapEnvironmentCloneGitRepoCommand())
	rootCmd.AddCommand(cmd.AbapEnvironmentCheckoutBranchCommand())
	rootCmd.AddCommand(cmd.AbapEnvironmentCreateTagCommand())
	rootCmd.AddCommand(cmd.AbapEnvironmentCreateSystemCommand())
	rootCmd.AddCommand(cmd.CheckmarxExecuteScanCommand())
	rootCmd.AddCommand(cmd.CheckmarxOneExecuteScanCommand())
	rootCmd.AddCommand(cmd.FortifyExecuteScanCommand())
	rootCmd.AddCommand(cmd.CodeqlExecuteScanCommand())
	rootCmd.AddCommand(cmd.CredentialdiggerScanCommand())
	rootCmd.AddCommand(cmd.MtaBuildCommand())
	rootCmd.AddCommand(cmd.ProtecodeExecuteScanCommand())
	rootCmd.AddCommand(cmd.MavenExecuteCommand())
	rootCmd.AddCommand(cmd.CloudFoundryCreateServiceKeyCommand())
	rootCmd.AddCommand(cmd.MavenBuildCommand())
	rootCmd.AddCommand(cmd.MavenExecuteIntegrationCommand())
	rootCmd.AddCommand(cmd.MavenExecuteStaticCodeChecksCommand())
	rootCmd.AddCommand(cmd.NexusUploadCommand())
	rootCmd.AddCommand(cmd.AbapEnvironmentPushATCSystemConfigCommand())
	rootCmd.AddCommand(cmd.AbapEnvironmentRunATCCheckCommand())
	rootCmd.AddCommand(cmd.NpmExecuteScriptsCommand())
	rootCmd.AddCommand(cmd.NpmExecuteLintCommand())
	rootCmd.AddCommand(cmd.GctsCreateRepositoryCommand())
	rootCmd.AddCommand(cmd.GctsExecuteABAPQualityChecksCommand())
	rootCmd.AddCommand(cmd.GctsExecuteABAPUnitTestsCommand())
	rootCmd.AddCommand(cmd.GctsDeployCommand())
	rootCmd.AddCommand(cmd.MalwareExecuteScanCommand())
	rootCmd.AddCommand(cmd.CloudFoundryCreateServiceCommand())
	rootCmd.AddCommand(cmd.CloudFoundryDeployCommand())
	rootCmd.AddCommand(cmd.GctsRollbackCommand())
	rootCmd.AddCommand(cmd.WhitesourceExecuteScanCommand())
	rootCmd.AddCommand(cmd.GctsCloneRepositoryCommand())
	rootCmd.AddCommand(cmd.JsonApplyPatchCommand())
	rootCmd.AddCommand(cmd.KanikoExecuteCommand())
	rootCmd.AddCommand(cnb.CnbBuildCommand())
	rootCmd.AddCommand(cmd.AbapEnvironmentBuildCommand())
	rootCmd.AddCommand(cmd.AbapEnvironmentAssemblePackagesCommand())
	rootCmd.AddCommand(cmd.AbapAddonAssemblyKitCheckCVsCommand())
	rootCmd.AddCommand(cmd.AbapAddonAssemblyKitCheckPVCommand())
	rootCmd.AddCommand(cmd.AbapAddonAssemblyKitCreateTargetVectorCommand())
	rootCmd.AddCommand(cmd.AbapAddonAssemblyKitPublishTargetVectorCommand())
	rootCmd.AddCommand(cmd.AbapAddonAssemblyKitRegisterPackagesCommand())
	rootCmd.AddCommand(cmd.AbapAddonAssemblyKitReleasePackagesCommand())
	rootCmd.AddCommand(cmd.AbapAddonAssemblyKitReserveNextPackagesCommand())
	rootCmd.AddCommand(cmd.CloudFoundryCreateSpaceCommand())
	rootCmd.AddCommand(cmd.CloudFoundryDeleteSpaceCommand())
	rootCmd.AddCommand(cmd.VaultRotateSecretIdCommand())
	rootCmd.AddCommand(cmd.IsChangeInDevelopmentCommand())
	rootCmd.AddCommand(cmd.TransportRequestUploadCTSCommand())
	rootCmd.AddCommand(cmd.TransportRequestUploadRFCCommand())
	rootCmd.AddCommand(cmd.NewmanExecuteCommand())
	rootCmd.AddCommand(cmd.IntegrationArtifactDeployCommand())
	rootCmd.AddCommand(cmd.TransportRequestUploadSOLMANCommand())
	rootCmd.AddCommand(cmd.IntegrationArtifactUpdateConfigurationCommand())
	rootCmd.AddCommand(cmd.IntegrationArtifactGetMplStatusCommand())
	rootCmd.AddCommand(cmd.IntegrationArtifactGetServiceEndpointCommand())
	rootCmd.AddCommand(cmd.IntegrationArtifactDownloadCommand())
	rootCmd.AddCommand(cmd.AbapEnvironmentAssembleConfirmCommand())
	rootCmd.AddCommand(cmd.IntegrationArtifactUploadCommand())
	rootCmd.AddCommand(cmd.IntegrationArtifactTriggerIntegrationTestCommand())
	rootCmd.AddCommand(cmd.IntegrationArtifactUnDeployCommand())
	rootCmd.AddCommand(cmd.IntegrationArtifactResourceCommand())
	rootCmd.AddCommand(cmd.TerraformExecuteCommand())
	rootCmd.AddCommand(cmd.ContainerExecuteStructureTestsCommand())
	rootCmd.AddCommand(cmd.GaugeExecuteTestsCommand())
	rootCmd.AddCommand(cmd.BatsExecuteTestsCommand())
	rootCmd.AddCommand(cmd.PipelineCreateScanSummaryCommand())
	rootCmd.AddCommand(cmd.TransportRequestDocIDFromGitCommand())
	rootCmd.AddCommand(cmd.TransportRequestReqIDFromGitCommand())
	rootCmd.AddCommand(cmd.WritePipelineEnv())
	rootCmd.AddCommand(cmd.ReadPipelineEnv())
	rootCmd.AddCommand(cmd.InfluxWriteDataCommand())
	rootCmd.AddCommand(cmd.AbapEnvironmentRunAUnitTestCommand())
	rootCmd.AddCommand(cmd.CheckStepActiveCommand())
	rootCmd.AddCommand(cmd.GolangBuildCommand())
	rootCmd.AddCommand(cmd.ShellExecuteCommand())
	rootCmd.AddCommand(cmd.ApiProxyDownloadCommand())
	rootCmd.AddCommand(cmd.ApiKeyValueMapDownloadCommand())
	rootCmd.AddCommand(cmd.ApiProviderDownloadCommand())
	rootCmd.AddCommand(cmd.ApiProxyUploadCommand())
	rootCmd.AddCommand(cmd.GradleExecuteBuildCommand())
	rootCmd.AddCommand(cmd.ApiKeyValueMapUploadCommand())
	rootCmd.AddCommand(cmd.PythonBuildCommand())
	rootCmd.AddCommand(cmd.AzureBlobUploadCommand())
	rootCmd.AddCommand(cmd.AwsS3UploadCommand())
	rootCmd.AddCommand(cmd.ApiProxyListCommand())
	rootCmd.AddCommand(cmd.AnsSendEventCommand())
	rootCmd.AddCommand(cmd.ApiProviderListCommand())
	rootCmd.AddCommand(cmd.TmsUploadCommand())
	rootCmd.AddCommand(cmd.TmsExportCommand())
	rootCmd.AddCommand(cmd.IntegrationArtifactTransportCommand())
	rootCmd.AddCommand(cmd.AscAppUploadCommand())

	addRootFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		log.Entry().WithError(err).Fatal("configuration error")
	}
}

func addRootFlags(rootCmd *cobra.Command) {
	var provider orchestrator.OrchestratorSpecificConfigProviding
	var err error

	provider, err = orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil {
		log.Entry().Error(err)
		provider = &orchestrator.UnknownOrchestratorConfigProvider{}
	}

	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.CorrelationID, "correlationID", provider.GetBuildURL(), "ID for unique identification of a pipeline run")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.CustomConfig, "customConfig", ".pipeline/config.yml", "Path to the pipeline configuration file")
	rootCmd.PersistentFlags().StringSliceVar(&cmd.GeneralConfig.GitHubTokens, "gitHubTokens", cmd.AccessTokensFromEnvJSON(os.Getenv("PIPER_gitHubTokens")), "List of entries in form of <hostname>:<token> to allow GitHub token authentication for downloading config / defaults")
	rootCmd.PersistentFlags().StringSliceVar(&cmd.GeneralConfig.DefaultConfig, "defaultConfig", []string{".pipeline/defaults.yaml"}, "Default configurations, passed as path to yaml file")
	rootCmd.PersistentFlags().BoolVar(&cmd.GeneralConfig.IgnoreCustomDefaults, "ignoreCustomDefaults", false, "Disables evaluation of the parameter 'customDefaults' in the pipeline configuration file")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.ParametersJSON, "parametersJSON", os.Getenv("PIPER_parametersJSON"), "Parameters to be considered in JSON format")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.EnvRootPath, "envRootPath", ".pipeline", "Root path to Piper pipeline shared environments")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.StageName, "stageName", "", "Name of the stage for which configuration should be included")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.StepConfigJSON, "stepConfigJSON", os.Getenv("PIPER_stepConfigJSON"), "Step configuration in JSON format")
	rootCmd.PersistentFlags().BoolVar(&cmd.GeneralConfig.NoTelemetry, "noTelemetry", false, "Disables telemetry reporting")
	rootCmd.PersistentFlags().BoolVarP(&cmd.GeneralConfig.Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.LogFormat, "logFormat", "default", "Log format to use. Options: default, timestamp, plain, full.")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.VaultServerURL, "vaultServerUrl", "", "The Vault server which should be used to fetch credentials")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.VaultNamespace, "vaultNamespace", "", "The Vault namespace which should be used to fetch credentials")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.VaultPath, "vaultPath", "", "The path which should be used to fetch credentials")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.GCPJsonKeyFilePath, "gcpJsonKeyFilePath", "", "File path to Google Cloud Platform JSON key file")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.GCSFolderPath, "gcsFolderPath", "", "GCS folder path. One of the components of GCS target folder")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.GCSBucketId, "gcsBucketId", "", "Bucket name for Google Cloud Storage")
	rootCmd.PersistentFlags().StringVar(&cmd.GeneralConfig.GCSSubFolder, "gcsSubFolder", "", "Used to logically separate results of the same step result type")

}
