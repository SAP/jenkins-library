import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.StageNameProvider
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'mavenExecuteStaticCodeChecks'

@Field STAGE_STEP_KEYS = [
    /** Executes static code checks for Maven based projects. The plugins SpotBugs and PMD are used. */
    'mavenExecuteStaticCodeChecks'
]
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Execute static code checks for Maven based projects.
 *
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: null
    def utils = parameters.juStabUtils ?: new Utils()
    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('mavenExecuteStaticCodeChecks', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.mavenExecuteStaticCodeChecks)
        .use()

    piperStageWrapper(stageName: stageName, script: script) {
        if (config.mavenExecuteStaticCodeChecks) {
            mavenExecuteStaticCodeChecks(script: script)
        }
    }
}
