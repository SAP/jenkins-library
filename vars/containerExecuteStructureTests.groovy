import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Print more detailed information into the log.
     * @possibleValues `true`, `false`
     */
    'verbose'

]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Only for Kubernetes environments: Command which is executed to keep container alive, defaults to '/usr/bin/tail -f /dev/null'
     */
    'containerCommand',
    /**
     * Only for Kubernetes environments: Shell to be used inside container, defaults to '/bin/sh'
     */
    'containerShell',
    /**
      * Docker image for code execution.
      */
    'dockerImage',
    /**
     * Options to be passed to Docker image when starting it (only relevant for non-Kubernetes case).
     */
    'dockerOptions',
    /**
      * Defines the behavior, in case tests fail.
      * @possibleValues `true`, `false`
      */
    'failOnError',
    /**
     * Only relevant for testDriver 'docker'.
     * @possibleValues `true`, `false`
     */
    'pullImage',
    /**
      * If specific stashes should be considered for the tests, you can pass this via this parameter.
      */
    'stashContent',
    /**
     * Container structure test configuration in yml or json format. You can pass a pattern in order to execute multiple tests.
     */
    'testConfiguration',
    /**
     * Container structure test driver to be used for testing, please see https://github.com/GoogleContainerTools/container-structure-test for details.
     */
    'testDriver',
    /**
    * Image to be tested
    */
    'testImage',
    /**
     * Path and name of the test report which will be generated
     */
    'testReportFilePath',
])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        def utils = parameters?.juStabUtils ?: new Utils()

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('testDriver', Boolean.valueOf(script.env.ON_K8S) ? 'tar' : 'docker')
            .addIfNull('pullImage', !Boolean.valueOf(script.env.ON_K8S))
            .withMandatoryProperty('dockerImage')
            .use()

        utils.pushToSWA([step: STEP_NAME], config)

        config.stashContent = utils.unstashAll(config.stashContent)

        List testConfig = findFiles(glob: config.testConfiguration)?.toList()
        if (testConfig.isEmpty()) {
            error "[${STEP_NAME}] No test description found with pattern '${config.testConfiguration}'"
        } else {
            echo "[${STEP_NAME}] Found files ${testConfig}"
        }

        def testConfigArgs = ''
        testConfig.each {conf ->
            testConfigArgs += "--config ${conf} "
        }

        //workaround for non-working '--pull' option in version 1.7.0 of container-structure-tests, see https://github.com/GoogleContainerTools/container-structure-test/issues/193
        if (config.pullImage) {
            if (config.verbose) echo "[${STEP_NAME}] Pulling image since configuration option pullImage is set to '${config.pullImage}'"
            sh "docker pull ${config.testImage}"
        }

        try {
            dockerExecute(
                script: script,
                containerCommand: config.containerCommand,
                containerShell: config.containerShell,
                dockerImage: config.dockerImage,
                dockerOptions: config.dockerOptions,
                stashContent: config.stashContent
            ) {
                sh """#!${config.containerShell?:'/bin/sh'}
container-structure-test test ${testConfigArgs} --driver ${config.testDriver} --image ${config.testImage} --test-report ${config.testReportFilePath}${config.verbose ? ' --verbosity debug' : ''}"""
            }
        } catch (err) {
            echo "[${STEP_NAME}] Test execution failed"
            script.currentBuild.result = 'UNSTABLE'
            if (config.failOnError) throw err
        } finally {
            echo "${readFile(config.testReportFilePath)}"
            archiveArtifacts artifacts: config.testReportFilePath, allowEmptyArchive: true
        }

    }
}
