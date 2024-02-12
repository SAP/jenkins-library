import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.StageNameProvider
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'performanceTests'

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Executes Gatling performance tests */
    'gatlingExecuteTests',
    /** Can perform both to cloud foundry and neo targets. Preferred over cloudFoundryDeploy and neoDeploy, if configured. */
    'multicloudDeploy',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this stage important performance-relevant checks will be conducted.<br />
 *
 * The stage will execute a Gatling test, if the step `gatlingExecuteTests` is configured.
 */
@GenerateStageDocumentation(defaultStageName = 'Performance')
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()
    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('multicloudDeploy', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.multicloudDeploy)
        .addIfEmpty('gatlingExecuteTests', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.gatlingExecuteTests)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {
        if (config.multicloudDeploy) {
            durationMeasure(script: script, measurementName: 'deploy_performance_multicloud_duration') {
                multicloudDeploy(script: script, stage: stageName)
            }
        }

        if (config.gatlingExecuteTests) {
            durationMeasure(script: script, measurementName: 'gatling_duration') {
                gatlingExecuteTests script: script
            }
        }
    }
}
