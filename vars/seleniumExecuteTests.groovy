import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field
import groovy.text.GStringTemplateEngine

@Field String STEP_NAME = getClass().getName()

//TODO: limit parameter visibility
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the tool which is used for executing the tests
     * @possibleValues `maven`, `npm`, `bundler`
     */
    'buildTool',
    /** @see dockerExecute */
    'containerPortMappings',
    /** @see dockerExecute */
    'dockerEnvVars',
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerName',
    /** @see dockerExecute */
    'dockerOptions',
    /** @see dockerExecute */
    'dockerWorkspace',
    /**
     * With `failOnError` the behavior in case tests fail can be defined.
     * @possibleValues `true`, `false`
     */
    'failOnError',
    /**
     * Only if `testRepository` is provided: Branch of testRepository, defaults to master.
     */
    'gitBranch',
    /**
     * Only if `testRepository` is provided: Credentials for a protected testRepository
     * @possibleValues Jenkins credentials id
     */
    'gitSshKeyCredentialsId',
    /** @see dockerExecute */
    'sidecarEnvVars',
    /** @see dockerExecute */
    'sidecarImage',
    /** @see dockerExecute */
    'sidecarName',
    /** @see dockerExecute */
    'sidecarVolumeBind',
    /** @see dockerExecute */
    'stashContent',
    /**
     * Define an additional repository where the test implementation is located.
     * For protected repositories the `testRepository` needs to contain the ssh git url.
     */
    'testRepository'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Enables UI test execution with Selenium in a sidecar container.
 *
 * The step executes a closure (see example below) connecting to a sidecar container with a Selenium Server.
 *
 * When executing in a
 *
 * * local Docker environment, please make sure to set Selenium host to **`selenium`** in your tests.
 * * Kubernetes environment, plese make sure to set Seleniums host to **`localhost`** in your tests.
 *
 * !!! note "Proxy Environments"
 *     If work in an environment containing a proxy, please make sure that `localhost`/`selenium` is added to your proxy exclusion list, e.g. via environment variable `NO_PROXY` & `no_proxy`. You can pass those via parameters `dockerEnvVars` and `sidecarEnvVars` directly to the containers if required.
 */
@GenerateDocumentation
void call(Map parameters = [:], Closure body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters?.juStabUtils ?: new Utils()

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this, script)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .dependingOn('buildTool').mixin('dockerImage')
            .dependingOn('buildTool').mixin('dockerName')
            .dependingOn('buildTool').mixin('dockerWorkspace')
            .use()

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], config)

        dockerExecute(
                script: script,
                containerPortMappings: config.containerPortMappings,
                dockerEnvVars: config.dockerEnvVars,
                dockerImage: config.dockerImage,
                dockerName: config.dockerName,
                dockerOptions: config.dockerOptions,
                dockerWorkspace: config.dockerWorkspace,
                sidecarEnvVars: config.sidecarEnvVars,
                sidecarImage: config.sidecarImage,
                sidecarName: config.sidecarName,
                sidecarVolumeBind: config.sidecarVolumeBind
        ) {
            try {
                config.stashContent = config.testRepository
                    ?[GitUtils.handleTestRepository(this, config)]
                    :utils.unstashAll(config.stashContent)
                body()
            } catch (err) {
                if (config.failOnError) {
                    throw err
                }
            }
        }
    }
}
