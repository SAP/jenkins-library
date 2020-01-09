import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the build tool used.
     * @possibleValues `docker`, `kaniko`, `maven`, `mta`, `npm`
     */
    'buildTool'
]
@Field STAGE_STEP_KEYS = [
    /** Triggers the build execution. */
    'buildExecute',
    /** Publishes check results to Jenkins. It will always be active. */
    'checksPublishResults',
    /**
     * Executes karma tests. For example suitable for OPA5 testing as well as QUnit testing of SAP UI5 apps.<br />
     * This step is not active by default. It can be activated by:
     *
     * * using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting).
     * * explicit activation via stage configuration.
     */
    'karmaExecuteTests',
    /** Publishes test results to Jenkins. It will always be active. */
    'testsPublishResults',
    /** Executes a WhiteSource scan
     * This step is not active by default. It can be activated by:
     *
     * * using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting).
     * * explicit activation via stage configuration.
     */
    'whitesourceExecuteScan'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This stage is executed for every pull-request.<br />
 * For non-Docker builds it will execute the respective build (including unit tests, static checks, ...).
 *
 * !!! note "Build Tool not in the list?"
 *
 *     For build tools which are currently not in the list a custom `dockerImage` can be used with a custom `dockerCommand` as per step [buildExecute](../steps/buildExecute.md)
 *
 * For `buildTool: docker` a local Docker build will be executed in case a Docker deamon is available, if not `buildTool: 'kaniko'` will be used instead.
 *
 * ## Advanced Pull-Request Voting
 *
 * It is possible to trigger dedicated tests/checks
 *
 * * pull request comments
 * * pull request labels
 *
 * Following steps are currently supported
 *
 * | step name | comment | pull-request label |
 * | --------- | ------- | ------------------ |
 * | karmaExecuteTests | `/piper karma` | `pr_karma`
 * | whitesourceExecuteScan | `/piper whitesource` | `pr_whitesource`
 *
 */
@GenerateStageDocumentation(defaultStageName = 'Pull-Request Voting')
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper.newInstance(this, script)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('karmaExecuteTests', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.karmaExecuteTests)
        .addIfEmpty('whitesourceExecuteScan', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.whitesourceExecuteScan)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        durationMeasure(script: script, measurementName: 'voter_duration') {

            //prevent push to registry in case of docker/kaniko
            def dockerRegistryUrl = null
            if (config.buildTool in ['docker', 'kaniko']) {
                dockerRegistryUrl = ''
            }

            buildExecute script: script, buildTool: config.buildTool, dockerRegistryUrl: dockerRegistryUrl

            //needs to run right after build, otherwise we may face "ERROR: Test reports were found but none of them are new"
            testsPublishResults script: script
            checksPublishResults script: script

            if (config.karmaExecuteTests) {
                karmaExecuteTests script: script
                testsPublishResults script: script
            }

            if (config.whitesourceExecuteScan) {
                whitesourceExecuteScan script: script, productVersion: env.BRANCH_NAME
            }
        }
    }
}
