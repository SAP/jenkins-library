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
		"abapEnvironmentCheckoutBranch":             abapEnvironmentCheckoutBranchMetadata(),
		"abapEnvironmentCloneGitRepo":               abapEnvironmentCloneGitRepoMetadata(),
		"abapEnvironmentCreateSystem":               abapEnvironmentCreateSystemMetadata(),
		"abapEnvironmentPullGitRepo":                abapEnvironmentPullGitRepoMetadata(),
		"abapEnvironmentRunATCCheck":                abapEnvironmentRunATCCheckMetadata(),
		"batsExecuteTests":                          batsExecuteTestsMetadata(),
		"checkChangeInDevelopment":                  checkChangeInDevelopmentMetadata(),
		"checkmarxExecuteScan":                      checkmarxExecuteScanMetadata(),
		"cloudFoundryCreateService":                 cloudFoundryCreateServiceMetadata(),
		"cloudFoundryCreateServiceKey":              cloudFoundryCreateServiceKeyMetadata(),
		"cloudFoundryCreateSpace":                   cloudFoundryCreateSpaceMetadata(),
		"cloudFoundryDeleteService":                 cloudFoundryDeleteServiceMetadata(),
		"cloudFoundryDeleteSpace":                   cloudFoundryDeleteSpaceMetadata(),
		"cloudFoundryDeploy":                        cloudFoundryDeployMetadata(),
		"containerExecuteStructureTests":            containerExecuteStructureTestsMetadata(),
		"detectExecuteScan":                         detectExecuteScanMetadata(),
		"fortifyExecuteScan":                        fortifyExecuteScanMetadata(),
		"gaugeExecuteTests":                         gaugeExecuteTestsMetadata(),
		"gctsCloneRepository":                       gctsCloneRepositoryMetadata(),
		"gctsCreateRepository":                      gctsCreateRepositoryMetadata(),
		"gctsDeploy":                                gctsDeployMetadata(),
		"gctsExecuteABAPUnitTests":                  gctsExecuteABAPUnitTestsMetadata(),
		"gctsRollback":                              gctsRollbackMetadata(),
		"githubCheckBranchProtection":               githubCheckBranchProtectionMetadata(),
		"githubCommentIssue":                        githubCommentIssueMetadata(),
		"githubCreateIssue":                         githubCreateIssueMetadata(),
		"githubCreatePullRequest":                   githubCreatePullRequestMetadata(),
		"githubPublishRelease":                      githubPublishReleaseMetadata(),
		"githubSetCommitStatus":                     githubSetCommitStatusMetadata(),
		"gitopsUpdateDeployment":                    gitopsUpdateDeploymentMetadata(),
		"hadolintExecute":                           hadolintExecuteMetadata(),
		"influxWriteData":                           influxWriteDataMetadata(),
		"integrationArtifactDeploy":                 integrationArtifactDeployMetadata(),
		"integrationArtifactDownload":               integrationArtifactDownloadMetadata(),
		"integrationArtifactGetMplStatus":           integrationArtifactGetMplStatusMetadata(),
		"integrationArtifactGetServiceEndpoint":     integrationArtifactGetServiceEndpointMetadata(),
		"integrationArtifactTriggerIntegrationTest": integrationArtifactTriggerIntegrationTestMetadata(),
		"integrationArtifactUnDeploy":               integrationArtifactUnDeployMetadata(),
		"integrationArtifactUpdateConfiguration":    integrationArtifactUpdateConfigurationMetadata(),
		"integrationArtifactUpload":                 integrationArtifactUploadMetadata(),
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
		"containerSaveImage":                        containerSaveImageMetadata(),
		"sonarExecuteScan":                          sonarExecuteScanMetadata(),
		"terraformExecute":                          terraformExecuteMetadata(),
		"transportRequestDocIDFromGit":              transportRequestDocIDFromGitMetadata(),
		"transportRequestReqIDFromGit":              transportRequestReqIDFromGitMetadata(),
		"transportRequestUploadCTS":                 transportRequestUploadCTSMetadata(),
		"transportRequestUploadRFC":                 transportRequestUploadRFCMetadata(),
		"transportRequestUploadSOLMAN":              transportRequestUploadSOLMANMetadata(),
		"uiVeri5ExecuteTests":                       uiVeri5ExecuteTestsMetadata(),
		"vaultRotateSecretId":                       vaultRotateSecretIdMetadata(),
		"artifactPrepareVersion":                    artifactPrepareVersionMetadata(),
		"whitesourceExecuteScan":                    whitesourceExecuteScanMetadata(),
		"xsDeploy":                                  xsDeployMetadata(),
	}
}
