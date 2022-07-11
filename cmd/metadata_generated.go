// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import "github.com/SAP/jenkins-library/pkg/config"

// GetStepMetadata return a map with all the step metadata mapped to their stepName
func GetAllStepMetadata() map[string]config.StepData {
	return map[string]config.StepData{
		"abapAddonAssemblyKitCheckCVs":              abapAddonAssemblyKitCheckCVsMetadata(),
		"abapAddonAssemblyKitCheckPV":               abapAddonAssemblyKitCheckPVMetadata(),
		"abapAddonAssemblyKitCreateTargetVector":    abapAddonAssemblyKitCreateTargetVectorMetadata(),
		"abapAddonAssemblyKitPublishTargetVector":   abapAddonAssemblyKitPublishTargetVectorMetadata(),
		"abapAddonAssemblyKitRegisterPackages":      abapAddonAssemblyKitRegisterPackagesMetadata(),
		"abapAddonAssemblyKitReleasePackages":       abapAddonAssemblyKitReleasePackagesMetadata(),
		"abapAddonAssemblyKitReserveNextPackages":   abapAddonAssemblyKitReserveNextPackagesMetadata(),
		"abapEnvironmentAssembleConfirm":            abapEnvironmentAssembleConfirmMetadata(),
		"abapEnvironmentAssemblePackages":           abapEnvironmentAssemblePackagesMetadata(),
		"abapEnvironmentBuild":                      abapEnvironmentBuildMetadata(),
		"abapEnvironmentCheckoutBranch":             abapEnvironmentCheckoutBranchMetadata(),
		"abapEnvironmentCloneGitRepo":               abapEnvironmentCloneGitRepoMetadata(),
		"abapEnvironmentCreateSystem":               abapEnvironmentCreateSystemMetadata(),
		"abapEnvironmentCreateTag":                  abapEnvironmentCreateTagMetadata(),
		"abapEnvironmentPullGitRepo":                abapEnvironmentPullGitRepoMetadata(),
		"abapEnvironmentPushATCSystemConfig":        abapEnvironmentPushATCSystemConfigMetadata(),
		"abapEnvironmentRunATCCheck":                abapEnvironmentRunATCCheckMetadata(),
		"abapEnvironmentRunAUnitTest":               abapEnvironmentRunAUnitTestMetadata(),
		"ansSendEvent":                              ansSendEventMetadata(),
		"apiKeyValueMapDownload":                    apiKeyValueMapDownloadMetadata(),
		"apiKeyValueMapUpload":                      apiKeyValueMapUploadMetadata(),
		"apiProviderDownload":                       apiProviderDownloadMetadata(),
		"apiProviderList":                           apiProviderListMetadata(),
		"apiProviderUpload":                         apiProviderUploadMetadata(),
		"apiProxyDownload":                          apiProxyDownloadMetadata(),
		"apiProxyList":                              apiProxyListMetadata(),
		"apiProxyUpload":                            apiProxyUploadMetadata(),
		"artifactPrepareVersion":                    artifactPrepareVersionMetadata(),
		"awsS3Upload":                               awsS3UploadMetadata(),
		"azureBlobUpload":                           azureBlobUploadMetadata(),
		"batsExecuteTests":                          batsExecuteTestsMetadata(),
		"checkmarxExecuteScan":                      checkmarxExecuteScanMetadata(),
		"cloudFoundryCreateService":                 cloudFoundryCreateServiceMetadata(),
		"cloudFoundryCreateServiceKey":              cloudFoundryCreateServiceKeyMetadata(),
		"cloudFoundryCreateSpace":                   cloudFoundryCreateSpaceMetadata(),
		"cloudFoundryDeleteService":                 cloudFoundryDeleteServiceMetadata(),
		"cloudFoundryDeleteSpace":                   cloudFoundryDeleteSpaceMetadata(),
		"cloudFoundryDeploy":                        cloudFoundryDeployMetadata(),
		"cnbBuild":                                  cnbBuildMetadata(),
		"codeqlExecuteScan":                         codeqlExecuteScanMetadata(),
		"containerExecuteStructureTests":            containerExecuteStructureTestsMetadata(),
		"containerSaveImage":                        containerSaveImageMetadata(),
		"detectExecuteScan":                         detectExecuteScanMetadata(),
		"fortifyExecuteScan":                        fortifyExecuteScanMetadata(),
		"gaugeExecuteTests":                         gaugeExecuteTestsMetadata(),
		"gctsCloneRepository":                       gctsCloneRepositoryMetadata(),
		"gctsCreateRepository":                      gctsCreateRepositoryMetadata(),
		"gctsDeploy":                                gctsDeployMetadata(),
		"gctsExecuteABAPQualityChecks":              gctsExecuteABAPQualityChecksMetadata(),
		"gctsExecuteABAPUnitTests":                  gctsExecuteABAPUnitTestsMetadata(),
		"gctsRollback":                              gctsRollbackMetadata(),
		"githubCheckBranchProtection":               githubCheckBranchProtectionMetadata(),
		"githubCommentIssue":                        githubCommentIssueMetadata(),
		"githubCreateIssue":                         githubCreateIssueMetadata(),
		"githubCreatePullRequest":                   githubCreatePullRequestMetadata(),
		"githubPublishRelease":                      githubPublishReleaseMetadata(),
		"githubSetCommitStatus":                     githubSetCommitStatusMetadata(),
		"gitopsUpdateDeployment":                    gitopsUpdateDeploymentMetadata(),
		"golangBuild":                               golangBuildMetadata(),
		"gradleExecuteBuild":                        gradleExecuteBuildMetadata(),
		"hadolintExecute":                           hadolintExecuteMetadata(),
		"helmExecute":                               helmExecuteMetadata(),
		"influxWriteData":                           influxWriteDataMetadata(),
		"integrationArtifactDeploy":                 integrationArtifactDeployMetadata(),
		"integrationArtifactDownload":               integrationArtifactDownloadMetadata(),
		"integrationArtifactGetMplStatus":           integrationArtifactGetMplStatusMetadata(),
		"integrationArtifactGetServiceEndpoint":     integrationArtifactGetServiceEndpointMetadata(),
		"integrationArtifactResource":               integrationArtifactResourceMetadata(),
		"integrationArtifactTriggerIntegrationTest": integrationArtifactTriggerIntegrationTestMetadata(),
		"integrationArtifactUnDeploy":               integrationArtifactUnDeployMetadata(),
		"integrationArtifactUpdateConfiguration":    integrationArtifactUpdateConfigurationMetadata(),
		"integrationArtifactUpload":                 integrationArtifactUploadMetadata(),
		"isChangeInDevelopment":                     isChangeInDevelopmentMetadata(),
		"jsonApplyPatch":                            jsonApplyPatchMetadata(),
		"kanikoExecute":                             kanikoExecuteMetadata(),
		"karmaExecuteTests":                         karmaExecuteTestsMetadata(),
		"kubernetesDeploy":                          kubernetesDeployMetadata(),
		"malwareExecuteScan":                        malwareExecuteScanMetadata(),
		"mavenBuild":                                mavenBuildMetadata(),
		"mavenExecute":                              mavenExecuteMetadata(),
		"mavenExecuteIntegration":                   mavenExecuteIntegrationMetadata(),
		"mavenExecuteStaticCodeChecks":              mavenExecuteStaticCodeChecksMetadata(),
		"mtaBuild":                                  mtaBuildMetadata(),
		"newmanExecute":                             newmanExecuteMetadata(),
		"nexusUpload":                               nexusUploadMetadata(),
		"npmExecuteLint":                            npmExecuteLintMetadata(),
		"npmExecuteScripts":                         npmExecuteScriptsMetadata(),
		"pipelineCreateScanSummary":                 pipelineCreateScanSummaryMetadata(),
		"protecodeExecuteScan":                      protecodeExecuteScanMetadata(),
		"pythonBuild":                               pythonBuildMetadata(),
		"shellExecute":                              shellExecuteMetadata(),
		"sonarExecuteScan":                          sonarExecuteScanMetadata(),
		"terraformExecute":                          terraformExecuteMetadata(),
		"transportRequestDocIDFromGit":              transportRequestDocIDFromGitMetadata(),
		"transportRequestReqIDFromGit":              transportRequestReqIDFromGitMetadata(),
		"transportRequestUploadCTS":                 transportRequestUploadCTSMetadata(),
		"transportRequestUploadRFC":                 transportRequestUploadRFCMetadata(),
		"transportRequestUploadSOLMAN":              transportRequestUploadSOLMANMetadata(),
		"uiVeri5ExecuteTests":                       uiVeri5ExecuteTestsMetadata(),
		"vaultRotateSecretId":                       vaultRotateSecretIdMetadata(),
		"whitesourceExecuteScan":                    whitesourceExecuteScanMetadata(),
		"xsDeploy":                                  xsDeployMetadata(),
	}
}
