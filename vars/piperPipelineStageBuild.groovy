import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

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
    /** Publishes test results to Jenkins. It will always be active. */
    'testsPublishResults',
    /** Publishes check results to Jenkins. It will always be active. */
    'checksPublishResults'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this stage a build is executed which typically also executes tests and code checks.
 *
 * They type of build is defined using the configuration `buildTool`, see also step [buildExecute](../steps/buildExecute.md)
 *
 */
@GenerateStageDocumentation(defaultStageName = 'Build')
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper.newInstance(this, script)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        durationMeasure(script: script, measurementName: 'build_duration') {

            buildExecute script: script
            pipelineStashFilesAfterBuild script: script

            testsPublishResults script: script, junit: [updateResults: true]
            checksPublishResults script: script
        }
    }
}
