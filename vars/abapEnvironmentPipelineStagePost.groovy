import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Deletes a SAP Cloud Platform ABAP Environment instance via the cloud foundry command line interface */
    'cloudFoundryDeleteService',
    /** If this parameter is set to true, a manual confirmation is required to delete the system if the pipeline status is UNSUCCESFUL*/
    'confirmSystemDeletionWhenUnsuccessful'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
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
        .use()

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {
        if(parameters.script.commonPipelineEnvironment.configuration.runStage?.get("Prepare System")) {
            if (config.confirmSystemDeletionWhenUnsuccessful && currentBuild.result == 'UNSUCCESFUL') {
                input(message: 'The pipeline was unsuccesful. Do you want to delete the system?')
            }
            cloudFoundryDeleteService script: parameters.script
        }
    }
}
