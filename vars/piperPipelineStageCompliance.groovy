import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.StageNameProvider
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'compliance'

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Executes a SonarQube scan */
    'sonarExecuteScan',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this stage important compliance-relevant checks will be conducted.<br />
 *
 * The stage will execute a SonarQube scan, if the step `sonarExecuteSan` is configured.
 */
@GenerateStageDocumentation(defaultStageName = 'Compliance')
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()
    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('sonarExecuteScan', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.sonarExecuteScan)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {
        if (config.sonarExecuteScan) {
            durationMeasure(script: script, measurementName: 'sonar_duration') {
                sonarExecuteScan script: script
            }
        }
    }
}
