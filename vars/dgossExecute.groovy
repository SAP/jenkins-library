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

    /** The tag of the image to dgoss. */
    'dockerImageTag',
    /**
     * gossFile The path to the goss file to use. Default value is 'goss.yaml'.
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
        // handle deprecated parameters
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        DockerUtils dockerUtils = new DockerUtils(script)
        ConfigurationHelper.newInstance(this, config)
            .withMandatoryProperty('dockerImageName')
            .withMandatoryProperty('dockerImageTag')
            .withMandatoryProperty('gossFile')

        def dockerImageNameAndTag = "${config.dockerImageName}:${config.dockerImageTag}"
        if (config.dockerRegistryUrl) {
            dockerImageNameAndTag = config.dockerRegistryUrl + "/" + dockerImageNameAndTag
        }

        if (dockerUtils.onKubernetes()){
            runOnK8S(config, dockerImageNameAndTag)

        }else if (dockerUtils.withDockerDaemon()){
            runOnNode(dockerImageNameAndTag)
        }else{
            error "[${STEP_NAME}] No Docker daemon available, dgoss require a runnig docker daemon"
        }
    }
}

def runOnNode(dockerImageNameAndTag){
    def targetImage = docker.image(dockerImageNameAndTag)
    docker.image('docker:18.06.3-dind').withRun('--privileged -it --name mydind') { c ->
        sh "while ! docker exec mydind docker stats --no-stream; do sleep 1; done"
        docker.image('kiwicom/dgoss').inside(""" --link ${c.id}:docker -v "${pwd()}":/src
        -e "GOSS_FILES_STRATEGY=cp"
        -e "DOCKER_HOST=tcp://docker:2375" """) {
            sh """
                cd /src
                dgoss run ${dockerImageNameAndTag}
            """
        }
    }
}

void runOnK8S(config, dockerImageNameAndTag) {
    stash name: '_gossfile', includes: config.gossFile
    podTemplate(yaml:"""
apiVersion: v1
kind: Pod
metadata:
    name: dgoss
    labels:
        name: dgoss
spec:
    volumes:
    - name: dind-storage
        emptyDir: {}
    containers:
    - name: dind
        image: docker:18.06.3-dind
        securityContext:
        privileged: true
        volumeMounts:
        - name: dind-storage
            mountPath: /var/lib/docker
    - name: goss
        image: kiwicom/dgoss
        env:
        - name: DOCKER_HOST
        value: tcp://localhost:2375
        - name: GOSS_FILES_STRATEGY
        value: cp
        command:
        - cat
        tty: true
    - name: jnlp
        image: docker.wdf.sap.corp:50001/sap-production/jnlp-alpine:3.26.1-sap-02
        args: ['\$(JENKINS_SECRET)', '\$(JENKINS_NAME)']
            """
    ){
        node(POD_LABEL){
            container('goss') {
                unstash '_gossfile'
                if (config.gossFile != "goss.yaml") {
                    sh "cp ${config.gossFile} goss.yaml"
                }
                sh "dgoss run ${dockerImageNameAndTag}"
            }
        }
    }
}

boolean isKubernetes() {
    return Boolean.valueOf(env.ON_K8S)
}
