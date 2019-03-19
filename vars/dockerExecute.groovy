import static com.sap.piper.Prerequisites.checkScript

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field def PLUGIN_ID_DOCKER_WORKFLOW = 'docker-workflow'

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     *
     */
    'jenkinsKubernetes'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Kubernetes only:
     * Allows to specify start command for container created with dockerImage parameter to overwrite Piper default (`/usr/bin/tail -f /dev/null`).
     */
    'containerCommand',
    /**
     * Map which defines per docker image the port mappings, e.g. `containerPortMappings: ['selenium/standalone-chrome': [[name: 'selPort', containerPort: 4444, hostPort: 4444]]]`.
     */
    'containerPortMappings',
    /**
     * Kubernetes only:
     * Allows to specify the shell to be used for execution of commands.
     */
    'containerShell',
    /**
     * Environment variables to set in the container, e.g. [http_proxy: 'proxy:8080'].
     */
    'dockerEnvVars',
    /**
     * Name of the docker image that should be used. If empty, Docker is not used and the command is executed directly on the Jenkins system.
     */
    'dockerImage',
    /**
     * Kubernetes only:
     * Name of the container launching `dockerImage`.
     * SideCar only:
     * Name of the container in local network.
     */
    'dockerName',
    /**
     * Docker options to be set when starting the container (List or String).
     */
    'dockerOptions',
    /**
     * Volumes that should be mounted into the container.
     */
    'dockerVolumeBind',
    /**
     * Set this to 'false' to bypass a docker image pull. Usefull during development process. Allows testing of images which are available in the local registry only.
     */
    'dockerPullImage',
    /**
     * Kubernetes only:
     * Specifies a dedicated user home directory for the container which will be passed as value for environment variable `HOME`.
     */
    'dockerWorkspace',
    /**
     * as `dockerEnvVars` for the sidecar container
     */
    'sidecarEnvVars',
    /**
     * as `dockerImage` for the sidecar container
     */
    'sidecarImage',
    /**
     * as `dockerName` for the sidecar container
     */
    'sidecarName',
    /**
     * as `dockerOptions` for the sidecar container
     */
    'sidecarOptions',
    /**
     * as `dockerVolumeBind` for the sidecar container
     */
    'sidecarVolumeBind',
    /**
     * Set this to 'false' to bypass a docker image pull. Usefull during development process. Allows testing of images which are available in the local registry only.
     */
    'sidecarPullImage',
    /**
     * as `dockerWorkspace` for the sidecar container
     */
    'sidecarWorkspace',
    /**
     * Command executed inside the container which returns exit code 0 when the container is ready to be used.
     */
    'sidecarReadyCommand',
    /**
     * Specific stashes that should be considered for the step execution.
     */
    'stashContent'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Executes a closure inside a docker container with the specified docker image.
 * The workspace is mounted into the docker image.
 * Proxy environment variables defined on the Jenkins machine are also available in the Docker container.
 */
@GenerateDocumentation
void call(Map parameters = [:], body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: this

        def utils = parameters?.juStabUtils ?: new Utils()

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null,
            stepParamKey2: 'kubernetes',
            stepParam2: isKubernetes()
        ], config)

        if (isKubernetes() && config.dockerImage) {
            if (env.POD_NAME && isContainerDefined(config)) {
                container(getContainerDefined(config)) {
                    echo "[INFO][${STEP_NAME}] Executing inside a Kubernetes Container."
                    body()
                    sh "chown -R 1000:1000 ."
                }
            } else {
                if (!config.sidecarImage) {
                    dockerExecuteOnKubernetes(
                        script: script,
                        containerCommand: config.containerCommand,
                        containerShell: config.containerShell,
                        dockerImage: config.dockerImage,
                        dockerPullImage: config.dockerPullImage,
                        dockerEnvVars: config.dockerEnvVars,
                        dockerWorkspace: config.dockerWorkspace,
                        stashContent: config.stashContent
                    ){
                        echo "[INFO][${STEP_NAME}] Executing inside a Kubernetes Pod"
                        body()
                    }
                } else {
                    if(!config.dockerName){
                        config.dockerName = UUID.randomUUID().toString()
                    }

                    Map paramMap = [
                        script: script,
                        containerCommands: [:],
                        containerEnvVars: [:],
                        containerPullImageFlags: [:],
                        containerMap: [:],
                        containerName: config.dockerName,
                        containerPortMappings: [:],
                        containerWorkspaces: [:],
                        stashContent: config.stashContent
                    ]

                    paramMap.containerCommands[config.sidecarImage] = ''

                    paramMap.containerEnvVars[config.dockerImage] = config.dockerEnvVars
                    paramMap.containerEnvVars[config.sidecarImage] = config.sidecarEnvVars

                    paramMap.containerPullImageFlags[config.dockerImage] = config.dockerPullImage
                    paramMap.containerPullImageFlags[config.sidecarImage] = config.sidecarPullImage

                    paramMap.containerMap[config.dockerImage] = config.dockerName
                    paramMap.containerMap[config.sidecarImage] = config.sidecarName

                    paramMap.containerPortMappings = config.containerPortMappings

                    paramMap.containerWorkspaces[config.dockerImage] = config.dockerWorkspace
                    paramMap.containerWorkspaces[config.sidecarImage] = ''

                    dockerExecuteOnKubernetes(paramMap){
                        echo "[INFO][${STEP_NAME}] Executing inside a Kubernetes Pod with sidecar container"
                        if(config.sidecarReadyCommand) {
                            waitForSidecarReadyOnKubernetes(config.sidecarName, config.sidecarReadyCommand)
                        }
                        body()
                    }
                }
            }
        } else {
            boolean executeInsideDocker = true
            if (!JenkinsUtils.isPluginActive(PLUGIN_ID_DOCKER_WORKFLOW)) {
                echo "[WARNING][${STEP_NAME}] Docker not supported. Plugin '${PLUGIN_ID_DOCKER_WORKFLOW}' is not installed or not active. Configured docker image '${config.dockerImage}' will not be used."
                executeInsideDocker = false
            }

            returnCode = sh script: 'docker ps -q > /dev/null', returnStatus: true
            if (returnCode != 0) {
                echo "[WARNING][$STEP_NAME] Cannot connect to docker daemon (command 'docker ps' did not return with '0'). Configured docker image '${config.dockerImage}' will not be used."
                executeInsideDocker = false
            }
            if (executeInsideDocker && config.dockerImage) {
                utils.unstashAll(config.stashContent)
                def image = docker.image(config.dockerImage)
                if (config.dockerPullImage) image.pull()
                else echo"[INFO][$STEP_NAME] Skipped pull of image '${config.dockerImage}'."
                if (!config.sidecarImage) {
                    image.inside(getDockerOptions(config.dockerEnvVars, config.dockerVolumeBind, config.dockerOptions)) {
                        body()
                    }
                } else {
                    def networkName = "sidecar-${UUID.randomUUID()}"
                    sh "docker network create ${networkName}"
                    try{
                        def sidecarImage = docker.image(config.sidecarImage)
                        if (config.sidecarPullImage) sidecarImage.pull()
                        else echo"[INFO][$STEP_NAME] Skipped pull of image '${config.sidecarImage}'."
                        config.sidecarOptions = config.sidecarOptions?:[]
                        if (config.sidecarName)
                            config.sidecarOptions.add("--network-alias ${config.sidecarName}")
                        config.sidecarOptions.add("--network ${networkName}")
                        sidecarImage.withRun(getDockerOptions(config.sidecarEnvVars, config.sidecarVolumeBind, config.sidecarOptions)) { container ->
                            config.dockerOptions = config.dockerOptions?:[]
                            if (config.dockerName)
                                config.dockerOptions.add("--network-alias ${config.dockerName}")
                            config.dockerOptions.add("--network ${networkName}")
                            if(config.sidecarReadyCommand) {
                                waitForSidecarReadyOnDocker(container.id, config.sidecarReadyCommand)
                            }
                            image.inside(getDockerOptions(config.dockerEnvVars, config.dockerVolumeBind, config.dockerOptions)) {
                                echo "[INFO][${STEP_NAME}] Running with sidecar container."
                                body()
                            }
                        }
                    }finally{
                        sh "docker network remove ${networkName}"
                    }
                }
            } else {
                echo "[INFO][${STEP_NAME}] Running on local environment."
                body()
            }
        }
    }
}

private waitForSidecarReadyOnDocker(String containerId, String command){
    String dockerCommand = "docker exec ${containerId} ${command}"
    waitForSidecarReady(dockerCommand)
}

private waitForSidecarReadyOnKubernetes(String containerName, String command){
    container(name: containerName){
        waitForSidecarReady(command)
    }
}

private waitForSidecarReady(String command){
    int sleepTimeInSeconds = 10
    int timeoutInSeconds = 5 * 60
    int maxRetries = timeoutInSeconds / sleepTimeInSeconds
    int retries = 0
    while(true){
        echo "Waiting for sidecar container"
        String status = sh script:command, returnStatus:true
        if(status == "0") return
        if(retries > maxRetries){
            error("Timeout while waiting for sidecar container to be ready")
        }

        sleep sleepTimeInSeconds
        retries++
    }
}

/*
 * Returns a string with docker options containing
 * environment variables (if set).
 * Possible to extend with further options.
 * @param dockerEnvVars Map with environment variables
 */
@NonCPS
private getDockerOptions(Map dockerEnvVars, Map dockerVolumeBind, def dockerOptions) {
    def specialEnvironments = [
        'http_proxy',
        'https_proxy',
        'no_proxy',
        'HTTP_PROXY',
        'HTTPS_PROXY',
        'NO_PROXY'
    ]
    def options = []
    if (dockerEnvVars) {
        for (String k : dockerEnvVars.keySet()) {
            options.add("--env ${k}=${dockerEnvVars[k].toString()}")
        }
    }

    for (String envVar : specialEnvironments) {
        if (dockerEnvVars == null || !dockerEnvVars.containsKey(envVar)) {
            options.add("--env ${envVar}")
        }
    }

    if (dockerVolumeBind) {
        for (String k : dockerVolumeBind.keySet()) {
            options.add("--volume ${k}:${dockerVolumeBind[k].toString()}")
        }
    }

    if (dockerOptions) {
        if (dockerOptions instanceof CharSequence) {
            dockerOptions = [dockerOptions]
        }
        if (dockerOptions instanceof List) {
            for (String option : dockerOptions) {
                options << escapeBlanks(option)
            }
        } else {
            throw new IllegalArgumentException("Unexpected type for dockerOptions. Expected was either a list or a string. Actual type was: '${dockerOptions.getClass()}'")
        }
    }
    return options.join(' ')
}


boolean isContainerDefined(config) {
    Map containerMap = ContainerMap.instance.getMap()

    if (!containerMap.containsKey(env.POD_NAME)) {
        return false
    }

    return containerMap.get(env.POD_NAME).containsKey(config.dockerImage)
}


def getContainerDefined(config) {
    return ContainerMap.instance.getMap().get(env.POD_NAME).get(config.dockerImage).toLowerCase()
}


boolean isKubernetes() {
    return Boolean.valueOf(env.ON_K8S)
}

/*
 * Escapes blanks for values in key/value pairs
 * E.g. <code>description=Lorem ipsum</code> is
 * changed to <code>description=Lorem\ ipsum</code>.
 */
@NonCPS
def escapeBlanks(def s) {

    def EQ='='
    def parts=s.split(EQ)

    if(parts.length == 2) {
        parts[1]=parts[1].replaceAll(' ', '\\\\ ')
        s = parts.join(EQ)
    }

    return s
}
