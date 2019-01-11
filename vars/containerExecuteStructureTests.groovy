import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    /**
     * Only for Kubernetes environments: Command which is executed to keep container alive.
     */
    'containerCommand',
    /**
     * Only for Kubernetes environments: Shell to be used inside container.
     */
    'containerShell',
    /**
      * Docker image for code execution.
      */
    'dockerImage',
    /**
      * Defines the behavior, in case tests fail.
      * @possibleValues `true`, `false`
      */
    'failOnError',
    /**
      * If specific stashes should be considered for the tests, you can pass this via this parameter.
      */
    'stashContent',
    /**
     * Container structure test configuration in yml or json format. You can pass a pattern in order to execute multiple tests.
     */
    'testConfiguration'
]

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
            .use()

        utils.pushToSWA([step: STEP_NAME], config)

        config.stashContent = utils.unstashAll(config.stashContent)

        List testConfig = findFiles(glob: config.newmanCollection)?.toList()
        if (testConfig.isEmpty()) {
            error "[${STEP_NAME}] No collection found with pattern '${config.newmanCollection}'"
        } else {
            echo "[${STEP_NAME}] Found files ${testConfig}"
        }

        def testConfigArgs
        testConfig.each {conf ->
            testConfigArgs += "--config ${conf} "
        }

        --config config.yaml

        dockerExecute(
            script: script,
            containerCommand: config.containerCommand,
            containerShell: config.containerShell,
            dockerImage: config.dockerImage,
            stashContent: config.stashContent
        ) {
            sh "container-structure-test test --image gcr.io/registry/image:latest ${testConfigArgs}"
        }
    }
}
