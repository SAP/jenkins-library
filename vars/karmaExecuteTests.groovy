import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field String STEP_NAME = 'karmaExecuteTests'
@Field Set GENERAL_CONFIG_KEYS = [
    /** port mappings required for containers. This will only take effect inside a Kubernetes pod, format [[containerPort: 1111, hostPort: 1111]] */
    'containerPortMappings',
    /** envVars to be set in the execution container if required */
    'dockerEnvVars',
    /** Docker image for code execution */
    'dockerImage',
    /** name of the Docker container. If not on Kubernetes pod, this will define the network-alias to the NPM container and is thus required for accessing the server, example http://karma:9876 (default). */
    'dockerName',
    /** user home directory for Docker execution. This will only take effect inside a Kubernetes pod. */
    'dockerWorkspace',
    'failOnError',
    'installCommand',
    'modules',
    'runCommand',
    /** envVars to be set in Selenium container if required */
    'sidecarEnvVars',
    /** image for Selenium execution which runs as sidecar to dockerImage */
    'sidecarImage',
    /** name of the Selenium container. If not on Kubernetes pod, this will define the network-alias to the Selenium container and is thus required for accessing the server, example http://selenium:4444 (default) */
    'sidecarName',
    /** volume bind. This will not take effect in Kubernetes pod. */
    'sidecarVolumeBind',
    /** list of stash names which are required to be unstashed before test run */
    'stashContent'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

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
            testJobs["Karma - ${path}"] = {
                seleniumExecuteTests(options){
                    sh "cd '${path}' && ${config.installCommand}"
                    sh "cd '${path}' && ${config.runCommand}"
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
