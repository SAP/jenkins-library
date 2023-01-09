import groovy.transform.Field
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    'abapEnvironmentAssemblyKitPublishTargetVector',
    'testBuild' // Parameter for test execution mode, if true stage will be skipped
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage publishes an AddOn for the SAP BTP ABAP Environment
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('testBuild', false)
        .use()

    if (config.testBuild) {
        echo "Stage 'Publish' skipped as parameter 'testBuild' is active"
    } else {
        piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {
            abapAddonAssemblyKitPublishTargetVector(script: parameters.script, targetVectorScope: 'P')
        }
    }
}
