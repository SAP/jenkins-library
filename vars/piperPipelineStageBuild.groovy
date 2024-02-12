import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.StageNameProvider
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'build'

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Starts build execution. This is always being executed.*/
    'buildExecute',
    /**
     * Executes stashing of files after build execution.<br /
     * Build results are stashed with stash name `buildResult`.
     *
     * **Note: Please make sure that your build artifacts are contained here since this stash is the foundation for subsequent tests and checks, e.g. deployment to a test landscape.**
     **/
    'pipelineStashFilesAfterBuild',
    /** Executes a Sonar scan.*/
    'sonarExecuteScan',
    /** Publishes test results to Jenkins. It will always be active. */
    'testsPublishResults',
    /** Publishes check results to Jenkins. It will always be active. */
    'checksPublishResults',
    /** Executes static code checks for Maven based projects. The plugins SpotBugs and PMD are used. */
    'mavenExecuteStaticCodeChecks',
    /** Executes linting for npm projects. */
    'npmExecuteLint'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this stage a build is executed which typically also executes tests and code checks.
 *
 * The type of build is defined using the configuration `buildTool`, see also step [buildExecute](../steps/buildExecute.md)
 *
 */
@GenerateStageDocumentation(defaultStageName = 'Build')
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()
    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('npmExecuteLint', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteLint)
        .addIfEmpty('mavenExecuteStaticCodeChecks', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.mavenExecuteStaticCodeChecks)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {
        durationMeasure(script: script, measurementName: 'build_duration') {

            buildExecute script: script
            pipelineStashFilesAfterBuild script: script

            try {
                testsPublishResults script: script, junit: [updateResults: true]
                checksPublishResults script: script
            } finally {
                if (config.sonarExecuteScan) {
                    sonarExecuteScan script: script
                }
            }
        }

        if (config.mavenExecuteStaticCodeChecks) {
            durationMeasure(script: script, measurementName: 'staticCodeChecks_duration') {
                mavenExecuteStaticCodeChecks(script: script)
            }
        }

        if (config.npmExecuteLint) {
            durationMeasure(script: script, measurementName: 'npmExecuteLint_duration') {
                npmExecuteLint script: script
            }
        }
    }
}
