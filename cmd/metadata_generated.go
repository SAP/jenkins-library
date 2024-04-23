// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import "github.com/SAP/jenkins-library/pkg/config"
import "github.com/SAP/jenkins-library/cmd/abapAddonAssemblyKit"

// GetStepMetadata return a map with all the step metadata mapped to their stepName
func GetAllStepMetadata() map[string]config.StepData {
	return map[string]config.StepData{
		"abapAddonAssemblyKit.AbapAddonAssemblyKitCheckCVs": abapAddonAssemblyKit.AbapAddonAssemblyKitCheckCVsMetadata(),
		"AbapAddonAssemblyKitCheckPV":                       AbapAddonAssemblyKitCheckPVMetadata(),
		"AbapAddonAssemblyKitCreateTargetVector":            AbapAddonAssemblyKitCreateTargetVectorMetadata(),
		"AbapAddonAssemblyKitPublishTargetVector":           AbapAddonAssemblyKitPublishTargetVectorMetadata(),
		"AbapAddonAssemblyKitRegisterPackages":              AbapAddonAssemblyKitRegisterPackagesMetadata(),
		"AbapAddonAssemblyKitReleasePackages":               AbapAddonAssemblyKitReleasePackagesMetadata(),
		"AbapAddonAssemblyKitReserveNextPackages":           AbapAddonAssemblyKitReserveNextPackagesMetadata(),
		"AbapEnvironmentAssembleConfirm":                    AbapEnvironmentAssembleConfirmMetadata(),
		"AbapEnvironmentAssemblePackages":                   AbapEnvironmentAssemblePackagesMetadata(),
		"AbapEnvironmentBuild":                              AbapEnvironmentBuildMetadata(),
		"AbapEnvironmentCheckoutBranch":                     AbapEnvironmentCheckoutBranchMetadata(),
		"AbapEnvironmentCloneGitRepo":                       AbapEnvironmentCloneGitRepoMetadata(),
		"AbapEnvironmentCreateSystem":                       AbapEnvironmentCreateSystemMetadata(),
		"AbapEnvironmentCreateTag":                          AbapEnvironmentCreateTagMetadata(),
		"AbapEnvironmentPullGitRepo":                        AbapEnvironmentPullGitRepoMetadata(),
		"AbapEnvironmentPushATCSystemConfig":                AbapEnvironmentPushATCSystemConfigMetadata(),
		"AbapEnvironmentRunATCCheck":                        AbapEnvironmentRunATCCheckMetadata(),
		"AbapEnvironmentRunAUnitTest":                       AbapEnvironmentRunAUnitTestMetadata(),
		"AbapLandscapePortalUpdateAddOnProduct":             AbapLandscapePortalUpdateAddOnProductMetadata(),
		"AnsSendEvent":                                      AnsSendEventMetadata(),
		"ApiKeyValueMapDownload":                            ApiKeyValueMapDownloadMetadata(),
		"ApiKeyValueMapUpload":                              ApiKeyValueMapUploadMetadata(),
		"ApiProviderDownload":                               ApiProviderDownloadMetadata(),
		"ApiProviderList":                                   ApiProviderListMetadata(),
		"ApiProviderUpload":                                 ApiProviderUploadMetadata(),
		"ApiProxyDownload":                                  ApiProxyDownloadMetadata(),
		"ApiProxyList":                                      ApiProxyListMetadata(),
		"ApiProxyUpload":                                    ApiProxyUploadMetadata(),
		"ArtifactPrepareVersion":                            ArtifactPrepareVersionMetadata(),
		"AscAppUpload":                                      AscAppUploadMetadata(),
		"AwsS3Upload":                                       AwsS3UploadMetadata(),
		"AzureBlobUpload":                                   AzureBlobUploadMetadata(),
		"BatsExecuteTests":                                  BatsExecuteTestsMetadata(),
		"CheckmarxExecuteScan":                              CheckmarxExecuteScanMetadata(),
		"CheckmarxOneExecuteScan":                           CheckmarxOneExecuteScanMetadata(),
		"CloudFoundryCreateService":                         CloudFoundryCreateServiceMetadata(),
		"CloudFoundryCreateServiceKey":                      CloudFoundryCreateServiceKeyMetadata(),
		"CloudFoundryCreateSpace":                           CloudFoundryCreateSpaceMetadata(),
		"CloudFoundryDeleteService":                         CloudFoundryDeleteServiceMetadata(),
		"CloudFoundryDeleteSpace":                           CloudFoundryDeleteSpaceMetadata(),
		"CloudFoundryDeploy":                                CloudFoundryDeployMetadata(),
		"CnbBuild":                                          CnbBuildMetadata(),
		"CodeqlExecuteScan":                                 CodeqlExecuteScanMetadata(),
		"ContainerExecuteStructureTests":                    ContainerExecuteStructureTestsMetadata(),
		"ContainerSaveImage":                                ContainerSaveImageMetadata(),
		"ContrastExecuteScan":                               ContrastExecuteScanMetadata(),
		"CredentialdiggerScan":                              CredentialdiggerScanMetadata(),
		"DetectExecuteScan":                                 DetectExecuteScanMetadata(),
		"FortifyExecuteScan":                                FortifyExecuteScanMetadata(),
		"GaugeExecuteTests":                                 GaugeExecuteTestsMetadata(),
		"GctsCloneRepository":                               GctsCloneRepositoryMetadata(),
		"GctsCreateRepository":                              GctsCreateRepositoryMetadata(),
		"GctsDeploy":                                        GctsDeployMetadata(),
		"GctsExecuteABAPQualityChecks":                      GctsExecuteABAPQualityChecksMetadata(),
		"GctsExecuteABAPUnitTests":                          GctsExecuteABAPUnitTestsMetadata(),
		"GctsRollback":                                      GctsRollbackMetadata(),
		"GithubCheckBranchProtection":                       GithubCheckBranchProtectionMetadata(),
		"GithubCommentIssue":                                GithubCommentIssueMetadata(),
		"GithubCreateIssue":                                 GithubCreateIssueMetadata(),
		"GithubCreatePullRequest":                           GithubCreatePullRequestMetadata(),
		"GithubPublishRelease":                              GithubPublishReleaseMetadata(),
		"GithubSetCommitStatus":                             GithubSetCommitStatusMetadata(),
		"GitopsUpdateDeployment":                            GitopsUpdateDeploymentMetadata(),
		"GolangBuild":                                       GolangBuildMetadata(),
		"GradleExecuteBuild":                                GradleExecuteBuildMetadata(),
		"HadolintExecute":                                   HadolintExecuteMetadata(),
		"HelmExecute":                                       HelmExecuteMetadata(),
		"ImagePushToRegistry":                               ImagePushToRegistryMetadata(),
		"InfluxWriteData":                                   InfluxWriteDataMetadata(),
		"IntegrationArtifactDeploy":                         IntegrationArtifactDeployMetadata(),
		"IntegrationArtifactDownload":                       IntegrationArtifactDownloadMetadata(),
		"IntegrationArtifactGetMplStatus":                   IntegrationArtifactGetMplStatusMetadata(),
		"IntegrationArtifactGetServiceEndpoint":             IntegrationArtifactGetServiceEndpointMetadata(),
		"IntegrationArtifactResource":                       IntegrationArtifactResourceMetadata(),
		"IntegrationArtifactTransport":                      IntegrationArtifactTransportMetadata(),
		"IntegrationArtifactTriggerIntegrationTest":         IntegrationArtifactTriggerIntegrationTestMetadata(),
		"IntegrationArtifactUnDeploy":                       IntegrationArtifactUnDeployMetadata(),
		"IntegrationArtifactUpdateConfiguration":            IntegrationArtifactUpdateConfigurationMetadata(),
		"IntegrationArtifactUpload":                         IntegrationArtifactUploadMetadata(),
		"IsChangeInDevelopment":                             IsChangeInDevelopmentMetadata(),
		"JsonApplyPatch":                                    JsonApplyPatchMetadata(),
		"KanikoExecute":                                     KanikoExecuteMetadata(),
		"KarmaExecuteTests":                                 KarmaExecuteTestsMetadata(),
		"KubernetesDeploy":                                  KubernetesDeployMetadata(),
		"MalwareExecuteScan":                                MalwareExecuteScanMetadata(),
		"MavenBuild":                                        MavenBuildMetadata(),
		"MavenExecute":                                      MavenExecuteMetadata(),
		"MavenExecuteIntegration":                           MavenExecuteIntegrationMetadata(),
		"MavenExecuteStaticCodeChecks":                      MavenExecuteStaticCodeChecksMetadata(),
		"MtaBuild":                                          MtaBuildMetadata(),
		"NewmanExecute":                                     NewmanExecuteMetadata(),
		"NexusUpload":                                       NexusUploadMetadata(),
		"NpmExecuteLint":                                    NpmExecuteLintMetadata(),
		"NpmExecuteScripts":                                 NpmExecuteScriptsMetadata(),
		"PipelineCreateScanSummary":                         PipelineCreateScanSummaryMetadata(),
		"ProtecodeExecuteScan":                              ProtecodeExecuteScanMetadata(),
		"PythonBuild":                                       PythonBuildMetadata(),
		"ShellExecute":                                      ShellExecuteMetadata(),
		"SonarExecuteScan":                                  SonarExecuteScanMetadata(),
		"TerraformExecute":                                  TerraformExecuteMetadata(),
		"TmsExport":                                         TmsExportMetadata(),
		"TmsUpload":                                         TmsUploadMetadata(),
		"TransportRequestDocIDFromGit":                      TransportRequestDocIDFromGitMetadata(),
		"TransportRequestReqIDFromGit":                      TransportRequestReqIDFromGitMetadata(),
		"TransportRequestUploadCTS":                         TransportRequestUploadCTSMetadata(),
		"TransportRequestUploadRFC":                         TransportRequestUploadRFCMetadata(),
		"TransportRequestUploadSOLMAN":                      TransportRequestUploadSOLMANMetadata(),
		"UiVeri5ExecuteTests":                               UiVeri5ExecuteTestsMetadata(),
		"VaultRotateSecretId":                               VaultRotateSecretIdMetadata(),
		"WhitesourceExecuteScan":                            WhitesourceExecuteScanMetadata(),
		"XsDeploy":                                          XsDeployMetadata(),
	}
}
