import groovy.transform.Field
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Creates a SAP BTP ABAP Environment instance via the cloud foundry command line interface */
    'abapEnvironmentCreateSystem',
    /** Creates Communication Arrangements for ABAP Environment instance via the cloud foundry command line interface */
    'cloudFoundryCreateServiceKey',
    /** Creates a BTP service instance for ABAP Environment */
    'btpCreateServiceInstance'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage prepares the SAP BTP ABAP Environment systems
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STAGE_STEP_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .use()

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {
        if (abapEnvironmentPipelineHelpers.isBTPMode(config)) {
            // BTP path: Create BTP service instance
            btpCreateServiceInstance script: parameters.script
        } else {
            // Cloud Foundry path: Use existing approach
            abapEnvironmentCreateSystem script: parameters.script
        }
    }

}
