import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.GitUtils
import com.sap.piper.Utils

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Map which defines per docker image the port mappings, e.g. `containerPortMappings: ['selenium/standalone-chrome': [[name: 'selPort', containerPort: 4444, hostPort: 4444]]]`.
     */
    'containerPortMappings',
    /** A map of environment variables to set in the container, e.g. [http_proxy:'proxy:8080']. */
    'dockerEnvVars',
    /** The name of the docker image that should be used. If empty, Docker is not used and the command is executed directly on the Jenkins system. */
    'dockerImage',
    /**
     * Kubernetes only:
     * Name of the container launching `dockerImage`.
     * SideCar only:
     * Name of the container in local network.
     */
    'dockerName',
    /**
     * Kubernetes only:
     * Specifies a dedicated user home directory for the container which will be passed as value for environment variable `HOME`.
     */
    'dockerWorkspace',
    /**
     * With `failOnError` the behavior in case tests fail can be defined.
     * @possibleValues `true`, `false`
     */
    'failOnError',
    /** The command that is executed to install the test tool. */
    'installCommand',
    /** Define the paths of the modules to execute tests on. */
    'modules',
    /** The command that is executed to start the tests. */
    'runCommand',
    /** A map of environment variables to set in the sidecar container, similar to `dockerEnvVars`. */
    'sidecarEnvVars',
    /** The name of the docker image of the sidecar container. If empty, no sidecar container is started. */
    'sidecarImage',
    /**
     * as `dockerName` for the sidecar container
     */
    'sidecarName',
    /** Volumes that should be mounted into the sidecar container. */
    'sidecarVolumeBind',
    /** If specific stashes should be considered for the tests, their names need to be passed via the parameter `stashContent`. */
    'stashContent'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this step the ([Karma test runner](http://karma-runner.github.io)) is executed.
 *
 * The step is using the `seleniumExecuteTest` step to spin up two containers in a Docker network:
 *
 * * a Selenium/Chrome container (`selenium/standalone-chrome`)
 * * a NodeJS container (`node:8-stretch`)
 *
 * In the Docker network, the containers can be referenced by the values provided in `dockerName` and `sidecarName`, the default values are `karma` and `selenium`. These values must be used in the `hostname` properties of the test configuration ([Karma](https://karma-runner.github.io/1.0/config/configuration-file.html) and [WebDriver](https://github.com/karma-runner/karma-webdriver-launcher#usage)).
 *
 * !!! note
 *     In a Kubernetes environment, the containers both need to be referenced with `localhost`.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        final script = checkScript(this, parameters) ?: this
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

        def testJobs = [:]
        def options = [
            script: script,
            containerPortMappings: config.containerPortMappings,
            dockerEnvVars: config.dockerEnvVars,
            dockerImage: config.dockerImage,
            dockerName: config.dockerName,
            dockerWorkspace: config.dockerWorkspace,
            failOnError: config.failOnError,
            sidecarEnvVars: config.sidecarEnvVars,
            sidecarImage: config.sidecarImage,
            sidecarName: config.sidecarName,
            sidecarVolumeBind: config.sidecarVolumeBind,
            stashContent: config.stashContent
        ]
        for(String path : config.modules){
            String modulePath = path
            testJobs["Karma - ${modulePath}"] = {
                seleniumExecuteTests(options){
                    sh "cd '${modulePath}' && ${config.installCommand}"
                    sh "cd '${modulePath}' && ${config.runCommand}"
                }
            }
        }

        if(testJobs.size() == 1){
            testJobs.each({ key, value -> value() })
        }else{
            parallel testJobs.plus([failFast: false])
        }
    }
}
