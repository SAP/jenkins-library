import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.StageNameProvider
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'additionalUnitTests'

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Executes bats tests which are for example suitable for testing Docker images via a shell.*/
    'batsExecuteTests',
    /** Executes karma tests which are for example suitable for OPA5 testing as well as QUnit testing of SAP UI5 apps.*/
    'karmaExecuteTests',
    /** Executes npm scripts to run frontend unit tests */
    'npmExecuteScripts',
    /** Publishes test results to Jenkins. It will automatically be active in cases tests are executed. */
    'testsPublishResults'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this stage unit tests, which can not or should not be executed in the central build environment, are executed.<br />
 * These are for example Karma(OPA5 & QUnit) tests.
 */
@GenerateStageDocumentation(defaultStageName = 'Additional Unit Tests')
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()
    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('batsExecuteTests', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.batsExecuteTests)
        .addIfEmpty('karmaExecuteTests', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.karmaExecuteTests)
        .addIfEmpty('npmExecuteScripts', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteScripts)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {
        if (config.batsExecuteTests) {
            durationMeasure(script: script, measurementName: 'bats_duration') {
                batsExecuteTests script: script
                testsPublishResults script: script
            }
        }

        if (config.karmaExecuteTests) {
            durationMeasure(script: script, measurementName: 'karma_duration') {
                karmaExecuteTests script: script
                testsPublishResults script: script
            }
        }

        if (config.npmExecuteScripts) {
            durationMeasure(script: script, measurementName: 'npmExecuteScripts_duration') {
                npmExecuteScripts script: script
                testsPublishResults script: script
            }
        }
    }
}
