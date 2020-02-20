import com.sap.piper.DockerUtils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.text.GStringTemplateEngine
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the tool used for the build.
     * @possibleValues `docker`, `kaniko`, `maven`, `mta`, `npm`
     */
    'buildTool',
    /** For Docker builds only (mandatory): name of the image to be built. */
    'dockerImageName',
    /** For Docker builds only: Defines the registry url where the image should be pushed to, incl. the protocol like `https://my.registry.com`. If it is not defined, image will not be pushed to a registry.*/
    'dockerRegistryUrl',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([

    /** Only for Docker builds on the local deamon: Defines the build options for the build.*/
    'containerBuildOptions',
    /** For custom build types: Defines the command to be executed within the `dockerImage` in order to execute the build. */
    'dockerCommand',
    /** For custom build types: Image to be used for builds in case they should run inside a custom Docker container */
    'dockerImage',
    /** For Docker builds only (mandatory): tag of the image to be built. */
    'dockerImageTag',
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This step serves as generic entry point in pipelines for building artifacts.
 *
 * You can use pre-defined `buildTool`s.
 *
 * Alternatively you can define a command via `dockerCommand` which should be executed in `dockerImage`.<br />
 * This allows you to trigger any build tool using a defined Docker container which provides the required build infrastructure.
 *
 * When using `buildTool: docker` or `buildTool: kaniko` the created container image is uploaded to a container registry.<br />
 * You need to make sure that the required credentials are provided to the step.
 *
 * For all other `buildTool`s the artifact will just be stored in the workspace and could then be `stash`ed for later use.
 *
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        final script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // handle deprecated parameters
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('dockerImageTag', script.commonPipelineEnvironment.getArtifactVersion())
            .use()

        // telemetry reporting
        utils.pushToSWA([stepParam1: config.buildTool, 'buildTool': config.buildTool], config)

        switch(config.buildTool){
            case 'maven':
                mavenExecute script: script
                break
            case 'mta':
                mtaBuild script: script
                break
            case 'npm':
                npmExecute script: script
                break
            case ['docker', 'kaniko']:
                DockerUtils dockerUtils = new DockerUtils(script)
                if (config.buildTool == 'docker' && !dockerUtils.withDockerDaemon()) {
                    config.buildTool = 'kaniko'
                    echo "[${STEP_NAME}] No Docker daemon available, thus switching to Kaniko build"
                }

                ConfigurationHelper.newInstance(this, config)
                    .withMandatoryProperty('dockerImageName')
                    .withMandatoryProperty('dockerImageTag')

                def dockerImageNameAndTag = "${config.dockerImageName}:${config.dockerImageTag}"

                if (config.buildTool == 'kaniko') {
                    def containerImageNameAndTag = config.dockerRegistryUrl ? "${dockerUtils.getRegistryFromUrl(config.dockerRegistryUrl)}/${dockerImageNameAndTag}" : ''
                    kanikoExecute script: script, containerImageNameAndTag: containerImageNameAndTag
                } else {
                    def dockerBuildImage = docker.build(dockerImageNameAndTag, "${config.containerBuildOptions ?: ''} .")
                    //only push if registry is defined
                    if (config.dockerRegistryUrl) {
                        containerPushToRegistry script: script, dockerBuildImage: dockerBuildImage, dockerRegistryUrl: config.dockerRegistryUrl
                    }
                }
                script.commonPipelineEnvironment.setValue('containerImage', dockerImageNameAndTag)
                break
            default:
                if (config.dockerImage && config.dockerCommand) {
                    dockerExecute(
                        script: script,
                        dockerImage: config.dockerImage,
                    ) {
                        sh "${config.dockerCommand}"
                    }
                } else {
                    error "[${STEP_NAME}] buildTool not set and no dockerImage & dockerCommand provided."
                }
        }
    }
}
