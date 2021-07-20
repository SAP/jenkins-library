import com.sap.piper.DockerUtils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /** The name of the docker image to dgoss. */
    'dockerImageName',
    /** Defines the registry url where the image should be located. */
    'dockerRegistryUrl',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /** The tag of the image to test. */
    'dockerImageTag',
    /**
     * The relative path to the goss file to use. Default value is 'goss.yaml'.
     */
    'gossFile'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This step execute goss validation agains your container.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        final script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        ConfigurationHelper.newInstance(this, config)
            .withMandatoryProperty('dockerImageName')
            .withMandatoryProperty('dockerImageTag')
            .withMandatoryProperty('gossFile')

        DockerUtils dockerUtils = new DockerUtils(script)

        def dockerImageNameAndTag = "${config.dockerImageName}:${config.dockerImageTag}"
        if (config.dockerRegistryUrl) {
            dockerImageNameAndTag = config.dockerRegistryUrl + "/" + dockerImageNameAndTag
        }

        if (dockerUtils.onKubernetes()){
            runOnK8S(config, dockerImageNameAndTag)

        }else if (dockerUtils.withDockerDaemon()){
            runOnNode(config, dockerImageNameAndTag)
        }else{
            error "[${STEP_NAME}] No Docker daemon available, dgoss require a runnig docker daemon"
        }
    }
}

void runOnNode(config, dockerImageNameAndTag){
    def targetImage

    docker.image('aelsabbahy/goss').withRun("--name goss"){ gossc ->

        docker.image(dockerImageNameAndTag).withRun("""--volumes-from goss -v "${pwd()}":"${pwd()}" """) { c ->
            sh """
            docker exec "${c.id}" sh -c "/goss/goss -g ${pwd()}/${config.gossFile} validate --format documentation"
            """
        }
    }


}

void runOnK8S(config, dockerImageNameAndTag) {
    stash name: '_gossfile', includes: config.gossFile
    podTemplate(
yaml:"""
apiVersion: v1
kind: Pod
metadata:
    labels:
        name: goss
spec:
    volumes:
        -
            name: shared-data
            emptyDir: {}
    containers:
        -
            name: goss
            image: aelsabbahy/goss
            volumeMounts:
                -
                    name: shared-data
                    mountPath: /shared-goss
            command:
                - cat
            tty: true
        -
            name: executor
            image: '${dockerImageNameAndTag}'
            volumeMounts:
                -
                    name: shared-data
                    mountPath: /shared-goss

"""
    ){
        node(POD_LABEL){
            container('goss') {
                unstash '_gossfile'
                sh """
                cp /goss/* /shared-goss
                """
            }
            container('executor') {
                unstash '_gossfile'
                sh """
                /shared-goss/goss -g ${config.gossFile} validate --format documentation
                """
            }
        }
    }
}

boolean isKubernetes() {
    return Boolean.valueOf(env.ON_K8S)
}
