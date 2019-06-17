import com.sap.piper.DockerUtils
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
//import com.sap.piper.Notify

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /** Defines the tool used for the build. */
    'buildTool',
    /** For Docker builds only (mandatory): name of the image to be built. */
    'dockerImageName',
    /** For Docker builds only (mandatory): Defines the registry url where the image should be pushed to, incl. the protocol like `https://my.registry.com`*/
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

void call(Map parameters = [:], body = '') {
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

        DockerUtils dockerUtils = new DockerUtils(script)
        if (config.buildTool == 'docker' && !dockerUtils.withDockerDeamon()) {
            config.buildTool = 'kaniko'
            echo "[${STEP_NAME}] no Docker deamon available, thus switching to Kaniko build"
        }

        // report to SWA
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
                ConfigurationHelper.newInstance(this, config)
                    .withMandatoryProperty('dockerImageName')
                    .withMandatoryProperty('dockerImageTag')
                    .withMandatoryProperty('dockerRegistryUrl')

                def dockerImageNameAndTag = "${config.dockerImageName}:${config.dockerImageTag}"

                if (config.buildTool == 'kaniko') {
                    kanikoExecute script: script, containerImageNameAndTag: "${dockerUtils.getRegistryFromUrl(config.dockerRegistryUrl)}/${dockerImageNameAndTag}"
                } else {
                    def dockerBuildImage = docker.build(dockerImageNameAndTag, "${config.containerBuildOptions} .")
                    containerPushToRegistry script: this, dockerBuildImage: dockerBuildImage, dockerRegistryUrl: config.dockerRegistryUrl
                }
                commonPipelineEnvironment.setValue('containerImage', dockerImageNameAndTag)
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
                    error("[${STEP_NAME}] buildTool not set and no dockerImage & dockerCommand provided")
                }
        }
    }
}


