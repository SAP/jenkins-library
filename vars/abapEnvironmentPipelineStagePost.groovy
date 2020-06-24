import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Deletes a SAP Cloud Platform ABAP Environment instance via the cloud foundry command line interface */
    'cloudFoundryDeleteService'
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

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {
        if(parameters.script.commonPipelineEnvironment.configuration.runStage?.get("Prepare System")) {
            cloudFoundryDeleteService script: parameters.script
        }
    }

}
