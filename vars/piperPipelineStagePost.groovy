import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.StageNameProvider
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'postPipelineHook'
@Field Set GENERAL_CONFIG_KEYS = ["vaultServerUrl", "vaultAppRoleTokenCredentialsId", "vaultAppRoleSecretTokenCredentialsId"]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(["vaultRotateSecretId"])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this stage reporting actions like mail notification or telemetry reporting are executed.
 *
 * This stage contains following steps:
 * - [vaultRotateSecretId](./vaultRotateSecretId.md)
 * - [influxWriteData](./influxWriteData.md)
 * - [mailSendNotification](./mailSendNotification.md)
 *
 * !!! note
 *     This stage is meant to be used in a [post](https://jenkins.io/doc/book/pipeline/syntax/#post) section of a pipeline.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()
    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)
    // ease handling extension
    stageName = stageName.replace('Declarative: ', '')

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults([:], stageName)
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty("vaultRotateSecretId", false)
        .use()

    piperStageWrapper (script: script, stageName: stageName, stageLocking: false) {
        // rotate vault secret id if necessary
        if (config.vaultRotateSecretId && config.vaultServerUrl && config.vaultAppRoleSecretTokenCredentialsId
            && config.vaultAppRoleTokenCredentialsId) {
            vaultRotateSecretId script: script
        }

        influxWriteData script: script
        if(env.BRANCH_NAME == parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch) {
            if(parameters.script.commonPipelineEnvironment.configuration.runStep?.get('Post Actions')?.slackSendNotification) {
                slackSendNotification script: parameters.script
            }
        }

        mailSendNotification script: script
        debugReportArchive script: script
        piperPublishWarnings script: script
    }
}
