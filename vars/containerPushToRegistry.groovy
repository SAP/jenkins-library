import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.DockerUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript
import static com.sap.piper.BashUtils.quoteAndEscape as q

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the id of the Jenkins username/password credentials containing the credentials for the target Docker registry.
     */
    'dockerCredentialsId',
    /** Defines the registry url where the image should be pushed to, incl. the protocol like `https://my.registry.com`*/
    'dockerRegistryUrl',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /** Not supported yet - Docker archive to be pushed to registry*/
    'dockerArchive',
    /** For images built locally on the Docker Deamon, reference to the image object resulting from `docker.build` execution */
    'dockerBuildImage',
    /** Defines the name (incl. tag) of the target image*/
    'dockerImage',
    /**
     * Only if no Docker daemon available on your Jenkins image: Docker image to be used for [Skopeo](https://github.com/containers/skopeo) calls
     * Unfortunately no proper image known to be available.
     * Simple custom Dockerfile could look as follows: <br>
     * ```
     * FROM fedora:29
     * RUN dnf install -y skopeo
     * ```
     */
    'skopeoImage',
    /** Defines the name (incl. tag) of the source image to be pushed to a new image defined in `dockerImage`.<br>
     * This is helpful for moving images from one location to another.
     */
    'sourceImage',
    /** Defines a registry url from where the image should optionally be pulled from, incl. the protocol like `https://my.registry.com`*/
    'sourceRegistryUrl',
    /** Defines the id of the Jenkins username/password credentials containing the credentials for the source Docker registry. */
    'sourceCredentialsId',
    /** Defines if the image should be tagged as `latest`*/
    'tagLatest',
    /** Defines if the image should be tagged with the artifact version */
    'tagArtifactVersion'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This step allows you to push a Docker image into a dedicated Container registry.
 *
 * By default an image available via the local Docker daemon will be pushed.
 *
 * In case you want to pull an existing image from a remote container registry, a source image and source registry needs to be specified.<br />
 * This makes it possible to move an image from one registry to another.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        final script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('sourceImage', script.commonPipelineEnvironment.getValue('containerImage'))
            .addIfEmpty('sourceRegistryUrl', script.commonPipelineEnvironment.getValue('containerRegistryUrl'))
            .mixin(artifactVersion: script.commonPipelineEnvironment.getArtifactVersion())
            .withMandatoryProperty('dockerCredentialsId')
            .withMandatoryProperty('dockerRegistryUrl')
            .use()

        DockerUtils dockerUtils = new DockerUtils(script)

        if (config.sourceRegistryUrl) {
            config.sourceRegistry = dockerUtils.getRegistryFromUrl(config.sourceRegistryUrl)
        }

        if (!config.dockerImage)
            config.dockerImage = config.sourceImage

        if (dockerUtils.withDockerDaemon()) {

            //Prevent NullPointerException in case no dockerImage nor dockerBuildImage is provided
            if (!config.dockerImage && !config.dockerBuildImage) {
                error "[${STEP_NAME}] Please provide a dockerImage (either in your config.yml or via step parameter)."
            }
            config.dockerBuildImage = config.dockerBuildImage?:docker.image(config.dockerImage)

            if (config.sourceRegistry && config.sourceImage) {

                def sourceBuildImage = docker.image(config.sourceImage)
                docker.withRegistry(
                    config.sourceRegistryUrl,
                    config.sourceCredentialsId
                ) {
                    sourceBuildImage.pull()
                }
                sh "docker tag ${q(config.sourceRegistry)}/${q(config.sourceImage)} ${q(config.dockerImage)}"
            }

            docker.withRegistry(
                config.dockerRegistryUrl,
                config.dockerCredentialsId
            ) {
                config.dockerBuildImage.push()
                if (config.tagLatest)
                    config.dockerBuildImage.push('latest')
                if (config.tagArtifactVersion )
                    config.dockerBuildImage.push(config.artifactVersion)
            }
        } else {
            //handling for Kubernetes case
            dockerExecute(
                script: script,
                dockerImage: config.skopeoImage
            ) {

                if (!config.dockerArchive && !config.dockerBuildImage) {
                    dockerUtils.moveImage([image: config.sourceImage, registryUrl: config.sourceRegistryUrl, credentialsId: config.sourceCredentialsId], [image: config.dockerImage, registryUrl: config.dockerRegistryUrl, credentialsId: config.dockerCredentialsId])
                    if (config.tagLatest) {
                        def latestImage = "${config.dockerImage.split(':')[0]}:latest"
                        dockerUtils.moveImage([image: config.sourceImage, registryUrl: config.sourceRegistryUrl, credentialsId: config.sourceCredentialsId], [image: latestImage, registryUrl: config.dockerRegistryUrl, credentialsId: config.dockerCredentialsId])
                    }
                    if (config.tagArtifactVersion) {
                        def imageName = "${config.dockerImage.split(':')[0]}:${config.artifactVersion}"
                        dockerUtils.moveImage([image: config.sourceImage, registryUrl: config.sourceRegistryUrl, credentialsId: config.sourceCredentialsId], [image: imageName, registryUrl: config.dockerRegistryUrl, credentialsId: config.dockerCredentialsId])
                    }
                } else {
                    error "[${STEP_NAME}] Running on Kubernetes: Only moving images from one registry to another supported."
                }
            }
        }
    }
}
