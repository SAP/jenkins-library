import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.StageNameProvider
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'pullRequestVoting'

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
    /** Runs backend integration tests via maven in the module integration-tests/pom.xml */
    'mavenExecuteIntegration',
    /** Executes static code checks for Maven based projects. The plugins SpotBugs and PMD are used. */
    'mavenExecuteStaticCodeChecks',
    /** Executes linting for npm projects. */
    'npmExecuteLint',
    /** Executes npm scripts to run frontend unit tests.
     * If custom names for the npm scripts are configured via the `runScripts` parameter the step npmExecuteScripts needs **explicit activation via stage configuration**. */
    'npmExecuteScripts',
    /** Executes a Sonar scan.*/
    'sonarExecuteScan',
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
    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('karmaExecuteTests', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.karmaExecuteTests)
        .addIfEmpty('mavenExecuteIntegration', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.mavenExecuteIntegration)
        .addIfEmpty('mavenExecuteStaticCodeChecks', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.mavenExecuteStaticCodeChecks)
        .addIfEmpty('npmExecuteLint', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteLint)
        .addIfEmpty('npmExecuteScripts', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteScripts)
        .addIfEmpty('whitesourceExecuteScan', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.whitesourceExecuteScan)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {
        durationMeasure(script: script, measurementName: 'voter_duration') {

            //prevent push to registry in case of docker/kaniko
            def dockerRegistryUrl = null
            if (config.buildTool in ['docker', 'kaniko']) {
                dockerRegistryUrl = ''
            }

            buildExecute script: script, buildTool: config.buildTool, dockerRegistryUrl: dockerRegistryUrl
            try {
                //needs to run right after build, otherwise we may face "ERROR: Test reports were found but none of them are new"
                testsPublishResults script: script
                checksPublishResults script: script
            } finally {
                if (config.sonarExecuteScan) {
                    sonarExecuteScan script: script
                }
            }

            if (config.karmaExecuteTests) {
                karmaExecuteTests script: script
                testsPublishResults script: script
            }

            if (config.mavenExecuteIntegration) {
                runMavenIntegrationTests(script)
            }

            if (config.mavenExecuteStaticCodeChecks) {
                mavenExecuteStaticCodeChecks script: script
            }

            if (config.npmExecuteLint) {
                npmExecuteLint script: script
            }

            if (config.npmExecuteScripts) {
                npmExecuteScripts script: script
                testsPublishResults script: script
            }

            if (config.whitesourceExecuteScan) {
                whitesourceExecuteScan script: script, productVersion: env.BRANCH_NAME
            }
        }
    }
}

private runMavenIntegrationTests(script){
    boolean publishResults = false
    try {
        writeTemporaryCredentials(script: script) {
            publishResults = true
            mavenExecuteIntegration script: script
        }
    }
    finally {
        if (publishResults) {
            testsPublishResults script: script
        }
    }
}
