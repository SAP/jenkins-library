import groovy.transform.Field
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Creates a SAP Cloud Platform ABAP Environment system via the cloud foundry command line interface */
    'abapEnvironmentCreateSystem',
    /** Deletes a SAP Cloud Platform ABAP Environment system via the cloud foundry command line interface */
    'cloudFoundryDeleteService',
    /** If set to true, a confirmation is required to delete the system */
    'confirmDeletion'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage creates a system for Integration Tests. The (custom) tests themselves can be added via a stage extension.
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('confirmDeletion', true)
        .use()

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {
        try {
            abapEnvironmentCreateSystem(script: parameters.script, includeAddon: true)
            if (config.confirmDeletion) {
                input message: "Add-on product was installed successfully? Once you proceed, the test system will be deleted."
                cloudFoundryDeleteService script: parameters.script
            }
        } catch (Exception e) {
            script.currentBuild.result = 'UNSTABLE'
        }
    }

}
