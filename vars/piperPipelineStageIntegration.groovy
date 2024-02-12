import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.StageNameProvider
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'integration'

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Runs npm scripts to run generic integration tests written on JavaScript */
    'npmExecuteScripts',
    /** Runs backend integration tests via the Jacoco Maven-plugin */
    'mavenExecuteIntegration',
    /** Publishes test results to Jenkins. It will automatically be active in cases tests are executed. */
    'testsPublishResults',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * The stage allows to execute project-specific integration tests.<br />
 * Typically, integration tests are very project-specific, thus they can be defined here using the [stage extension mechanism](../extensibility.md).
 */
@GenerateStageDocumentation(defaultStageName = 'Integration')
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()
    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('npmExecuteScripts', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteScripts)
        .addIfEmpty('mavenExecuteIntegration', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.mavenExecuteIntegration)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {
        boolean publishResults = false
        try {
            writeTemporaryCredentials(script: script) {
                if (config.npmExecuteScripts) {
                    publishResults = true
                    npmExecuteScripts script: script
                }
                if (config.mavenExecuteIntegration) {
                    publishResults = true
                    mavenExecuteIntegration script: script
                }
            }
        }
        finally {
            if (publishResults) {
                testsPublishResults script: script
            }
        }
    }
}
