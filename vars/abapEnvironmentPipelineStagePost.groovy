import groovy.transform.Field
import com.sap.piper.ConfigurationHelper

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /** Deletes a SAP Cloud Platform ABAP Environment instance via the cloud foundry command line interface */
    'cloudFoundryDeleteService',
    /** If set to true, a confirmation is required to delete the system in case the pipeline was not successful */
    'confirmDeletion',
    /** If set to true, the system is never deleted */
    'debug'
]
@Field Set STAGE_STEP_KEYS = GENERAL_CONFIG_KEYS
@Field Set STEP_CONFIG_KEYS = STAGE_STEP_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage cleans up the ABAP Environment Pipeline run
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def stageName = parameters.stageName?:env.STAGE_NAME
    stageName = stageName.replace('Declarative: ', '')
    stageName = stageName.replace(' Actions', '')

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('confirmDeletion', false)
        .addIfEmpty('debug', false)
        .use()

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {
        if(parameters.script.commonPipelineEnvironment.configuration.runStage?.get("Prepare System")) {

            if (config.confirmDeletion && script.currentBuild.result != 'SUCCESS') {
                input message: "Pipeline status is not successful. Once you proceed, the system will be deleted."
            }
            if (!config.debug) {
                cloudFoundryDeleteService script: parameters.script
            }
        }
    }
}
