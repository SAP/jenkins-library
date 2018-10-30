import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = 'seleniumExecuteTests'
@Field Set STEP_CONFIG_KEYS = [
    'buildTool', //defines the tool which is used for executing the tests
    'containerPortMappings', //port mappings required for containers. This will only take effect inside a Kubernetes pod, format [[containerPort: 1111, hostPort: 1111]]
    'dockerEnvVars', //envVars to be set in the execution container if required
    'dockerImage', //Docker image for code execution
    'dockerName', //name of the Docker container. This will only take effect inside a Kubernetes pod.
    'dockerWorkspace', //user home directory for Docker execution. This will only take effect inside a Kubernetes pod.
    'failOnError',
    'gitBranch', //only if testRepository is used: branch of testRepository. Default is master
    'gitSshKeyCredentialsId', //only if testRepository is used: ssh credentials id in case a protected testRepository is used
    'sidecarEnvVars', //envVars to be set in Selenium container if required
    'sidecarImage', //image for Selenium execution which runs as sidecar to dockerImage
    'sidecarName', //name of the Selenium container. If not on Kubernetes pod, this will define the name of the link to the Selenium container and is thus required for accessing the server, example http://selenium:4444 (default)
    'sidecarVolumeBind', //volume bind. This will not take effect in Kubernetes pod.
    'stashContent', //list of stash names which are required to be unstashed before test run
    'testRepository' //if tests are in a separate repository, git url can be defined. For protected repositories the git ssh url is required
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:], Closure body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: [commonPipelineEnvironment: commonPipelineEnvironment]
        def utils = parameters?.juStabUtils ?: new Utils()

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .dependingOn('buildTool').mixin('dockerImage')
            .dependingOn('buildTool').mixin('dockerName')
            .dependingOn('buildTool').mixin('dockerWorkspace')
            .use()

        utils.pushToSWA([step: STEP_NAME,
                         stepParam1: parameters?.script == null], config)

        dockerExecute(
                script: script,
                containerPortMappings: config.containerPortMappings,
                dockerEnvVars: config.dockerEnvVars,
                dockerImage: config.dockerImage,
                dockerName: config.dockerName,
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
