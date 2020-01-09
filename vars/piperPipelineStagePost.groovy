import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this stage reporting actions like mail notification or telemetry reporting are executed.
 *
 * This stage contains following steps:
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
    def stageName = parameters.stageName?:env.STAGE_NAME
    // ease handling extension
    stageName = stageName.replace('Declarative: ', '')
    Map config = ConfigurationHelper.newInstance(this, script)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .use()

    piperStageWrapper (script: script, stageName: stageName, stageLocking: false) {
        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        influxWriteData script: script

        if(env.BRANCH_NAME == parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch) {
            if(parameters.script.commonPipelineEnvironment.configuration.runStep?.get('Post Actions')?.slackSendNotification) {
                slackSendNotification script: parameters.script
            }
        }
        mailSendNotification script: script
        piperPublishWarnings script: script
    }
}
