import com.sap.piper.ContainerUtils

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    // Username/password credentials for container Registry.
    'dockerCredentialsId',
    // Full http(s) url of the target docker registry.
    'dockerRegistryUrl'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /** Docker image object created in the pipeline using docker.build() */
    'dockerBuildImage',
    /** Name of the target image including path (if applicable) and tag. */
    'targetImage',
    /** Name of the source image including path (if applicable) and tag. If not set and running on a Docker deamon, a locally available image will be used as defined with `dockerImage`*/
    'sourceImage',
    /** Full http(s) url of the docker registry of the source Image. If not set, a dockerImage available in the local daemon will be used. */
    'sourceRegistryUrl',
    /**
     * Tag the image with `latest` tag when pushing to registry.
     * @possibleValues `true`, `false`
     */
    'tagLatest'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(
                dockerBuildImage: script.commonPipelineEnvironment.getDockerBuildImage(),

                //ToDo: mix in source registry and source image information from commonPipelineEnvironment if available
                //sourceImage: script.commonPipelineEnvironment. ....
                //sourceRegistryUrl: script.commonPipelineEnvironment. ...
            )
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('dockerCredentialsId')
            .withMandatoryProperty('dockerRegistryUrl')
            .use()

        ContainerUtils containerUtils = new ContainerUtils(this)

        if (config.sourceRegistryUrl) {
            config.sourceRegistry = containerUtils.getRegistryFromUrl(config.sourceRegistryUrl)
        }

        // report to SWA
        utils.pushToSWA([step: STEP_NAME], config)

        if (!config.dockerImage)
            config.dockerImage = config.sourceImage

        if (containerUtils.withDockerDeamon()) {

            //ToDo: evaluate if option dockerBuildImage can be removed, if not make dockerImage mandatory if no dockerBuildImage is available, otherwise we will run into a NullPointerException here!
            config.dockerBuildImage = config.dockerBuildImage?:docker.image(config.dockerImage)
            ConfigurationHelper.newInstance(this, config)
                .withMandatoryProperty('dockerBuildImage')

            if (config.sourceRegistry && config.sourceImage) {

                def sourceBuildImage = docker.image(config.sourceImage)
                docker.withRegistry(config.sourceRegistryUrl) {
                    sourceBuildImage.pull()
                }
                sh "docker tag ${config.sourceRegistry}/${config.sourceImage} ${config.dockerImage}"
            }

            docker.withRegistry(
                config.dockerRegistryUrl,
                config.dockerCredentialsId
            ) {
                config.dockerBuildImage.push()
                if (config.tagLatest)
                    config.dockerBuildImage.push('latest')
            }
        } else {
            //handling for Kubernetes case
            dockerExecute(
                script: script,
                dockerImage: 'docker.wdf.sap.corp:50000/piper/skopeo'
            ) {
                def sourceImageFullName = (config.sourceRegistry ? "${config.sourceRegistry}/" : '') + config.sourceImage
                def targetImageFullName = (config.dockerRegistryUrl ? "${containerUtils.getRegistryFromUrl(config.dockerRegistryUrl)}/" : '') + config.dockerImage

                withCredentials([usernamePassword(
                    credentialsId: config.dockerCredentialsId,
                    passwordVariable: 'password',
                    usernameVariable: 'userid'
                )]) {
                    containerUtils.skopeoMoveImage(sourceImageFullName, targetImageFullName, userid, password)
                }
            }
        }
    }
}
