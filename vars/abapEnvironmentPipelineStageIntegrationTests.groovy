import groovy.transform.Field
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /** Creates a SAP Cloud Platform ABAP Environment system via the cloud foundry command line interface */
    'abapEnvironmentCreateSystem',
    /** Deletes a SAP Cloud Platform ABAP Environment system via the cloud foundry command line interface */
    'cloudFoundryDeleteService',
    /** If set to true, a confirmation is required to delete the system */
    'confirmDeletion',
    /** If set to true, the system is never deleted */
    'debug'
]
@Field Set STAGE_STEP_KEYS = GENERAL_CONFIG_KEYS
@Field Set STEP_CONFIG_KEYS = STAGE_STEP_KEYS
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
        .addIfEmpty('debug', false)
        .use()

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {
        try {
            abapEnvironmentCreateSystem(script: parameters.script, includeAddon: true)
        } catch (Exception e) {
            echo "Deployment test of add-on product failed."
            throw e
        } finally {
            if (config.confirmDeletion) {
                input message: "Deployment test has been executed. Once you proceed, the test system will be deleted."
            }
            if (!config.debug) {
                cloudFoundryDeleteService script: parameters.script
            }
        }
    }

}
